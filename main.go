package main

import (
	"fmt"
	"sync"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type Item struct {
	Id          int    `db:"id"`
	Title       string `db:"title"`
	Description string `db:"description"`
}

// ConnectPostgresDB -> connect postgres db
func ConnectPostgresDB() *sqlx.DB {
	connstring := "user=postgres dbname=postgres sslmode=disable password=postgres host=localhost port=8080"
	db, err := sqlx.Open("postgres", connstring)
	if err != nil {
		fmt.Println(err)
		return db
	}
	return db
}

func InsertItem(item Item, wg *sync.WaitGroup) {
	defer wg.Done()
	db := ConnectPostgresDB()
	defer db.Close()
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
	var wg sync.WaitGroup
	//db, err := sqlx.Connect("postgres", "user=postgres dbname=postgres sslmode=disable password=postgres host=localhost port=8080")
	for i := 1; i <= 2000; i++ {
		item := Item{Id: i, Title: "TestBook", Description: "TestDescription"}
		//go GetItem(db, i, &wg)
		wg.Add(1)
		go InsertItem(item, &wg)

	}
	wg.Wait()
	fmt.Println("All DB Connection is Completed")
}
