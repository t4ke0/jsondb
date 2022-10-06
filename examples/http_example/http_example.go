package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/t4ke0/jsondb"
)

type User struct {
	Username string `json:"username"`
	Email    string `json:"email"`
}

func main() {

	conn, err := jsondb.Connect[User]("users.db")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	if err := conn.Init(); err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/list", func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if r := recover(); r != nil {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, "Error: %v", r)
				return
			}
		}()
		users, err := conn.ReadFromDB()
		if err != nil {
			panic(err)
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(users); err != nil {
			panic(err)
		}
	})

	http.HandleFunc("/new/user", func(w http.ResponseWriter, r *http.Request) {
		if r.ContentLength == 0 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		defer func() {
			if r := recover(); r != nil {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, "Error: %v", r)
				return
			}
		}()

		defer func() {
			if err := r.Body.Close(); err != nil {
				panic(err)
			}
		}()

		data, err := io.ReadAll(r.Body)
		if err != nil {
			panic(err)
		}

		var req User

		if err := json.Unmarshal(data, &req); err != nil {
			panic(err)
		}

		users, err := conn.ReadFromDB()
		if err != nil {
			panic(err)
		}

		var found bool
		for _, u := range users {
			if u.Username == req.Username {
				found = true
				break
			}
		}
		if found {
			w.WriteHeader(http.StatusConflict)
			return
		}

		if err := conn.WriteToDB(req); err != nil {
			panic(err)
		}
		w.WriteHeader(http.StatusCreated)
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}
