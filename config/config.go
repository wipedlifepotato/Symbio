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
    
    MoneroAddress string
    MoneroCommission float64
    BitcoinAddress string
    BitcoinCommission float64
    
    MaxProfiles	int64
    MaxAvatarSize int64
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

    moneroCommStr := os.Getenv("MONERO_COMMISSION")
    bitcoinCommStr := os.Getenv("BITCOIN_COMMISSION")
    maxProfilesStr := os.Getenv("MAX_PROFILES")
    maxAvatarSizeStr := os.Getenv("MAX_AVATAR_SIZE_MB")

    moneroComm, err := strconv.ParseFloat(moneroCommStr, 64)
    if err != nil {
        log.Printf("Invalid MONERO_COMMISSION, using 5: %v", err)
        moneroComm = 5
    }

    bitcoinComm, err := strconv.ParseFloat(bitcoinCommStr, 64)
    if err != nil {
        log.Printf("Invalid BITCOIN_COMMISSION, using 5: %v", err)
        bitcoinComm = 5
    }
    maxProfiles, err := strconv.ParseInt(maxProfilesStr, 10, 64)
    if err != nil {
        log.Printf("Invalid maxProfilesStr, using 25: %v", err)
        maxProfiles = 25
    }
    maxAvatarSize, err := strconv.ParseInt(maxAvatarSizeStr, 10, 64)
    if err != nil {
        log.Printf("Invalid maxAvatarSizeStr, using 2: %v", err)
        maxProfiles = 2
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

        MoneroAddress:    os.Getenv("MONERO_ADDR"),
        MoneroCommission: moneroComm,
        BitcoinAddress:   os.Getenv("BITCOIN_ADDR"),
        BitcoinCommission: bitcoinComm,
        MaxProfiles:	maxProfiles,
        MaxAvatarSize:	maxAvatarSize,
    }

    log.Println("Loaded commissions:", "BTC:", AppConfig.BitcoinCommission, "XMR:", AppConfig.MoneroCommission)
}

