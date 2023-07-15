package main

import (
    "database/sql"
    "errors"
    "fmt"
    "io"
    "encoding/json"
    "log"
    "math"
    "net/http"
    "os"
    "strconv"
    _ "github.com/go-sql-driver/mysql"
    _ "github.com/joho/godotenv"
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

type AddressAccessLogs struct {
	AccessLogs []AccessLog `json:"access_logs"`
}

type AccessLog struct {
    PostalCode   string `json:"postal_code"`
    RequestCount int    `json:"request_count"`
}

func returnHomePage(w http.ResponseWriter, _ *http.Request) {
    fmt.Fprint(w, "Hello World!\n")
}

func handleRequests() {
    http.HandleFunc("/", returnHomePage)
    http.HandleFunc("/address", returnAddress)
    http.HandleFunc("/address/access_logs", returnAccessLogs)
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
        fmt.Printf("get error\n")
        w.WriteHeader(http.StatusInternalServerError)
        return
    }
    defer resp.Body.Close()
    byteArray, _ := io.ReadAll(resp.Body)
    apiResponse := ApiResponse{}
    err = json.Unmarshal(byteArray, &apiResponse)
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        fmt.Printf("unmarshal error\n")
        return
    }

    address := Address{}
    address.PostalCode = param
    address.HitCount = len(apiResponse.Response.Location)
    address.Address = extractCommonAddress(&apiResponse)
    address.TokyoStaDistance, err = calcTokyoStaDistance(&apiResponse)
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        fmt.Printf("parsing float error\n")
        return
    }
    err = saveAccessLog(address.PostalCode)
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        fmt.Printf("save access log error\n")
        return
    }
    json.NewEncoder(w).Encode(address)
}

func returnAccessLogs(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json; charset=utf-8")

    // Method validation
    if r.Method != http.MethodGet {
        w.WriteHeader(http.StatusMethodNotAllowed)
        return
    }

    logs, err := getAccessLogs()
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        return
    }
    json.NewEncoder(w).Encode(logs)
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

func connect() (*sql.DB, error) {
    user := os.Getenv("MYSQL_USER")
    password := os.Getenv("MYSQL_PASSWORD")
    host := "mysql"
    port := os.Getenv("MYSQL_PORT")
    dbname := os.Getenv("MYSQL_DATABASE")

    dbconf := user + ":" + password + "@tcp(" + host + ":" + port +")/" + dbname + "?charset=utf8mb4"
    fmt.Printf(dbconf)

    db, err := sql.Open("mysql", dbconf)
    if err != nil {
        fmt.Printf("OPEN ERROR\n")
        return nil, err
    }
    err = db.Ping()
    if err != nil {
        fmt.Printf("PING ERROR\n")
    }
    return db, nil
}

func saveAccessLog(postalCode string) error {
    db, err := connect()
    defer db.Close()
    if err != nil {
        return err
    }
    _, err = db.Exec(`
        INSERT INTO
            access_logs(postal_code) VALUES(?)`, postalCode)
    if err != nil {
        return err
    }
    return nil
}

func getAccessLogs() (AddressAccessLogs, error) {
    db, err := connect()
    defer db.Close()
    if err != nil {
        return AddressAccessLogs{}, err
    }

    rows, err := db.Query(`
        SELECT
            postal_code, COUNT(id)
        FROM
            access_logs
        GROUP BY
            postal_code
        ORDER BY
            COUNT(id) DESC`)
    defer rows.Close()
    if err != nil {
        return AddressAccessLogs{}, err
    }

    logs := AddressAccessLogs{}
    logs.AccessLogs = make([]AccessLog, 0)
    for rows.Next() {
        log := AccessLog{}
         err := rows.Scan(&log.PostalCode, &log.RequestCount)
        if err != nil {
            return AddressAccessLogs{}, err
        }
        logs.AccessLogs = append(logs.AccessLogs, log)
    }
    return logs, nil
}

func main() {
    handleRequests()
}
