# syntax=docker/dockerfile:1
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod .
RUN go mod download
COPY . .
RUN go mod tidy && CGO_ENABLED=0 GOOS=linux go build -o /out/app

FROM gcr.io/distroless/base-debian12
WORKDIR /
COPY --from=builder /out/app /app
COPY --from=builder /app/docs/swagger.json /docs/swagger.json
ENV PORT=8080
ENV MONGO_URI=mongodb://mongo:27017
ENV MONGO_DB=goapi
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/app"]
