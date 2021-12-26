package main

import (
	"github.com/go-redis/redis/v8"
	redis_client "github.com/wilbertthelam/prop-ock/db"
	"github.com/wilbertthelam/prop-ock/handlers/health"
	"github.com/wilbertthelam/prop-ock/handlers/message"
	auction_repo "github.com/wilbertthelam/prop-ock/repos/auction"
	league_repo "github.com/wilbertthelam/prop-ock/repos/league"
	user_repo "github.com/wilbertthelam/prop-ock/repos/user"
	auction_service "github.com/wilbertthelam/prop-ock/services/auction"
	callups_service "github.com/wilbertthelam/prop-ock/services/callups"
	league_service "github.com/wilbertthelam/prop-ock/services/league"
	user_service "github.com/wilbertthelam/prop-ock/services/user"
)

type DIModules = map[string]interface{}

func LoadModules() DIModules {
	modules := make(map[string]interface{})

	modules = registerClients(modules)
	modules = registerRepos(modules)
	modules = registerServices(modules)
	modules = registerHandlers(modules)

	return modules
}

// Add clients to the DI registry
func registerClients(modules DIModules) DIModules {
	modules[redis_client.GetName()] = redis_client.New()

	return modules
}

// Add repos to the DI registry
func registerRepos(modules DIModules) DIModules {
	modules[auction_repo.GetName()] = auction_repo.New(
		modules[redis_client.GetName()].(*redis.Client),
	)
	modules[user_repo.GetName()] = user_repo.New(
		modules[redis_client.GetName()].(*redis.Client),
	)
	modules[league_repo.GetName()] = league_repo.New(
		modules[redis_client.GetName()].(*redis.Client),
	)

	return modules
}

// Add services to the DI registry
func registerServices(modules DIModules) DIModules {
	modules[user_service.GetName()] = user_service.New(
		modules[user_repo.GetName()].(*user_repo.UserRepo),
	)

	modules[auction_service.GetName()] = auction_service.New(
		modules[auction_repo.GetName()].(*auction_repo.AuctionRepo),
		modules[user_service.GetName()].(*user_service.UserService),
	)
	modules[league_service.GetName()] = league_service.New(
		modules[league_repo.GetName()].(*league_repo.LeagueRepo),
	)
	modules[callups_service.GetName()] = callups_service.New(
		modules[redis_client.GetName()].(*redis.Client),
	)

	return modules
}

// Add handlers to the DI registry
func registerHandlers(modules DIModules) DIModules {
	modules[health.GetName()] = health.New()
	modules[message.GetName()] = message.New(
		modules[auction_service.GetName()].(*auction_service.AuctionService),
		modules[callups_service.GetName()].(*callups_service.CallupsService),
		modules[user_service.GetName()].(*user_service.UserService),
		modules[league_service.GetName()].(*league_service.LeagueService),
	)

	return modules
}
