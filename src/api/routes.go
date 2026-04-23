package main

import "net/http"

const healthCheckRoute string = "/api/v1/healthcheck"
const todoRoute string = "/api/v1/todos"
const todoIDRoute string = "/api/v1/todos/"
const signUpRoute string = "/api/v1/auth/signup"
const signInRoute string = "/api/v1/auth/signin"
const refreshTokenRoute string = "/api/v1/auth/refresh"
const swaggerRoute string = "/swagger/"

func (app *application) routes() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc(healthCheckRoute, app.healthcheckHandler)
	mux.HandleFunc(todoRoute, app.createGetAllTodosHandler)
	mux.HandleFunc(todoIDRoute, app.getSingleUpdateDeleteTodoHandler)
	mux.HandleFunc(signUpRoute, app.signUpRouteHandler)
	mux.HandleFunc(signInRoute, app.signInRouteHandler)
	mux.HandleFunc(refreshTokenRoute, app.refreshTokenRouteHandler)
	mux.Handle(swaggerRoute, SwaggerHandler())

	return mux
}
