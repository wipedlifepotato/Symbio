package config

import (
	"github.com/joho/godotenv"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"log"
	"strconv"
	"time"
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

	MoneroHost     string
	MoneroPort     string
	MoneroUser     string
	MoneroPassword string

	MoneroAddress     string
	MoneroCommission  float64
	BitcoinAddress    string
	BitcoinCommission float64

	MaxProfiles     int64
	MaxAvatarSize   int64
	MaxAddrPerBlock int64

	WalletSyncInterval  time.Duration
	TxBlockInterval     time.Duration
	TxPoolFlushInterval time.Duration

	TaskMinInterval     time.Duration
	TaskDuplicateWindow time.Duration

	CaptchaEnabled             bool
	CaptchaRateLimitPerMinute  int
	CaptchaRateLimitPerHour    int
	CaptchaFontPath		   string
}

var AppConfig Config

func Init() {
	_ = godotenv.Load()

	// CLI flags
	pflag.String("config", "", "Path to config file")
	pflag.String("port", "", "Server port")
	pflag.String("listen_addr", "", "Listen address")

	pflag.String("electrum.host", "", "Electrum RPC host")
	pflag.Int("electrum.port", 0, "Electrum RPC port")
	pflag.String("electrum.user", "", "Electrum RPC user")
	pflag.String("electrum.password", "", "Electrum RPC password")

	pflag.String("monero.host", "", "Monero RPC host")
	pflag.Int("monero.port", 0, "Monero RPC port")
	pflag.String("monero.user", "", "Monero RPC user")
	pflag.String("monero.password", "", "Monero RPC password")

	pflag.String("postgres.host", "localhost", "Postgres host")
	pflag.String("postgres.port", "5432", "Postgres port")
	pflag.String("postgres.user", "user", "Postgres user")
	pflag.String("postgres.password", "password", "Postgres password")
	pflag.String("postgres.db", "db", "Postgres database")

	pflag.String("redis.host", "localhost", "Redis host")
	pflag.String("redis.port", "6379", "Redis port")
	pflag.String("redis.password", "", "Redis password")
	viper.SetDefault("wallet_sync_interval", "30s")
	viper.SetDefault("tx_block_interval", "1h")
	viper.SetDefault("tx_pool_flush_interval", "15s")
	viper.SetDefault("max_addr_per_block", 100)

	pflag.Parse()
	_ = viper.BindPFlags(pflag.CommandLine)

	configPath := viper.GetString("config")
	if configPath != "" {
		viper.SetConfigFile(configPath)
	} else {
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		viper.AddConfigPath(".")
		viper.AddConfigPath("./config")
	}

	viper.AutomaticEnv()

	viper.SetDefault("postgres.host", "localhost")
	viper.SetDefault("postgres.port", "5432")
	viper.SetDefault("postgres.user", "mfreelance")
	viper.SetDefault("postgres.password", "yourpassword091928374654ikd83km")
	viper.SetDefault("postgres.db", "mfreelance")

	viper.SetDefault("redis.host", "127.0.0.1")
	viper.SetDefault("redis.port", "6379")
	viper.SetDefault("redis.password", "")

	viper.SetDefault("jwt.token", "supersecrettoken123")
	viper.SetDefault("server.port", 9999)
	viper.SetDefault("listen_addr", "127.0.0.1")

	viper.SetDefault("electrum.host", "127.0.0.1")
	viper.SetDefault("electrum.port", 7777)
	viper.SetDefault("electrum.user", "Electrum")
	viper.SetDefault("electrum.password", "Electrum")

	viper.SetDefault("monero.host", "127.0.0.1")
	viper.SetDefault("monero.port", 28088)
	viper.SetDefault("monero.user", "monero")
	viper.SetDefault("monero.password", "rpcPassword")
	viper.SetDefault("monero.address", "9w49jr2CCtHcYkaVpwMX29Sq6AdnRXNTsZv85WWqwzCQYKmdF3ZggaiisJMFtci8LTBRNKkwpMfQ9g2qMMwr4De16Es8F4M")
	viper.SetDefault("monero.commission", 5)

	viper.SetDefault("bitcoin.address", "tb1q4zue4uyep4dgx96erac2ey3efdw2q6537wh3j7")
	viper.SetDefault("bitcoin.commission", 25)

	viper.SetDefault("max.profiles", 120)
	viper.SetDefault("max.avatar_size_mb", 2)
	viper.SetDefault("max.addr_per_block", 100)

	// Tasks
	viper.SetDefault("tasks.min_interval", "12m")
	viper.SetDefault("tasks.duplicate_window", "24h")

	// Captcha
	viper.SetDefault("captcha.enabled", true)
	viper.SetDefault("captcha.rate_limit_per_minute", 10)
	viper.SetDefault("captcha.rate_limit_per_hour", 100)

	if err := viper.ReadInConfig(); err != nil {
		log.Println("No config file found, falling back to defaults/env vars")
	} else {
		log.Println("Loaded config file:", viper.ConfigFileUsed())
	}

	AppConfig = Config{
		PostgresHost:     viper.GetString("postgres.host"),
		PostgresPort:     strconv.Itoa(viper.GetInt("postgres.port")),
		PostgresUser:     viper.GetString("postgres.user"),
		PostgresPassword: viper.GetString("postgres.password"),
		PostgresDB:       viper.GetString("postgres.db"),

		RedisHost:     viper.GetString("redis.host"),
		RedisPort:     strconv.Itoa(viper.GetInt("redis.port")),
		RedisPassword: viper.GetString("redis.password"),

		JWTToken:   viper.GetString("jwt.token"),
		ListenAddr: viper.GetString("listen_addr"),
		Port:       strconv.Itoa(viper.GetInt("server.port")),

		ElectrumHost:     viper.GetString("electrum.host"),
		ElectrumPort:     strconv.Itoa(viper.GetInt("electrum.port")),
		ElectrumUser:     viper.GetString("electrum.user"),
		ElectrumPassword: viper.GetString("electrum.password"),

		MoneroHost:       viper.GetString("monero.host"),
		MoneroPort:       strconv.Itoa(viper.GetInt("monero.port")),
		MoneroUser:       viper.GetString("monero.user"),
		MoneroPassword:   viper.GetString("monero.password"),
		MoneroAddress:    viper.GetString("monero.address"),
		MoneroCommission: viper.GetFloat64("monero.commission"),

		BitcoinAddress:    viper.GetString("bitcoin.address"),
		BitcoinCommission: viper.GetFloat64("bitcoin.commission"),

		MaxProfiles:     viper.GetInt64("max.profiles"),
		MaxAvatarSize:   viper.GetInt64("max.avatar_size_mb"),
		MaxAddrPerBlock: viper.GetInt64("max.addr_per_block"),

		WalletSyncInterval:  viper.GetDuration("wallet_sync_interval"),
		TxBlockInterval:     viper.GetDuration("tx_block_interval"),
		TxPoolFlushInterval: viper.GetDuration("tx_pool_flush_interval"),

		TaskMinInterval:     viper.GetDuration("tasks.min_interval"),
		TaskDuplicateWindow: viper.GetDuration("tasks.duplicate_window"),

		CaptchaEnabled:             viper.GetBool("captcha.enabled"),
		CaptchaRateLimitPerMinute:  viper.GetInt("captcha.rate_limit_per_minute"),
		CaptchaRateLimitPerHour:    viper.GetInt("captcha.rate_limit_per_hour"),
		CaptchaFontPath:	    viper.GetString("captcha.font_path"),
	}

	log.Println("Loaded commissions:", "BTC:", AppConfig.BitcoinCommission, "XMR:", AppConfig.MoneroCommission)
	log.Println("MaxProfiles:", AppConfig.MaxProfiles, "MaxAvatarSize:", AppConfig.MaxAvatarSize, "MaxAddrPerBlock:", AppConfig.MaxAddrPerBlock)
}
