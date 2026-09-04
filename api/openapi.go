package api

import _ "embed"

// OpenAPISpec is the API contract served by the Swagger UI and the API server.
//
//go:embed openapi.yaml
var OpenAPISpec []byte
