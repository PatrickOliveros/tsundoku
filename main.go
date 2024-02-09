package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/joho/godotenv"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
	"tsundoku/driver"
	"tsundoku/models"

	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
)

var config models.AppConfig
var appDb *driver.DB

const layout = "2006-01-02"

func runApplication(connectionString string) (*driver.DB, error) {
	log.Println(">>> Connecting to database...")
	db, err := driver.ConnectSQL(connectionString)

	if err != nil {
		log.Fatal(">>> Cannot connect to database! Dying...")
	}

	log.Println(">>> Connected to DB using driver!...")

	return db, nil
}

func main() {

	/* REQUIRES .ENV FILE
	CONNECTION_STRING=connection_string
	*/

	// Load the environment variables from the .env file
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	connectionString := os.Getenv("CONNECTION_STRING")
	fmt.Println(connectionString)

	db, _ := runApplication(connectionString)
	appDb = db

	defer db.SQL.Close()

	fmt.Println("=========================================")
	fmt.Println("Book Populator")
	fmt.Println("=========================================")
	fmt.Println("Please enter ISBN: ")
	fmt.Println("=========================================")

	reader := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print(" -> ")
		reader.Scan()

		text := reader.Text()

		books, title := checkBookIfExists(text)

		if books != 0 {
			fmt.Println(fmt.Sprintf(`Title: '%s' found already.`, title))
			fmt.Println("Skip? (default) or add?")
			fmt.Print(" -> ")
			reader.Scan()
			text2 := reader.Text()

			if strings.ToLower(text2) == "n" {
				fetchData(text, reader)
			} else {
				fmt.Println(fmt.Sprintf(`Skipped: %s - (%s)`, title, text))
			}

		} else {
			fetchData(text, reader)
		}

		if text == "q" {
			fmt.Println("Quit")
			break
		}
	}
}

// TODO: do not trust user input. make sure that the input is totally just the numbers itself
func sanitizedIsbn(src string) string {
	return strings.TrimSpace(src)
}

func checkBookIfExists(isbn string) (int, string) {

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	var availability int
	var title string

	query := `select 
					count(id), title
				from 
					books 
				where 
					(isbn10 = $1) or (isbn13 = $1)
				group by
					id, title`

	row := appDb.SQL.QueryRowContext(ctx, query, isbn)
	err := row.Scan(&availability, &title)

	if err != nil {
		return 0, title
	}

	return availability, title
}

func fetchData(isbn string, reader *bufio.Scanner) (int, error) {

	isbn = sanitizedIsbn(isbn)
	if len(isbn) < 10 {
		return 0, errors.New("invalid input. please ensure input string is at least 10")
	}

	bookRequest := fmt.Sprintf(`https://www.googleapis.com/books/v1/volumes?q=isbn:%s`, isbn)

	res, err := http.Get(bookRequest)
	if err != nil {
		panic(err.Error())
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err.Error())
	}

	parsedData := new(models.Result)
	err = json.Unmarshal(body, &parsedData)
	if err != nil {
		fmt.Println("whoops:", err)
	}

	if parsedData.TotalItems < 1 {
		fmt.Println(fmt.Sprintf(`No records found for: %s`, isbn))
		return 0, nil
	}

	parsedTitle := parsedData.Items[0].VolumeInfo.Title

	fmt.Println(fmt.Sprintf(`Title: '%s' found. Process this?`, parsedTitle))

	fmt.Print(" -> ")
	reader.Scan()
	text := reader.Text()

	if strings.ToLower(text) != "n" {
		saveJson(parsedData)
	} else {
		fmt.Println(fmt.Sprintf(`Skipped: %s`, parsedTitle))
	}

	return 0, nil
}

func getParsedTime(strDate string) time.Time {
	tmptime, err := time.Parse(layout, strDate)

	if err != nil {
		tmptime = time.Now().UTC()
	}

	return tmptime
}

func saveJson(source *models.Result) (int, error) {

	itemData := source.Items[0].VolumeInfo
	parsedDate := getParsedTime(itemData.PublishedDate)
	tmpIsbn13 := ""
	tmpIsbn10 := ""

	for _, itemx := range itemData.IndustryIdentifiers {
		if itemx.Type == "ISBN_13" {
			tmpIsbn13 = itemx.Identifier
		}

		if itemx.Type == "ISBN_10" {
			tmpIsbn10 = itemx.Identifier
		}
	}

	item := models.BookData{
		Title:         itemData.Title,
		Subtitle:      itemData.Subtitle,
		Description:   itemData.Description,
		Publisher:     itemData.Publisher,
		Thumbnail:     itemData.ImageLinks.Thumbnail,
		PublishedDate: parsedDate,
		SelfLink:      source.Items[0].SelfLink,
		Categories:    strings.Join(itemData.Categories, ","),
		Authors:       strings.Join(itemData.Authors, ","),
		ISBN10:        tmpIsbn10,
		ISBN13:        tmpIsbn13,
	}

	var newID int

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	stmt := `insert into books (Title, Subtitle, PublishedDate, Description,
		Publisher, BookThumnail, SelfLink, Categories, Authors, ISBN10, ISBN13, inlibby, notes) values
		($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13) returning id `

	err := appDb.SQL.QueryRowContext(ctx, stmt,
		item.Title, item.Subtitle, item.PublishedDate, item.Description, item.Publisher,
		item.Thumbnail, item.SelfLink, item.Categories, item.Authors,
		item.ISBN10, item.ISBN13, false, "test").Scan(&newID)

	if err != nil {
		fmt.Println(err)
		return 0, err
	} else {
		fmt.Println(fmt.Sprintf("(%d) Created record for: '%s'", newID, item.Title))
	}

	return newID, nil
}
