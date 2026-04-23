package main

import (
	"fmt"
	"net/http"
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

	fmt.Fprintln(w, "status: available")
	fmt.Fprintf(w, "environment: %s\n", app.config.env)
	fmt.Fprintf(w, "version: %s\n", app.config.version)
}
