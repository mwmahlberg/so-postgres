package main

import (
	"flag"
	"fmt"
	"log"
	"sync"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

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

func InsertItem(item Item, db *sqlx.DB) {
	defer wg.Done()
	tx, err := db.Beginx()
	if err != nil {
		fmt.Println(err)
		return
	}

	_, err = tx.Queryx("INSERT INTO items(id, title, description) VALUES($1, $2, $3)", item.Id, item.Title, item.Description)
	if err != nil {
		fmt.Println(err)
	}

	err = tx.Commit()
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Data is Successfully inserted!!")
}

func main() {

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

	for i := 1; i <= 2000; i++ {
		item := Item{Id: i, Title: "TestBook", Description: "TestDescription"}
		wg.Add(1)
		go InsertItem(item, db)

	}
	wg.Wait()
	fmt.Println("All DB Connection is Completed")
}
