package controllers

import (
	"net/http"
	"os"
	"path/filepath"
)

const swaggerUIHTML = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8" />
  <title>X Clone API Documentation</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css" />
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script>
    window.onload = () => SwaggerUIBundle({ url: "/openapi.yaml", dom_id: "#swagger-ui" });
  </script>
</body>
</html>`

func SwaggerUI(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(swaggerUIHTML))
}

func OpenAPISpec(w http.ResponseWriter, _ *http.Request) {
	spec, err := readOpenAPISpec()
	if err != nil {
		http.Error(w, "OpenAPI specification is unavailable", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/yaml; charset=utf-8")
	_, _ = w.Write(spec)
}

func readOpenAPISpec() ([]byte, error) {
	paths := []string{
		filepath.Join("openapi", "openapi.yaml"),
		filepath.Join("backend", "openapi", "openapi.yaml"),
		filepath.Join("..", "..", "openapi", "openapi.yaml"),
	}

	for _, path := range paths {
		spec, err := os.ReadFile(path)
		if err == nil {
			return spec, nil
		}
	}

	return nil, os.ErrNotExist
}
