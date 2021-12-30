package redis_client

import (
	"github.com/go-redis/redis/v8"
	"github.com/wilbertthelam/prop-ock/secrets"
)

func New() *redis.Client {
	db := redis.NewClient(&redis.Options{
		Addr:     secrets.HOST,
		Password: secrets.PASSWORD,
		DB:       0, // use default DB
	})

	return db
}
