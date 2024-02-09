# tsundoku

>## the practice of buying a lot of books and keeping them in a pile because you intend to read them but have not done so yet

I have bought a lot of books from thrift stores but I find it hard to remember if I have them already or not. I wrote this application inspired about working with Books in my stint in Wolters Kluwer where I can get book information either with the use of ISBN-10 or 13. Also, instead of doing this in C#, my primary language, I decided to use Go to get my hands dirty.


## Requirement
- Go
- PostgreSQL

## Database
The application expects to have a database with a table named "books" and has the following schema:

```
String Columns
        Title 
        SubTitle
        Description
        Publisher
        BookThumbnail
        SelfLink
        Categories
        Authors       
        ISBN10 
        ISBN13 
        Notes 

DateTime Column
        PublishedDate
```

## How to Run

- This file requires to have a .env file that contains a key "CONNECTION_STRING" and a corresponding entry. For this example I am using a PostgreSQL basic connection string

```
CONNECTION_STRING=host=123.456.789.123 port=1234 user=dbUser password=dbPassword dbname=dbName sslmode=disable
```

- Once the application has successfully connected to the database, you can enter the ISBN of a book and the application will get the corresponding data from the internet. For now, access to the Google API is public so this may not work in the future. 
- If book data exists, it will show you a summary of the book. It will ask you if you want to save the data or not. 
- Press "q" to quit the application.