# prop-ock

Blind bidding fantasy auction for MLB-debuting baseball prospects.

## Architecture

Prop-ock is a single deployable Go binary containing 4 discrete functionalities:

- Core API server (CRUD operations for all entities) (`/api/*`)
- Daily batch job for fetching and transforming new prospect entities
- Facebook Messenger chat bot (`/message/*`)
- Lightweight UI for webviews (`/public/*`)

All data is currently stored on a Redis instance on Kamatera Cloud.

Dependency injection is managed using [wire](https://github.com/google/wire).

## Development

### Setup

- Make sure you have `Go v1.17+` installed
- Run `go install` to download and install necessary dependencies
- Local secrets and credentials are not publically available

### Build

- Run `go run .` to boot up the server
- If you've added any new Dependency Injected modules, run `$GOPATH/bin/wire` to regenerate DI scaffolding code

### Testing

- To hit a route, use `CURL` or any HTTP client using `localhost:8000/[route]`
- Full API route paths are listed in `main.go`

## Production

Production servers are hosted on Heroku.

### Deployments

Deployments are made using continuous deployment to a Heroku dyno instance. This is triggered by any remote push to `main` on this Github repo. A new binary will be created by Heroku before every deploy (pinned to GOVERSION 1.17).

### Secrets/environment variables

Secrets are stored as environment variables on Heroku servers. The server will panic if secrets are not properly configured on server startup.
