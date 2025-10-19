# Go Animals API

A simple REST API in Go (Gin) with MongoDB that supports CRUD, filtering, sorting, pagination, validation, Swagger docs, and Docker deployment.

## Features

- CRUD for animals: name, species, age, adopted
- Filtering by species/name/age/adopted
- Sorting by createdAt/name/age, asc/desc
- Pagination: page + limit
- Validation using go-playground/validator
- Swagger UI at /swagger/index.html
- Dockerfile + docker-compose (includes MongoDB)

## Prerequisites

- Go 1.22+
- Docker (optional but recommended for easy setup)

## Run locally

1. Start MongoDB quickly with Docker Compose (optional but easiest):

```pwsh
docker compose up -d mongo
```

2. Run the API:

```pwsh
go run ./...
```

The API listens on http://localhost:8080.

Health check: http://localhost:8080/health

Swagger UI: http://localhost:8080/swagger/index.html

## Docker

Build and run both MongoDB and the API:

```pwsh
docker compose up --build
```

## API

Base path: `/api/v1`

- POST `/animals`
- GET `/animals`
- GET `/animals/{id}`
- PUT `/animals/{id}`
- DELETE `/animals/{id}`

Example body:

```json
{
  "name": "Luna",
  "species": "cat",
  "age": 3,
  "adopted": false
}
```

Query params for listing:

- `species=cat`
- `name=lu` (contains, case-insensitive)
- `minAge=1&maxAge=5`
- `adopted=true`
- `sort=age&order=asc`
- `page=1&limit=10`

## Configuration

Environment variables:

- `PORT` (default 8080)
- `MONGO_URI` (default mongodb://localhost:27017)
- `MONGO_DB` (default goapi)

## Notes

- For richer Swagger docs, install `swag` CLI and run `swag init` to generate the docs from annotations.

# go-api
