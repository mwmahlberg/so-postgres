package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

const schema = `
CREATE TABLE IF NOT EXISTS items (
		id integer primary key,
		title text,
		description text
);
`

type Item struct {
	Id          int    `db:"id"`
	Title       string `db:"title"`
	Description string `db:"description"`
}

var (
	// Make the waitgroup global: Easier to use and less error-prone
	wg sync.WaitGroup

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
	defer wg.Done()
	// With the beginning of the transaction, a connection is acquired from the pool
	tx, err := db.Beginx()
	if err != nil {
		panic(fmt.Errorf("beginning transaction: %s", err))
	}

	_, err = tx.Exec("INSERT INTO items(id, title, description) VALUES($1, $2, $3)", item.Id, item.Title, item.Description)
	if err != nil {
		// the rollback is rather superfluous here
		// but it's good practice to include it
		tx.Rollback()

		// panic will cause the goroutine to exit and the waitgroup to decrement
		// Also, the handlePanics function will catch the panic and log the error
		panic(fmt.Errorf("inserting data: %s", err))
	}

	err = tx.Commit()
	if err != nil {
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

	// use the actual value
	maxOpen := db.DB.Stats().MaxOpenConnections

	var mutex sync.Mutex
	for i := 1; i <= 2000; i++ {

		wg.Add(1)

		// For goroutines, you must explicitly set the panic handler
		go func(i int) {

			defer handlePanics()

			// use a label to ensure that the goroutine breaks out of inner loop
		waitForOpenConnection:
			for {
				// Lock the mutex to check the number of open connections.
				// We need to do this otherwise another goroutine could increment the number of open connections
				mutex.Lock()

				// Get the connections in the pool that are currently in use
				switch open := db.DB.Stats().InUse; {

				// If the number of open connections is less than the maximum, insert the item
				case open <= maxOpen:
					InsertItem(Item{Id: i, Title: "TestBook", Description: "TestDescription"}, db)
					// Now that the item has been inserted, unlock the mutex and break out of the inner loop
					mutex.Unlock()
					break waitForOpenConnection
				default:
					// Allow other goroutines to read the number of open connections
					mutex.Unlock()
				}
			}
		}(i)
	}
	wg.Wait()
	log.Printf("All DB Inserts completed after %s\n", time.Since(start))
}
