package main

import (
	"fmt"
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

// @Summary Health Check
// @Description Check if the API is running and healthy
// @ID health-check
// @Accept json
// @Produce json
// @Success 200 {object} map[string]string "API status and info"
// @Router /health [get]
func (app *application) healthcheckHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	tracer := otel.Tracer("healthcheckHandler")
	_, span := tracer.Start(r.Context(), "Health Check")
	defer span.End()

	span.SetAttributes(
		attribute.String("http.method", r.Method),
		attribute.String("http.url", r.URL.Path),
	)
	fmt.Fprintln(w, "status: available")
	fmt.Fprintf(w, "environment: %s\n", app.config.env)
	fmt.Fprintf(w, "version: %s\n", app.config.version)
}
