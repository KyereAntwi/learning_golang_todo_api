package main

import "net/http"

const healthCheckRoute string = "/api/v1/healthcheck"
const todoRoute string = "/api/v1/todos"
const todoIDRoute string = "/api/v1/todos/"

func (app *application) routes() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc(healthCheckRoute, app.healthcheckHandler)
	mux.HandleFunc(todoRoute, app.createGetAllTodosHandler)
	mux.HandleFunc(todoIDRoute, app.getSingleUpdateDeleteTodoHandler)

	return mux
}
