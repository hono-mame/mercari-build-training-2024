package main

import (
	"encoding/json"
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
	Name string `json:"name"`
	Category string `json:"category"`
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


func addItem(c echo.Context) error {
	// Get form data
	name := c.FormValue("name")
	category := c.FormValue("category")
	// Check if name or category is empty
	if name == "" || category == "" {
		return c.JSON(http.StatusBadRequest, 
			Response{Message: "Name or category cannot be empty"})
	}
	// Create new item to add to JSON file
	new_item := Item{Name: name, Category: category}
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

func main() {
	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Logger.SetLevel(log.INFO)

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
	e.GET("/image/:imageFilename", getImg)

	e.Logger.Fatal(e.Start(":9000"))
}
