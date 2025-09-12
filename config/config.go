package config

import (
    "log"
    "os"
    "strconv"
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
    JWTToken         string
    ListenAddr       string

    ElectrumHost     string
    ElectrumPort     string
    ElectrumUser     string
    ElectrumPassword string

    MoneroHost       string
    MoneroPort       string
    MoneroUser       string
    MoneroPassword   string
}

var AppConfig Config

func MustAtoi(s string) int {
    i, err := strconv.Atoi(s)
    if err != nil {
        log.Fatalf("Invalid port: %v", err)
    }
    return i
}

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
        JWTToken:         os.Getenv("JWT_TOKEN"),
        ListenAddr:       os.Getenv("LISTEN_ADDR"),

        ElectrumHost:     os.Getenv("ELECTRUM_HOST"),
        ElectrumPort:     os.Getenv("ELECTRUM_PORT"),
        ElectrumUser:     os.Getenv("ELECTRUM_USER"),
        ElectrumPassword: os.Getenv("ELECTRUM_PASSWORD"),

        MoneroHost:       os.Getenv("MONERO_HOST"),
        MoneroPort:       os.Getenv("MONERO_PORT"),
        MoneroUser:       os.Getenv("MONERO_USER"),
        MoneroPassword:   os.Getenv("MONERO_PASSWORD"),
    }
}

