package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

const (
	schema = `
CREATE TABLE IF NOT EXISTS items (
		id integer primary key,
		title text,
		description text
);
`
	insert = `
INSERT INTO items(id, title, description) VALUES($1, $2, $3)
`
)

type Item struct {
	Id          int    `db:"id"`
	Title       string `db:"title"`
	Description string `db:"description"`
}

var (
	// Make the database URL a configurable flag
	dburl string
)

func init() {
	// Make the database URL a configurable flag
	flag.StringVar(&dburl, "dburl", "user=postgres dbname=postgres sslmode=disable password=postgres host=localhost port=5432", "Postgres DB URL")
}

// handlePanics is a simple function to log the error that caused a panic and exit the program
func handlePanics() {
	if r := recover(); r != nil {
		log.Println("encountered panic: ", r)
		os.Exit(1)
	}
}

// InsertItem inserts an item into the database.
// Note that the db is passed as an argument.
func InsertItem(item Item, db *sqlx.DB) {

	var (
		tx  *sqlx.Tx
		err error
	)

	// With the beginning of the transaction, a connection is acquired from the pool
	if tx, err = db.Beginx(); err != nil {
		panic(fmt.Errorf("beginning transaction: %s", err))
	}

	if _, err = tx.Exec(insert, item.Id, item.Title, item.Description); err != nil {
		// the rollback is rather superfluous here
		// but it's good practice to include it
		tx.Rollback()

		// panic will cause the goroutine to exit and the waitgroup to decrement
		// Also, the handlePanics function will catch the panic and log the error
		panic(fmt.Errorf("inserting data: %s", err))
	}

	if err = tx.Commit(); err != nil {
		panic(fmt.Errorf("committing transaction: %s", err))
	}

}

func main() {

	// Recover from panics and log the error for the main goroutine
	defer handlePanics()

	flag.Parse()
	log.Printf("DB URL: %s\n", dburl)

	var (
		db  *sqlx.DB
		err error
	)

	// Only open one connection to the database.
	// The postgres driver will open a pool of connections for you.
	if db, err = sqlx.Connect("postgres", dburl); err != nil {
		log.Fatalln(err)
	}
	defer db.Close()

	// Create the items table
	// Note that if this panics, the handlePanics function will catch it and log the error
	db.MustExec(schema)
	start := time.Now()

	// Set the number of connections in the pool
	db.DB.SetMaxOpenConns(10)

	for i := 1; i <= 2000; i++ {
		// use a label to ensure that the goroutine breaks out of inner loop
		InsertItem(Item{Id: i, Title: "TestBook", Description: "TestDescription"}, db)
	}
	log.Printf("All DB Inserts completed after %s\n", time.Since(start))
}
