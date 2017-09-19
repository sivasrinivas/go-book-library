package main

import (
	"encoding/json"
	"encoding/xml"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
)

type Book struct {
	PK             int    `json:"pk" gorm:"primary_key"`
	Title          string `json:"title"`
	Author         string `json:"author"`
	Classification string `json:"classification"`
}

type ClassifySearchResponse struct {
	Results []SearchResult `xml:"works>work"`
}

type ClassifyBookResponse struct {
	BookData struct {
		Title  string `xml:"title,attr"`
		Author string `xml:"author,attr"`
		Id     string `xml:"owi,attr"`
	} `xml:"work"`
	Classification struct {
		MostPopular string `xml:"sfa,attr"`
	} `xml:"recommendations>ddc>mostPopular"`
}

type SearchResult struct {
	Title  string `xml:"title,attr"`
	Author string `xml:"author,attr"`
	Year   string `xml:"hyr,attr"`
	Id     string `xml:"owi,attr"`
}

func initDb() {
	var err error
	dbName := "dev.db"
	db, err = gorm.Open("sqlite3", "dev.db")
	if err != nil {
		panic("Failed to connect to database " + dbName)
	}
}

var db *gorm.DB

func main() {
	initDb()

	gin.SetMode(gin.DebugMode)
	router := gin.Default()

	router.Static("/static", "./static")
	router.StaticFile("/", "./static/index.html")

	router.GET("/search", func(context *gin.Context) {
		searchStr := context.Query("search-string")
		results, err := search(searchStr)
		if err != nil {
			context.Error(err)
		} else {
			encoder := json.NewEncoder(context.Writer)
			if err := encoder.Encode(results); err != nil {
				context.Error(err)
			}
		}
	})

	router.GET("/books", func(context *gin.Context) {
		var books []Book
		if db.Find(&books).Error != nil {
			context.Error(db.Error)
		}

		context.JSON(http.StatusOK, books)
	})

	router.POST("/books/:id", func(context *gin.Context) {
		bookId := context.Param("id")
		var book ClassifyBookResponse
		var err error

		if book, err = find(bookId); err != nil {
			context.Error(err)
		}
		newBook := Book{Title: book.BookData.Title, Author: book.BookData.Author, Classification: book.Classification.MostPopular}
		if db.Create(&newBook).Error != nil {
			context.Error(db.Error)
		}
		context.JSON(http.StatusOK, newBook)
	})

	router.DELETE("/books/:pk", func(context *gin.Context) {
		pk := context.Param("pk")
		var book Book
		db.First(&book, "pk=?", pk)
		if db.Delete(&book).Error != nil {
			context.Status(http.StatusNotFound)
		}
		context.Status(http.StatusOK)
	})

	port := os.Getenv("PORT")
	router.Run(":" + port)
}

func find(id string) (ClassifyBookResponse, error) {
	var c ClassifyBookResponse

	body, err := classifyAPI("http://classify.oclc.org/classify2/Classify?&summary=true&owi=" + url.QueryEscape(id))
	if err != nil {
		return ClassifyBookResponse{}, err
	}
	err = xml.Unmarshal(body, &c)
	return c, err
}

func search(query string) ([]SearchResult, error) {
	var c ClassifySearchResponse

	body, err := classifyAPI("http://classify.oclc.org/classify2/Classify?&summary=true&title=" + url.QueryEscape(query))
	if err != nil {
		return []SearchResult{}, err
	}
	err = xml.Unmarshal(body, &c)
	return c.Results, err
}

func classifyAPI(url string) ([]byte, error) {
	var resp *http.Response
	var err error

	if resp, err = http.Get(url); err != nil {
		return []byte{}, err
	}
	defer resp.Body.Close()

	return ioutil.ReadAll(resp.Body)
}
