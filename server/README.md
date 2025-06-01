
# Setup 
1. Install go
   - [https://go.dev/doc/install](https://go.dev/doc/install)
2. Install dependencies 
   - `go mod tidy`
3. copy `.env.example` to `.env` and set the environment variables as needed.

# Run the server 

## locally 
1. run the caches and db
    - `docker compose up user-offer-cache offer-cache share-db -d`
2. start the server
    - `go run .`

## in docker

`docker compose up -d`