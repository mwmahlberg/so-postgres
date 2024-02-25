package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

const schema = `
CREATE TABLE IF NOT EXISTS items (
		id serial primary key,
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
	wg    sync.WaitGroup
	dburl string
)

func init() {
	flag.StringVar(&dburl, "dburl", "user=postgres dbname=postgres sslmode=disable password=postgres host=localhost port=5432", "Postgres DB URL")
}

func handlePanics() {
	if r := recover(); r != nil {
		log.Println("encountered panic: ", r)
		os.Exit(1)
	}
}

func InsertItem(item Item, db *sqlx.DB) {
	defer wg.Done()
	tx, err := db.Beginx()
	if err != nil {
		panic(fmt.Errorf("beginning transaction: %s", err))
	}

	_, err = tx.Queryx("INSERT INTO items(id, title, description) VALUES($1, $2, $3)", item.Id, item.Title, item.Description)
	if err != nil {
		tx.Rollback()
		panic(fmt.Errorf("inserting data: %s", err))
	}

	err = tx.Commit()
	if err != nil {
		panic(fmt.Errorf("committing transaction: %s", err))
	}

	fmt.Println("Data is Successfully inserted!!")
}

func main() {

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
	db.MustExec(schema)
	for i := 1; i <= 2000; i++ {
		item := Item{Id: i, Title: "TestBook", Description: "TestDescription"}
		wg.Add(1)
		go func() {
			defer handlePanics()
			InsertItem(item, db)
		}()
	}
	wg.Wait()
	fmt.Println("All DB Connection is Completed")
}
