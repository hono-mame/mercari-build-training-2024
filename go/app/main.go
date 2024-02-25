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
	DBPath    = "../../db/mercari.sqlite3"
	DBSchemaPath = "../../db/items.db"
)

type Item struct {
	Name string `json:"name"`
	Category_id int `json:"category"`
	ImageName string `json:"image_name"`
}

type Items struct {
	Items []Item `json:"items"`
}

type Response struct {
	Message string `json:"message"`
}

type ItemWithCategory struct {
	Id          int    `json:"id"`
    Name        string `json:"name"`
    Category    string `json:"category"`
    Image_name   string `json:"image_name"`
}

type ItemsResponse struct {
    Items []ItemWithCategory `json:"items"`
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

func getCategoryID(db *sql.DB, categoryName string) (int, error) {
	var categoryID int
	err := db.QueryRow("SELECT id FROM categories WHERE name = ?", categoryName).Scan(&categoryID)
	if err != nil {
		if err == sql.ErrNoRows {
			result, err := db.Exec("INSERT INTO categories (name) VALUES (?)", categoryName)
			if err != nil {
				return 0, err
			}
			categoryID, err := result.LastInsertId()
			if err != nil {
				return 0, err
			}
			message := fmt.Sprintf("New category '%s' added with ID: %d", categoryName, categoryID)
			log.Printf(message)
			return int(categoryID), nil
		}
		return 0, err
	}
	return categoryID, nil
}

func addItem(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
	// Get form data
	name := c.FormValue("name")
	categoryName := c.FormValue("category")
	image, err := c.FormFile("image")
	if err != nil {
		c.Logger().Debugf("Failed to load image")
		return err
	}
	// Check if id or name or category or image is empty
	if name == "" || categoryName == "" {
		return c.JSON(http.StatusBadRequest,
			Response{Message: "Name or category cannot be empty"})
	}
	// Get category ID from category name
	categoryID, err := getCategoryID(db, categoryName)
	if err != nil {
		if httpErr, ok := err.(*echo.HTTPError); ok {
			// Handle HTTP errors
			return c.JSON(httpErr.Code, Response{Message: httpErr.Message.(string)})
		}
		c.Logger().Debugf("Category error")
		return err
	}
	// Save the image file
	imageName, err := hashAndSaveImage(image, ImgDir)
	if err != nil {
		c.Logger().Debugf("Image processing failed")
		return err
	}
	// Remove "images/" prefix from imageName
	imageName = strings.TrimPrefix(imageName, "images/")
	// define query and execute
	insertQuery := "INSERT INTO items (name, category_id, image_name) VALUES (?, ?, ?)"
	result, err := db.Exec(insertQuery, name, categoryID, imageName)
	if err != nil {
		c.Logger().Debugf("Failed to inser the data to database")
		return err
	}
	// get ID which is created automatically
	itemID, err := result.LastInsertId()
	if err != nil {
		c.Logger().Debugf("Failed to get ID")
		return err
	}
	message := fmt.Sprintf("Item added with ID: %d", itemID)
	res := Response{Message: message}
	return c.JSON(http.StatusOK, res)
	}
}

func getItems(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		query := `
            SELECT items.id, items.name, categories.name, items.image_name
            FROM items
            JOIN categories ON items.category_id = categories.id
        `
		rows, err := db.Query(query)
		if err != nil {
			return err
		}
		defer rows.Close()
        var items []ItemWithCategory
        for rows.Next() {
            var item ItemWithCategory
            if err := rows.Scan(&item.Id, &item.Name, &item.Category, &item.Image_name); err != nil {
                return err
            }
            items = append(items, item)
        }
        return c.JSON(http.StatusOK, ItemsResponse{Items: items})
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
        var item ItemWithCategory
		query := `
            SELECT items.name, categories.name, items.image_name
            FROM items
            JOIN categories ON items.category_id = categories.id
			WHERE items.id = ?
        `
        err = db.QueryRow(query, itemID).
    		Scan(&item.Name, &item.Category, &item.Image_name)

        if err == sql.ErrNoRows {
            return c.JSON(http.StatusNotFound, Response{Message: "Item not found"})
        } else if err != nil {
            return err // Return other errors
        }
        // Return item details
        return c.JSON(http.StatusOK, item)
    }
}

func searchItemsByKeyword(db *sql.DB) echo.HandlerFunc {
    return func(c echo.Context) error {
		// get keyword from parameter
        keyword := c.QueryParam("keyword")
		// Construct the query to search for items containing the keyword in their name
        query := `
            SELECT items.name, categories.name, items.image_name
            FROM items
            JOIN categories ON items.category_id = categories.id
            WHERE items.name LIKE ?
        `
		rows, err := db.Query(query, "%"+keyword+"%")
        if err != nil {
            return err
        }
        defer rows.Close()
		// for result
        var items []ItemWithCategory
        for rows.Next() {
            var item ItemWithCategory
            if err := rows.Scan(&item.Name, &item.Category, &item.Image_name); err != nil {
                return err
            }
            items = append(items, item)
        }
        return c.JSON(http.StatusOK, ItemsResponse{Items: items})
    }
}

func setupDatabase(DBPath string) (*sql.DB, error) {
	// Open the database
	db, err := sql.Open("sqlite3", DBPath)
	if err != nil {
		return nil, err
	}
	// Create table if not exists
	result, err := os.ReadFile(DBSchemaPath)
	if err != nil {
		return nil, err
	}
	if _, err := db.Exec(string(result)); err != nil {
		return nil, fmt.Errorf("failed to create table: %v", err)
	}
	return db, nil
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

	// Setup the database
	db, err := setupDatabase(DBPath)
	if err != nil {
		fmt.Printf("Failed to setup the database: %v\n", err)
		return
	}
	defer db.Close()

	e.GET("/", root)
	e.POST("/items", addItem(db))
	e.GET("/items", getItems(db))
	e.GET("/image/:imageFilename", getImg)
	e.GET("/items/:itemID",getItemFromId(db))
	e.GET("/search", searchItemsByKeyword(db))
	e.Logger.Fatal(e.Start(":9000"))
}
