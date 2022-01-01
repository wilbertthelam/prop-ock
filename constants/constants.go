package constants

import "github.com/google/uuid"

// For now we treat only ONE league
var LEAGUE_ID = uuid.MustParse("894098e8-8cfe-4c92-9e32-332aac801899")

// Amount each user starts with by default
const STARTING_WALLET_AMOUNT = 500

// Send messages with this tag to allow us to continuously send
// them the auction players for bidding daily
const CONFIRM_TAG_UPDATE = "CONFIRMED_EVENT_UPDATE"

const WINNING_BID_TITLE = "You my new owner for only"
const CLAIM_INSTRUCTIONS = "Go claim me on ESPN. If no good, talk to Addy."

// Key for getting a transaction out of the Echo context
const TX = "transaction"
