package db

import (
    "fmt"
    "log"

    "github.com/jmoiron/sqlx"
    _ "github.com/lib/pq"

    "mFrelance/config"
    //"time"

)

var Postgres *sqlx.DB

func Connect() {
    cfg := config.AppConfig
    dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
        cfg.PostgresHost,
        cfg.PostgresPort,
        cfg.PostgresUser,
        cfg.PostgresPassword,
        cfg.PostgresDB,
    )

    var err error
    Postgres, err = sqlx.Connect("postgres", dsn)
    if err != nil {
        log.Fatal("Postgres connection failed:", err)
    }

    log.Println("Postgres connected")
}


