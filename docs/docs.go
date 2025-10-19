package docs

// This file exists so the docs package builds. For richer docs, use swag init to generate.

var SwaggerInfo = struct{
    Title       string
    Description string
    Version     string
    Host        string
    BasePath    string
    Schemes     []string
}{
    Title:       "Animals API",
    Description: "A simple REST API for managing animals.",
    Version:     "1.0",
    Host:        "localhost:8080",
    BasePath:    "/api/v1",
    Schemes:     []string{"http"},
}
