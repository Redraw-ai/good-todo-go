package main

import (
    "database/sql"
    "fmt"
    "log"

    _ "github.com/lib/pq"
)

func main() {
    connStr := "postgres://goodtodo:secret@localhost:5434/goodtodo_dev?sslmode=disable"
    
    db, err := sql.Open("postgres", connStr)
    if err != nil {
        log.Fatalf("sql.Open failed: %v", err)
    }
    defer db.Close()
    
    fmt.Println("Pinging...")
    if err := db.Ping(); err != nil {
        log.Fatalf("Ping failed: %v", err)
    }
    fmt.Println("Ping successful!")
    
    var result int
    if err := db.QueryRow("SELECT 1").Scan(&result); err != nil {
        log.Fatalf("Query failed: %v", err)
    }
    fmt.Printf("Query result: %d\n", result)
}
