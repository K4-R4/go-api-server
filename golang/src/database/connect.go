package database

import (
    "database/sql"
    "fmt"
    "os"

    _ "github.com/go-sql-driver/mysql"
    _ "github.com/joho/godotenv"
)

type AccessLog struct {
    PostalCode   string `json:"postal_code"`
    RequestCount int    `json:"request_count"`
}

type AddressAccessLogs struct {
	AccessLogs []AccessLog `json:"access_logs"`
}

func Connect() (*sql.DB, error) {
    user := os.Getenv("MYSQL_USER")
    password := os.Getenv("MYSQL_PASSWORD")
    host := "mysql"
    port := os.Getenv("MYSQL_PORT")
    dbname := os.Getenv("MYSQL_DATABASE")

    dbconf := user + ":" + password + "@tcp(" + host + ":" + port +")/" + dbname + "?charset=utf8mb4"

    db, err := sql.Open("mysql", dbconf)
    if err != nil {
        fmt.Println(err.Error())
        return nil, err
    }
    return db, nil
}

func SaveAccessLog(postalCode string) error {
    db, err := Connect()
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

func GetAccessLogs() (AddressAccessLogs, error) {
    db, err := Connect()
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

