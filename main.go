package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

type echoResponse struct {
	Payload   interface{} `json:"payload"`
	Timestamp int64       `json:"timestamp"`
}

func main() {
	var err error
	db, err = sql.Open("sqlite3", "./database.sqlite")

	if err != nil {
		log.Fatalf("can't open database: %v", err)
	}

	fmt.Printf("Setup database\n")
	ddlStatement := "CREATE TABLE IF NOT EXISTS payload (id INTEGER PRIMARY KEY AUTOINCREMENT, payload VARCHAR);"
	if _, err := runStatement(ddlStatement); err != nil {
		log.Fatalf("can't setup database: %v", err)
	}

	fmt.Printf("Starting server at port 8080\n")
	http.HandleFunc("/echoservice", handler)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "invalid request method, only POST is supported", http.StatusBadRequest)
		return
	}

	var payload interface{}
	t := time.Now().Unix()

	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid request data", http.StatusBadRequest)
		return
	}

	if err := savePayload(payload); err != nil {
		http.Error(w, fmt.Sprintf("%v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(echoResponse{payload, t}); err != nil {
		http.Error(w, fmt.Sprintf("%v", err), http.StatusInternalServerError)
	}
}

func savePayload(payload interface{}) error {
	p, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("can't save the payload: %v", err)
	}
	if _, err := runStatement(fmt.Sprintf("INSERT INTO payload (payload) VALUES ('%v');", string(p))); err != nil {
		return fmt.Errorf("can't save the payload: %v", err)
	}
	return nil
}

func runStatement(statement string) (*sql.Result, error) {
	stmt, err := db.Prepare(statement)
	if err != nil {
		return nil, fmt.Errorf("can't prepare statement: %v", err)
	}
	result, err := stmt.Exec()
	if err != nil {
		return nil, fmt.Errorf("can't execute statement: %v", err)
	}
	return &result, nil
}
