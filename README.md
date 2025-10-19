# Go Animals API

## What’s included

- Animals CRUD with extended fields: name, species, age, adopted, image, owner, location (GeoJSON Point)
- Categories CRUD (name/category_name)
- Species CRUD (name/species_name, category)
- Advanced features (≥3):
  - Flexible filtering (name or animal_name, species string or ObjectId, adopted)
  - Sorting (createdAt, name, age, birthdate if present, animal_name)
  - Pagination (page + limit)
  - Validation using go-playground/validator
- Swagger UI at `/swagger/index.html` (served from static `/openapi/doc.json`)
- Dockerfile + docker-compose (includes MongoDB service for local dev)

## Quickstart (Docker)

1. Copy env file and fill in values (Atlas or local). For Atlas, set MONGO_URI to your connection string; for local compose Mongo, use mongodb://mongo:27017.

2. Start the stack:

```pwsh
docker compose up -d --build
```

App URLs:

- API base: http://localhost:8080/api/v1
- Health: http://localhost:8080/health
- Swagger UI: http://localhost:8080/swagger/index.html
- OpenAPI JSON: http://localhost:8080/openapi/doc.json

If you’re using Atlas only, you can choose to start only the API service:

```pwsh
docker compose up -d --build api
```

## Run natively (without Docker)

Prereqs: Go 1.22+, a running MongoDB (local or Atlas), and a `.env` file.

```pwsh
# Ensure Mongo is reachable (e.g., local: mongodb://localhost:27017 or Atlas URI in .env)
go run ./...
```

## API overview

Base path: `/api/v1`

Animals

- POST `/animals`
- GET `/animals`
- GET `/animals/{id}`
- PUT `/animals/{id}`
- DELETE `/animals/{id}`

Categories

- POST `/categories`
- GET `/categories`
- GET `/categories/{id}`
- PUT `/categories/{id}`
- DELETE `/categories/{id}`

Species

- POST `/species`
- GET `/species`
- GET `/species/{id}`
- PUT `/species/{id}`
- DELETE `/species/{id}`

### Animals request examples

Minimal (our schema):

```json
{
  "name": "Luna",
  "species": "cat",
  "age": 3,
  "adopted": false
}
```

Dataset-style (aliases supported) + geo:

```json
{
  "animal_name": "Gustave",
  "species": "642d1e873e9c108f66a50009",
  "birthdate": "2001-04-03",
  "adopted": false,
  "image": "https://example.com/img/gator.jpg",
  "owner": "City Shelter",
  "location": { "type": "Point", "coordinates": [24.94, 60.17] }
}
```

List query params:

- `species=cat` or `species=642d1e...` (matches string or ObjectId)
- `name=lu` (contains; matches `name` or `animal_name`, case-insensitive)
- `minAge=1&maxAge=5`
- `adopted=true`
- `sort=age|name|createdAt|birthdate|animal_name` and `order=asc|desc`
- `page=1&limit=10`

Notes:

- If `birthdate` exists, age is derived when not provided.
- `location.coordinates` follows GeoJSON order: [longitude, latitude].

## Swagger/OpenAPI

- UI: `/swagger/index.html`
- Spec JSON: `/openapi/doc.json` (static file at `docs/swagger.json`)

## Development

- Health check: GET `/health`
