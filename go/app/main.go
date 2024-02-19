package main

import (
	"mime/multipart"
	"crypto/sha256"
	"encoding/hex"
	"path/filepath"
	"encoding/json"
	"io"
	"io/ioutil"
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
)

const (
	ImgDir = "images"
	JSONFile = "items.json"
)

type Item struct {
	ID string `json:"id"`
	Name string `json:"name"`
	Category string `json:"category"`
	ImageName string `json:"image_name"`
}

type Items struct {
	Items []Item `json:"items"`
}

type Response struct {
	Message string `json:"message"`
}

func root(c echo.Context) error {
	res := Response{Message: "Hello, world!"}
	return c.JSON(http.StatusOK, res)
}

func readItemsFromFile() (*Items, error) {
	file, err := ioutil.ReadFile(JSONFile)
	if err != nil {
		return nil, err
	}
	var items Items
	// storing file contents in items
	// JSON -> Items
	if err := json.Unmarshal(file, &items); err != nil {
		return nil, err
	}
	return &items, nil
}

func writeItemsToFile(items *Items) error {
    file, err := os.Create(JSONFile)
    if err != nil {
        return err
    }
    defer file.Close()
	// Items -> JSON
    encoder := json.NewEncoder(file)
    if err := encoder.Encode(items); err != nil {
        return err
    }
    return nil
}

func hashAndSaveImage(image *multipart.FileHeader, imgDir string) (string, error){
	// open the image file
	src, err := image.Open()
	if(err != nil){
		return "", err
	}
	defer src.Close()
	// create a hash of the image file
	hash := sha256.New()
	if _, err := io.Copy(hash, src); err != nil {
		return "", err
	}
	imageHash := hex.EncodeToString(hash.Sum(nil))
	// Reset the file position to the beginning
    if _, err := src.Seek(0, io.SeekStart); err != nil {
        return "", err
    }
	// create the image file in the directory
	imageName := filepath.Join(ImgDir, imageHash+".jpg")
	dst, err := os.Create(imageName)
	if err != nil {
		return "", err
	}
	defer dst.Close()
	// Copy the image contents to the destination file
    if _, err := io.Copy(dst, src); err != nil {
        return "", err
    }
    return imageName, nil
}


func addItem(c echo.Context) error {
	// Get form data
	name := c.FormValue("name")
	category := c.FormValue("category")
	image, err := c.FormFile("image")
	// Save the image file
    imageName, err := hashAndSaveImage(image, ImgDir)
    if err != nil {
        return err
    }

	// Check if name or category is empty
	if name == "" || category == "" {
		return c.JSON(http.StatusBadRequest, 
			Response{Message: "Name or category cannot be empty"})
	}
	// Create new item to add to JSON file
	new_item := Item{Name: name, Category: category, ImageName: imageName}
	// for debug: Received item
	fmt.Printf("Received item: %+v\n", new_item)
	// Read existing items from JSON file
	items, err := readItemsFromFile()
	if err != nil {
		return err
	}
	// Append new item to items
	items.Items = append(items.Items, new_item)
	// Write items back to JSON file
	if err := writeItemsToFile(items); err != nil {
		return err
	}
	message := fmt.Sprintf("Item received: %s", new_item.Name)
	res := Response{Message: message}
	return c.JSON(http.StatusOK, res)
}

func getItems(c echo.Context) error {
	items, err := readItemsFromFile()
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, items)
}

func getImg(c echo.Context) error {
	imgPath := path.Join(ImgDir, c.Param("imageFilename"))

	if !strings.HasSuffix(imgPath, ".jpg") {
		res := Response{Message: "Image path does not end with .jpg"}
		return c.JSON(http.StatusBadRequest, res)
	}
	if _, err := os.Stat(imgPath); err != nil {
		c.Logger().Debugf("Image not found: %s", imgPath)
		imgPath = path.Join(ImgDir, "default.jpg")
	}
	return c.File(imgPath)
}

func getItemFromId(c echo.Context) error {
	// get item ID from URL parameter
    itemID := c.Param("itemID")
	// read items from JSON file
    items, err := readItemsFromFile()
    if err != nil {
        return err
    }
    // Find the item with the given ID
    var foundItem *Item
    for _, item := range items.Items {
        if item.ID == itemID {
            foundItem = &item
            break
        }
    }
    // if item is not found, return 404 Not Found
    if foundItem == nil {
        return c.JSON(http.StatusNotFound, Response{Message: "Item not found"})
    }
    // otherwise, return the item's details
    return c.JSON(http.StatusOK, foundItem)
}

func main() {
	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Logger.SetLevel(log.DEBUG)

	frontURL := os.Getenv("FRONT_URL")
	if frontURL == "" {
		frontURL = "http://localhost:3000"
	}
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{frontURL},
		AllowMethods: []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete},
	}))

	e.GET("/", root)
	e.POST("/items", addItem)
	e.GET("/items", getItems)
	e.GET("/image/:imageFilename", getImg)
	e.GET("/items/:itemID",getItemFromId)

	e.Logger.Fatal(e.Start(":9000"))
}
