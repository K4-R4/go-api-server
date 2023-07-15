package main

import (
    "log"
    "net/http"

    "go-api-server/handler"
)


func handleRequests() {
    http.HandleFunc("/", handler.ReturnHomePage)
    http.HandleFunc("/address", handler.ReturnAddress)
    http.HandleFunc("/address/access_logs", handler.ReturnAccessLogs)
    log.Fatal(http.ListenAndServe(":8080", nil))
}

func main() {
    handleRequests()
}
