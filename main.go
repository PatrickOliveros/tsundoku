package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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
	"github.com/joho/godotenv"
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
			strInput := reader.Text()

			if strings.ToLower(strInput) == "n" {
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

	body, err := io.ReadAll(res.Body)
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
		fmt.Println(fmt.Sprintf(`Trying open library API for ISBN: %s`, isbn))

		return fetchOpenLibrary(isbn, reader)
	} else {
		parsedTitle := parsedData.Items[0].VolumeInfo.Title

		fmt.Println(fmt.Sprintf(`Title: '%s' found. Process this?`, parsedTitle))

		fmt.Print(" -> ")
		reader.Scan()
		text := reader.Text()

		if strings.ToLower(text) != "n" {
			buildFromGoogle(parsedData)
		} else {
			fmt.Println(fmt.Sprintf(`Skipped: %s`, parsedTitle))
		}
	}

	return 0, nil
}

func fetchOpenLibrary(isbn string, reader *bufio.Scanner) (int, error) {

	bookRequest := fmt.Sprintf(`https://openlibrary.org/api/books?bibkeys=ISBN:%s&jscmd=details&format=json`, isbn)

	res, err := http.Get(bookRequest)
	if err != nil {
		panic(err.Error())
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		panic(err.Error())
	}

	// Define a map to hold the dynamic keys and their values
	var bookData map[string]json.RawMessage

	// Unmarshal JSON data into the map
	if err := json.Unmarshal(body, &bookData); err != nil {
		fmt.Println("Error:", err)
		return 0, err
	}

	var bookInfo map[string]interface{}

	// Access the dynamic keys and their values
	for _, value := range bookData {
		if err := json.Unmarshal(value, &bookInfo); err != nil {
			fmt.Println("Error unmarshaling book info:", err)
			continue
		}

		// We'll get only the first object
		break
	}

	bookDetails, ok := bookInfo["details"].(map[string]interface{})

	if !ok {
		fmt.Println(`Book not found from Open Library API`)

	} else {
		return buildFromOpenLibrary(bookInfo, bookDetails, reader)
	}

	return defaultError()
}

func GetKeyValue(data map[string]interface{}, key string) (string, error) {
	// Queue to perform iterative BFS
	queue := []interface{}{data}

	current := queue[0]
	queue = queue[1:]

	// Check if the current element is a map
	if node, ok := current.(map[string]interface{}); ok {
		if keyValue, ok := node[key]; ok {

			if ok {
				if str, ok := keyValue.(string); ok {
					return str, nil
				}

				if m, ok := keyValue.(map[string]interface{}); ok {
					// Check if the map has a "value" property
					if value, ok := m["value"]; ok {
						// Check if the "value" property is a string
						if str, ok := value.(string); ok {
							return str, nil
						}
					}
				}

				if strArray, ok := keyValue.([]interface{}); ok && len(strArray) > 0 {
					var strValues []string
					for _, v := range strArray {
						if str, ok := v.(string); ok {
							strValues = append(strValues, str)
						}
					}

					return strings.Join(strValues, ","), nil
				}

			}
		}
	}

	return "", errors.New("something went wrong")
}

func getParsedTime(strDate string) time.Time {
	tmptime, err := time.Parse(layout, strDate)

	if err != nil {
		inputLayout := "January 2, 2006"
		parsedTime, err := time.Parse(inputLayout, strDate)
		if err != nil {
			return time.Now().UTC()
		}

		tmptime, err = time.Parse(layout, parsedTime.Format(layout))
		if err != nil {
			return time.Now().UTC()
		}
	}

	return tmptime
}

func buildFromOpenLibrary(bookInfo map[string]interface{}, bookDetails map[string]interface{}, reader *bufio.Scanner) (int, error) {

	itemData, ok := GetKeyValue(bookDetails, "title")

	if ok == nil {
		item := models.BookData{
			Title: itemData,
		}

		itemData, ok = GetKeyValue(bookDetails, "subtitle")
		if ok == nil {
			item.Subtitle = itemData
		}

		itemData, ok = GetKeyValue(bookDetails, "description")
		if ok == nil {
			item.Description = itemData
		}

		itemData, ok = GetKeyValue(bookDetails, "by_statement")
		if ok == nil {
			item.Authors = itemData
		}

		itemData, ok = GetKeyValue(bookDetails, "publishers")
		if ok == nil {
			item.Publisher = itemData
		}

		itemData, ok = GetKeyValue(bookDetails, "isbn_13")
		if ok == nil {
			item.ISBN13 = itemData
		}

		itemData, ok = GetKeyValue(bookDetails, "isbn_10")
		if ok == nil {
			item.ISBN10 = itemData
		}

		itemData, ok = GetKeyValue(bookDetails, "publish_date")
		if ok == nil {
			item.PublishedDate = getParsedTime(itemData)
		}

		item.Thumbnail = bookInfo["thumbnail_url"].(string)
		item.SelfLink = bookInfo["info_url"].(string)
		item.Source = "openlibrary"

		parsedTitle := item.Title

		fmt.Println(fmt.Sprintf(`Title: '%s' found. Process this?`, parsedTitle))

		fmt.Print(" -> ")
		reader.Scan()
		text := reader.Text()

		if strings.ToLower(text) != "n" {
			return savetoDB(item)
		} else {
			fmt.Println(fmt.Sprintf(`Skipped: %s`, parsedTitle))
			return 0, nil
		}
	}

	return defaultError()
}

func buildFromGoogle(source *models.Result) (int, error) {

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
		Source:        "google",
	}

	return savetoDB(item)
}

func savetoDB(source models.BookData) (int, error) {
	var newID int

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	stmt := `insert into books (Title, Subtitle, PublishedDate, Description,
		Publisher, BookThumnail, SelfLink, Categories, Authors, ISBN10, ISBN13, inlibby, notes, datasource) values
		($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14) returning id `

	err := appDb.SQL.QueryRowContext(ctx, stmt,
		source.Title, source.Subtitle, source.PublishedDate, source.Description, source.Publisher,
		source.Thumbnail, source.SelfLink, source.Categories, source.Authors,
		source.ISBN10, source.ISBN13, false, "test", source.Source).Scan(&newID)

	if err != nil {
		fmt.Println(err)
		return 0, err
	}

	fmt.Println(fmt.Sprintf("(%d) Created record for: '%s'", newID, source.Title))

	return newID, nil
}

func defaultError() (int, error) {
	return 0, errors.New("something wrong happened")
}
