default:
    @just --list

# Runs docker compose up -d and copies .env.example if .env is not found
up:
    #!/usr/bin/env sh
    if [ ! -f .env ]; then
        cp .env.example .env
    fi
    docker compose up -d

# docker compose down
down:
    docker compose down

# copies .env.example if .env doesnt exist, removes orphans, builds, and runs. Does not detach!
fresh:
    #!/usr/bin/env sh
    if [ ! -f .env ]; then
        cp .env.example .env
    fi
    docker compose down --remove-orphans
    docker compose build --no-cache
    docker compose up --build -V

# Runs docker compose logs -f. Requires up
logs: up
    docker compose logs -f

test:
    go test -v -race -cover -count=1 -failfast ./...

lint:
    golangci-lint run -v

# executes go run . migrate in the docker container
migrate:
    docker compose exec debateshare-server go run . migrate