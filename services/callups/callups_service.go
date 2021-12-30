package callups_service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/go-redis/redis/v8"
	"github.com/labstack/echo/v4"
)

type TransactionsResponse struct {
	TransactionsAll TransactionsAll `json:"transaction_all"`
}

type TransactionsAll struct {
	Results TransactionsResult `json:"queryResults"`
}

type TransactionsResult struct {
	Transactions []Transaction `json:"row"`
	Size         int           `json:"totalSize,string"`
	CreatedAt    string        `json:"created"`
}

type Transaction struct {
	Id         string `json:"transaction_id"`
	PlayerId   string `json:"player_id"`
	PlayerName string `json:"player"`
	Team       string `json:"team"`
	TeamId     string `json:"team_id"`
	Date       string `json:"trans_date"`
	Type       string `json:"type_cd"`
	FullType   string `json:"type"`
}

type PeopleResponse struct {
	People []Person `json:"people"`
}

type Person struct {
	Id        int     `json:"id"`
	FullName  string  `json:"fullName"`
	DebutDate *string `json:"mlbDebutDate"`
}

type CallupsService struct {
	redisClient *redis.Client
}

func New(redisClient *redis.Client) *CallupsService {
	return &CallupsService{
		redisClient,
	}
}

func (c *CallupsService) GetLatestCallups(context echo.Context) error {
	allTransactions, err := getTransactionsFromMLB("20210901", "20211126")
	if err != nil {
		return context.JSON(http.StatusServiceUnavailable, fmt.Sprintf("failed to get most recent transactions: %+v", err))
	}

	if allTransactions != nil && allTransactions.TransactionsAll.Results.Size < 1 {
		return context.JSON(http.StatusNotFound, "no results retrieved")
	}

	// If these are callups (type "CU"), check if they are making their debut
	debutTransactions, err := c.getDebutTransactions(context, allTransactions.TransactionsAll.Results.Transactions)
	if err != nil {
		return context.JSON(http.StatusServiceUnavailable, fmt.Sprintf("failed to get debut transactions: %+v", err))
	}

	return context.JSON(http.StatusOK, debutTransactions)
}

func (c *CallupsService) getDebutTransactions(context echo.Context, transactions []Transaction) ([]Transaction, error) {
	var debutTransactions []Transaction

	fmt.Printf("trans: %+v", transactions)

	for _, transaction := range transactions {
		// Filter by recalls and contracted selected
		// Recalls (CU) are for those in the 40-man roster but may not have debuted yet.
		// Selected (SE) are for those who are added to the 40-man roster
		if transaction.Type != "CU" && transaction.Type != "SE" {
			continue
		}

		// Check if this is the debut of the player
		isDebut, err := c.isPlayerDebut(context, transaction.PlayerId)
		if err != nil {
			return nil, err
		}
		if isDebut {
			debutTransactions = append(debutTransactions, transaction)
		}
	}

	return debutTransactions, nil
}

type RedisPlayer struct {
	PlayerId string
}

func generatePlayerDebutRedisKey(playerId string) string {
	return "debut:" + playerId
}

// Check Redis cache to see if we know this player has already debuted
// If no results, make a call to MLB API to check if they have debuted
func (c *CallupsService) isPlayerDebut(context echo.Context, playerId string) (bool, error) {
	if playerId == "" {
		return false, nil
	}

	// Check Redis to see if playerId exists already
	playerExists, err := c.redisClient.Get(context.Request().Context(), generatePlayerDebutRedisKey(playerId)).Result()
	if err != nil {
		fmt.Println("error grabbing from Redis")
		return false, nil
	}

	if playerExists != "" && playerExists != "false" {
		return true, nil
	}

	return false, nil

	// playerAPIURL := "https://statsapi.mlb.com/api/v1/people/" + playerId

	// resp, err := http.Get(playerAPIURL)
	// if err != nil {
	// 	return false, err
	// }

	// defer resp.Body.Close()
	// body, err := io.ReadAll(resp.Body)
	// if err != nil {
	// 	return false, err
	// }

	// var peopleResponse PeopleResponse
	// if err := json.Unmarshal(body, &peopleResponse); err != nil {
	// 	return false, err
	// }

	// if peopleResponse.People != nil && len(peopleResponse.People) < 1 {
	// 	return false, nil
	// }

	// // If person's debut hasn't been made yet, this is their first callup
	// person := peopleResponse.People[0]
	// if person.DebutDate == nil || *person.DebutDate == "" {
	// 	return true, nil
	// }

	// // If their debut day is today, this is their first callup
	// listedDebutDate, err := time.Parse("2006-01-02", *person.DebutDate)
	// if err != nil {
	// 	return false, err
	// }

	// // year, month, day := time.Now().Date()
	// fmt.Printf("listedDebutDate: %+v", listedDebutDate)

	// // combinedDate := string(year) + "" + string(month) + "" + string(day)

	// return true, nil
}

func getTransactionsFromMLB(startDate string, endDate string) (*TransactionsResponse, error) {
	u := url.Values{}
	u.Add("sport_code", "'mlb'")
	u.Add("start_date", startDate)
	u.Add("end_state", endDate)

	espnTransactionsURL := "http://lookup-service-prod.mlb.com/json/named.transaction_all.bam?" + u.Encode()

	resp, err := http.Get(espnTransactionsURL)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var transactionsResponse *TransactionsResponse
	if err := json.Unmarshal(body, &transactionsResponse); err != nil {
		return nil, err
	}

	return transactionsResponse, nil
}
