package main

import (
    "fmt"
    "log"
    "net/http"
)

func main() {
    http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        fmt.Fprintf(w, `{"status":"healthy","service":"optimizer"}`)
    })
    
    log.Println("Optimizer service starting on :8004")
    if err := http.ListenAndServe(":8004", nil); err != nil {
        log.Fatal(err)
    }
}