package main

import (
    "fmt"
    "log"
    "net/http"
)

type Address struct {
    postal_code string `json:"postal_code"`
    hit_count int64 `json:"hit_count"`
    address string `json:"address"`
    tokyo_sta_distance float64 `json:"tokyo_sta_distance"`
}

func homePage (w http.ResponseWriter, _ *http.Request) {
    fmt.Fprint(w, "Hello World!\n")
}

func handleRequests() {
    http.HandleFunc("/", homePage)
    log.Fatal(http.ListenAndServe(":8080", nil))
}

func main() {
    handleRequests()
}
