package main

import (
    "errors"
    "fmt"
    "io"
    "encoding/json"
    "log"
    "math"
    "net/http"
    "strconv"
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
    address.Address = extractCommonAddress(&apiResponse)
    address.TokyoStaDistance, err = calcTokyoStaDistance(&apiResponse)
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        return
    }
    fmt.Printf("Response: %v\n", address)
}

func extractCommonAddress(resp *ApiResponse) string {
    var pref, city, town string
    if len(resp.Response.Location) > 0 {
        pref = resp.Response.Location[0].Prefecture
        city = resp.Response.Location[0].City
        town = resp.Response.Location[0].Town
    } else {
        return ""
    }
    for _, v := range resp.Response.Location {
        if v.Prefecture != pref {
            pref, city, town = "", "", ""
        }
        if v.City != city {
            city, town = "", ""
        }
        if v.Town != town {
            town = ""
        }
    }
    return pref + city + town
}

func calcTokyoStaDistance(resp *ApiResponse) (float64, error) {
    // Longitude and latitude of Tokyo Station and radius of the earth
    const tokyoStaX = 139.7673068
    const tokyoStaY = 35.6809591
    const r = 6371
    // -math.MaxFloat64 is the minimum of Float64
    maxDist := -math.MaxFloat64
    for _, v := range resp.Response.Location {
        x, errX := strconv.ParseFloat(v.X, 64)
        y, errY := strconv.ParseFloat(v.Y, 64)
        if errX != nil || errY != nil {
            return 0, errors.New("Failed in parsing float")
        }
        d := (math.Pi * r) / 180 * math.Sqrt(math.Pow((x - tokyoStaX) * math.Cos((math.Pi * (y + tokyoStaY)) / 360), 2) + math.Pow(y - tokyoStaY, 2))
        maxDist = math.Max(maxDist, d)
    }
    const base = 10
    return math.Round(maxDist * base) / base, nil
}

func main() {
    handleRequests()
}
