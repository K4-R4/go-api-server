package main

import (
    "fmt"
    "io"
    "encoding/json"
    "log"
    "net/http"
)

type ApiResponse struct {
	Response struct {
		Location []struct {
			City       string `json:"city"`
			CityKana   string `json:"city_kana"`
			Town       string `json:"town"`
			TownKana   string `json:"town_kana"`
			X          string `json:"x"`
			Y          string `json:"y"`
			Prefecture string `json:"prefecture"`
			Postal     string `json:"postal"`
		} `json:"location"`
	} `json:"response"`
}

type Address struct {
    PostalCode string `json:"postal_code"`
    HitCount int `json:"hit_count"`
    Address string `json:"address"`
    TokyoStaDistance float64 `json:"tokyo_sta_distance"`
}

func returnHomePage(w http.ResponseWriter, _ *http.Request) {
    fmt.Fprint(w, "Hello World!\n")
}

func handleRequests() {
    http.HandleFunc("/", returnHomePage)
    http.HandleFunc("/address", returnAddress)
    log.Fatal(http.ListenAndServe(":8080", nil))
}

func returnAddress(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json; charset=utf-8")

    // Method validation
    if r.Method != http.MethodGet {
        w.WriteHeader(http.StatusMethodNotAllowed)
        return
    }

    // Param validation
    param := r.URL.Query().Get("post_code")
    if len(param) == 0 {
        w.WriteHeader(http.StatusBadRequest)
        return
    }

    resp, err := http.Get("https://geoapi.heartrails.com/api/json?method=searchByPostal&postal=" + param)
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        return
    }
    defer resp.Body.Close()
    byteArray, _ := io.ReadAll(resp.Body)
    fmt.Printf("log: %s\n", byteArray)
    apiResponse := ApiResponse{}
    err = json.Unmarshal(byteArray, &apiResponse)
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        return
    }

    address := Address{}
    address.PostalCode = param
    address.HitCount = len(apiResponse.Response.Location)
    fmt.Printf("log: %v\n", len(apiResponse.Response.Location))
}

func main() {
    handleRequests()
}
