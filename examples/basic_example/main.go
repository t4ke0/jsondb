package main

import (
	"fmt"
	"log"

	"github.com/t4ke0/jsondb"
)

type Table struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

var testData = []Table{
	{Id: 1, Name: "u1"},
	{Id: 2, Name: "u2"},
	{Id: 3, Name: "u3"},
	{Id: 4, Name: "u4"},
}

func main() {
	const dbFile string = "db.json"

	conn, err := jsondb.Connect[Table](dbFile)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	if err := conn.Init(); err != nil {
		log.Fatal(err)
	}

	for _, t := range testData {
		if err := conn.WriteToDB(t); err != nil {
			log.Fatal(err)
		}
	}

	// Example of deleting an element from the database. delete index 3, which
	// is the last element in testData array.
	if err := conn.DeleteFromDB(3); err != nil {
		log.Fatal(err)
	}

	// Example of deleting an element from the database.
	if err := conn.UpdateDB(0, Table{Id: 69, Name: "NICE"}); err != nil {
		log.Fatal(err)
	}

	content, err := conn.ReadFromDB()
	if err != nil {
		log.Fatal(err)
	}

	for _, n := range content {
		fmt.Println(n)
	}
}
