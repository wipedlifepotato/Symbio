package db

import (
	"context"
	"log"
	//"time"

	"github.com/go-redis/redis/v8"

	"mFrelance/config"
)

var RedisClient *redis.Client
var Ctx = context.Background()

func ConnectRedis() {
	cfg := config.AppConfig
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     cfg.RedisHost + ":" + cfg.RedisPort,
		Password: cfg.RedisPassword,
		DB:       0,
	})

	_, err := RedisClient.Ping(Ctx).Result()
	if err != nil {
		log.Fatal("Redis connection failed:", err)
	}

	log.Println("Redis connected")
}
