package redis_client

import (
	"github.com/go-redis/redis/v8"
	"github.com/labstack/echo/v4"
	"github.com/wilbertthelam/prop-ock/constants"
	config_service "github.com/wilbertthelam/prop-ock/services/config"
)

func New(config *config_service.Config) *redis.Client {
	redisConfig := config.GetRedisConfig()

	db := redis.NewClient(&redis.Options{
		Addr:     redisConfig.HostAddress + ":" + redisConfig.Port,
		Password: redisConfig.Password,
		DB:       0, // use default DB
	})

	return db
}

// GetCmdable calls a Redis command using either:
// 	(1) the default instantiate Redis client
// 	(2) an open transaction
// Only use the transaction instance if the context object
// has an open transaction on it.
func GetCmdable(
	context echo.Context,
	redisClient redis.Cmdable,
) redis.Cmdable {
	tx := context.Get(constants.TX)

	var client redis.Cmdable
	if tx == nil {
		client = redisClient
	} else {
		client = tx.(redis.Cmdable)
	}

	return client
}

// StartTransaction creates a Redis transaction and takes in a function
// containing all the commands
func StartTransaction(
	context echo.Context,
	redisClient *redis.Client,
	commandList func() error,
) error {
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
