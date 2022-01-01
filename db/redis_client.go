package redis_client

import (
	"github.com/go-redis/redis/v8"
	"github.com/labstack/echo/v4"
	"github.com/wilbertthelam/prop-ock/constants"
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

func GetCmdable(context echo.Context, redisClient redis.Cmdable) redis.Cmdable {
	tx := context.Get(constants.TX)

	var client redis.Cmdable
	if tx == nil {
		client = redisClient
	} else {
		client = tx.(redis.Cmdable)
	}

	return client
}

func StartTransaction(context echo.Context, redisClient *redis.Client, commandList func() error) error {
	return redisClient.Watch(
		context.Request().Context(),
		func(tx *redis.Tx) error {
			context.Set(constants.TX, tx)

			err := commandList()

			// Clear transaction from context once it's finished
			// so further Redis calls in the request chain don't
			// try to use this context again
			context.Set(constants.TX, nil)

			return err
		},
	)
}
