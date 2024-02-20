package main

import (
	"database/sql"
	"strconv"
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
	_ "github.com/mattn/go-sqlite3"
)

const (
	ImgDir = "images"
	JSONFile = "items.json"
	DBPath    = "/Users/honokakobayashi/Desktop/mercari-build-training/mercari-build-training-2024/db/mercari.sqlite3"
)

type Item struct {
	ID int `json:"id"`
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


func addItem(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
	// Get form data
	name := c.FormValue("name")
	category := c.FormValue("category")
	image, err := c.FormFile("image")
	// Save the image file
	imageName, err := hashAndSaveImage(image, ImgDir)
	if err != nil {
		return err
	}
	// Check if id or name or category or image is empty
	if name == "" || category == "" {
		return c.JSON(http.StatusBadRequest,
			Response{Message: "Name or category cannot be empty"})
	}
	// define query and execute
	insertQuery := "INSERT INTO items (name, category, image_name) VALUES (?, ?, ?)"
	result, err := db.Exec(insertQuery, name, category, imageName)
	if err != nil {
		return err
	}
	// get ID which is created automatically
	itemID, err := result.LastInsertId()
	if err != nil {
		return err
	}
	message := fmt.Sprintf("Item added with ID: %d", itemID)
	res := Response{Message: message}
	return c.JSON(http.StatusOK, res)
	}
}

func getItems(db *sql.DB) echo.HandlerFunc {
    return func(c echo.Context) error {
        // select all and put it to rows
        rows, err := db.Query("SELECT id, name, category, image_name FROM items")
        if err != nil {
            return err
        }
        defer rows.Close()
        var items []Item
        for rows.Next() {
            var item Item
            err := rows.Scan(&item.ID, &item.Name, &item.Category, &item.ImageName)
            if err != nil {
                return err
            }
            items = append(items, item)
        }
        return c.JSON(http.StatusOK, Items{Items: items})
    }
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

func getItemFromId(db *sql.DB) echo.HandlerFunc {
    return func(c echo.Context) error {
        // Get item ID from URL parameter
        itemIDStr := c.Param("itemID")

        // Convert itemIDStr to integer
        itemID, err := strconv.Atoi(itemIDStr)
        if err != nil {
            return c.JSON(http.StatusBadRequest, Response{Message: "Invalid item ID"})
        }

        // Query database to get item with given ID
        var item Item
        err = db.QueryRow("SELECT id, name, category, image_name FROM items WHERE id = ?", itemID).
    		Scan(&item.ID, &item.Name, &item.Category, &item.ImageName)

        if err == sql.ErrNoRows {
            return c.JSON(http.StatusNotFound, Response{Message: "Item not found"})
        } else if err != nil {
            return err // Return other errors
        }
        // Return item details
        return c.JSON(http.StatusOK, item)
    }
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

	// open the database
	db, err := sql.Open("sqlite3", DBPath)
	if err != nil {
		e.Logger.Infof("Failed to open the database: %v", err)
	}
	defer db.Close()

	e.GET("/", root)
	e.POST("/items", addItem(db))
	e.GET("/items", getItems(db))
	e.GET("/image/:imageFilename", getImg)
	e.GET("/items/:itemID",getItemFromId(db))

	e.Logger.Fatal(e.Start(":9000"))
}
