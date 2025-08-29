package main

import (
    "fmt"
    "log"
    "net/http"
)

func main() {
    http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        fmt.Fprintf(w, `{"status":"healthy","service":"policy"}`)
    })
    
    log.Println("Policy service starting on :8005")
    if err := http.ListenAndServe(":8005", nil); err != nil {
        log.Fatal(err)
    }
}