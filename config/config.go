package config

import (
    "log"
    "os"

    "github.com/joho/godotenv"
)

type Config struct {
    PostgresHost     string
    PostgresPort     string
    PostgresUser     string
    PostgresPassword string
    PostgresDB       string
    RedisHost        string
    RedisPort        string
    RedisPassword    string
    Port             string
}

var AppConfig Config

func Init() {
    err := godotenv.Load()
    if err != nil {
        log.Println("No .env file found, using system environment")
    }

    AppConfig = Config{
        PostgresHost:     os.Getenv("POSTGRES_HOST"),
        PostgresPort:     os.Getenv("POSTGRES_PORT"),
        PostgresUser:     os.Getenv("POSTGRES_USER"),
        PostgresPassword: os.Getenv("POSTGRES_PASSWORD"),
        PostgresDB:       os.Getenv("POSTGRES_DB"),
        RedisHost:        os.Getenv("REDIS_HOST"),
        RedisPort:        os.Getenv("REDIS_PORT"),
        RedisPassword:    os.Getenv("REDIS_PASSWORD"),
        Port:             os.Getenv("PORT"),
    }
}
