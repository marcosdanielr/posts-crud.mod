package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
)

type post struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/posts", listPostsHandler)
	mux.HandleFunc("/posts/{id}", getPostHandler)
	mux.HandleFunc("POST /posts", createPostHandler)
	http.ListenAndServe(":8080", mux)
}

func createPostHandler(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open("sqlite3", "posts.db")

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	defer db.Close()

	var p post
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, "Error encoding JSON", http.StatusInternalServerError)
		return
	}

	fmt.Println(r.Body)

	if _, err := db.Exec("INSERT INTO posts (title) VALUES (?)", p.Title); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func getPostHandler(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open("sqlite3", "posts.db")

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer db.Close()

	postId := r.PathValue("id")

	if postId == "" {
		http.Error(w, "Post ID is required", http.StatusBadRequest)
		return
	}

	row := db.QueryRow("SELECT * FROM posts WHERE id=?", postId)

	var p post

	if err := row.Scan(&p.ID, &p.Title); err == sql.ErrNoRows {
		http.NotFound(w, r)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(p)
}

func listPostsHandler(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open("sqlite3", "posts.db")

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer db.Close()

	rows, err := db.Query("SELECT * FROM posts")

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer rows.Close()

	posts := []*post{}

	for rows.Next() {
		var p post

		if err := rows.Scan(&p.ID, &p.Title); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		posts = append(posts, &p)
	}

	if err := json.NewEncoder(w).Encode(posts); err != nil {
		http.Error(w, "Error encoding JSON", http.StatusInternalServerError)
		return
	}
}
