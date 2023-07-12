package main

import (
    "fmt"
    "io"
    "log"
    "net/http"
)

type Address struct {
    PostalCode string `json:"postal_code"`
    HitCount int64 `json:"hit_count"`
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
    fmt.Printf("log: %s", byteArray)
}

/*
func doRequest(method, path string, values url.Values, body io.Reader) ([]byte, error) {
    client := &http.Client{
        Timeout: 20 * time.Second,
    }

    req, err := http.NewRequest(method, path, body)
    if err != nil {
        return nil, err
    }
    req.URL.RawQuery = values.Encode()

    resp, err := client.Do(req)
    if err != nil {
        return nil, err
    }

    data, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return nil, err
    }

    return data, nil
}
*/

func main() {
    handleRequests()
}
