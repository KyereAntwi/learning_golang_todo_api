package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"todoapi.com/m/src/domain"
	"todoapi.com/m/src/repositories"
)

// @Summary Create or Get All Todos
// @Description Create a new todo or get all todos
// @ID create-get-all-todos
// @Accept json
// @Produce json
// @Param todo body domain.CreateTodoDto true "Create Todo"
// @Success 200 {array} domain.TodoDto
// @Success 201 {object} map[string]interface{} "Created Todo ID"
// @Router /todos [get]
// @Router /todos [post]
func (app *application) createGetAllTodosHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		app.getAllTodos(w, r)
	case http.MethodPost:
		app.createTodo(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// @Summary Get, Update, or Delete a Todo
// @Description Get, update, or delete a todo by ID
// @ID get-update-delete-todo
// @Accept json
// @Produce json
// @Param id path int true "Todo ID"
// @Param todo body domain.UpdateTodoDto true "Update Todo"
// @Success 200 {object} domain.TodoDetailDto
// @Success 204 "No Content"
// @Router /todos/{id} [get]
// @Router /todos/{id} [put]
// @Router /todos/{id} [delete]
func (app *application) getSingleUpdateDeleteTodoHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		app.getTodo(w, r)
	case http.MethodPut:
		app.updateTodo(w, r)
	case http.MethodDelete:
		app.deleteTodo(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (app *application) getAllTodos(w http.ResponseWriter, r *http.Request) {

	tracer := otel.Tracer("GetAllTodosHandler")

	ctx, span := tracer.Start(r.Context(), "GetAllTodosHandler")
	defer span.End()

	r = r.WithContext(ctx)

	if app.db == nil {
		app.logger.Printf("Database connection is nil")
		http.Error(w, "Server Error", http.StatusInternalServerError)
		return
	}

	searchKey := r.URL.Query().Get("search_key")
	page := r.URL.Query().Get("page")
	pageSize := r.URL.Query().Get("page_size")

	var todoRepo repositories.ITodoRepository = repositories.NewTodoRepository(app.db, app.rabbitConn)

	pageInt, err := strconv.ParseInt(page, 10, 64)
	if err != nil {
		span.AddEvent("Invalid page query parameter, defaulting to 1")
		span.SetStatus(codes.Error, err.Error())
		pageInt = 1
	}

	pageSizeInt, err := strconv.ParseInt(pageSize, 10, 64)
	if err != nil {
		span.AddEvent("Invalid page_size query parameter, defaulting to 10")
		span.SetStatus(codes.Error, err.Error())
		pageSizeInt = 10
	}

	loggedInUserID := r.Context().Value("userID")
	if loggedInUserID == nil {
		app.logger.Printf("User ID not found in context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userUUID, err := uuid.Parse(loggedInUserID.(string))
	if err != nil {
		span.AddEvent("Error parsing user ID from context")
		span.SetStatus(codes.Error, err.Error())
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	todos, err := todoRepo.GetAll(searchKey, userUUID, pageInt, pageSizeInt)
	if err != nil {
		app.logger.Printf("Error fetching todos: %v", err)
		span.AddEvent("Error fetching todos")
		span.SetStatus(codes.Error, err.Error())
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	todosResponse := make([]domain.TodoDto, len(todos))

	for i, todo := range todos {
		todosResponse[i] = domain.TodoDto{
			ID:          todo.GetID(),
			Title:       todo.GetTitle(),
			Description: todo.GetDescription(),
			Status:      todo.GetStatus(),
		}
	}

	todoJson, err := json.Marshal(todosResponse)
	if err != nil {
		app.logger.Printf("Error marshaling todos response: %v", err)
		span.AddEvent("Error marshaling todos response")
		span.SetStatus(codes.Error, err.Error())
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	w.Write(todoJson)
}

func (app *application) createTodo(w http.ResponseWriter, r *http.Request) {

	tracer := otel.Tracer("CreateTodoHandler")
	ctx, span := tracer.Start(r.Context(), "CreateTodoHandler")
	defer span.End()
	r = r.WithContext(ctx)

	if app.db == nil {
		app.logger.Printf("Database connection is nil")
		http.Error(w, "Server Error", http.StatusInternalServerError)
		return
	}

	var createTodoDto domain.CreateTodoDto
	err := json.NewDecoder(r.Body).Decode(&createTodoDto)
	if err != nil {
		app.logger.Printf("Error decoding create todo request: %v", err)
		span.AddEvent("Error decoding create todo request")
		span.SetStatus(codes.Error, err.Error())
		http.Error(w, "Server Error", http.StatusInternalServerError)
		return
	}

	// log body content for debugging
	loggedInUserID := r.Context().Value("userID")
	span.SetAttributes(attribute.String("user_id", loggedInUserID.(string)))
	span.AddEvent(fmt.Sprintf("CreateTodo request body: %+v", createTodoDto))
	span.SetStatus(codes.Ok, fmt.Sprintf("CreateTodo request body: %+v", createTodoDto))

	var todoRepo repositories.ITodoRepository = repositories.NewTodoRepository(app.db, app.rabbitConn)

	userUUID, err := uuid.Parse(loggedInUserID.(string))
	if err != nil {
		app.logger.Printf("Error parsing user ID: %v", err)
		span.AddEvent("Error parsing user ID from context")
		span.SetStatus(codes.Error, err.Error())
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	exists, err := todoRepo.DoesTodoExist(createTodoDto.Title, userUUID)

	if err != nil {
		app.logger.Printf("Error checking if todo exists: %v", err)
		span.AddEvent("Error checking if todo exists")
		span.SetStatus(codes.Error, err.Error())
		http.Error(w, "Server Error", http.StatusInternalServerError)
		return
	}

	if exists {
		http.Error(w, "Todo with the same title already exists", http.StatusConflict)
		return
	}

	if loggedInUserID == nil {
		app.logger.Printf("User ID not found in context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	id, err := todoRepo.Create(createTodoDto.Title, createTodoDto.Description, userUUID)
	if err != nil {
		app.logger.Printf("Error creating todo: %v", err)
		span.AddEvent("Error creating todo")
		span.SetStatus(codes.Error, err.Error())
		http.Error(w, "Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Location", fmt.Sprintf("%s/%d", r.URL.Path, id))
	w.WriteHeader(http.StatusCreated)

	response := map[string]interface{}{
		"id": id,
	}

	responseJson, err := json.Marshal(response)
	if err != nil {
		app.logger.Printf("Error marshaling create todo response: %v", err)
		span.AddEvent("Error marshaling create todo response")
		span.SetStatus(codes.Error, err.Error())
		http.Error(w, "Server Error", http.StatusInternalServerError)
		return
	}

	w.Write(responseJson)
}

func (app *application) getTodo(w http.ResponseWriter, r *http.Request) {
	tracer := otel.Tracer("GetTodoHandler")
	ctx, span := tracer.Start(r.Context(), "GetTodoHandler")
	defer span.End()
	r = r.WithContext(ctx)

	if app.db == nil {
		app.logger.Printf("Database connection is nil")
		http.Error(w, "Server Error", http.StatusInternalServerError)
		return
	}

	idStr := r.URL.Path[len(todoIDRoute):]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		span.AddEvent("Invalid ID")
		span.SetStatus(codes.Error, err.Error())
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var todoRepo repositories.ITodoRepository = repositories.NewTodoRepository(app.db, app.rabbitConn)

	todo, err := todoRepo.GetTodoByID(id)
	if err != nil {
		app.logger.Printf("Error fetching todo: %v", err)

		if err.Error() == "todo not found" {
			http.Error(w, "Todo not found", http.StatusNotFound)
		} else {
			span.AddEvent("Error fetching todo")
			span.SetStatus(codes.Error, err.Error())
			http.Error(w, "Server error", http.StatusInternalServerError)
		}
		return
	}

	todoResponse := domain.TodoDetailDto{
		ID:          todo.GetID(),
		Title:       todo.GetTitle(),
		Description: todo.GetDescription(),
		Status:      todo.GetStatus(),
		CreatedAt:   todo.GetCreatedAt().Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   todo.GetUpdatedAt().Format("2006-01-02T15:04:05Z07:00"),
	}

	todoJson, err := json.Marshal(todoResponse)
	if err != nil {
		app.logger.Printf("Error marshaling todo response: %v", err)
		span.AddEvent("Error marshaling todo response")
		span.SetStatus(codes.Error, err.Error())
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	w.Write(todoJson)
}

func (app *application) updateTodo(w http.ResponseWriter, r *http.Request) {
	tracer := otel.Tracer("UpdateTodoHandler")
	ctx, span := tracer.Start(r.Context(), "UpdateTodoHandler")
	defer span.End()
	r = r.WithContext(ctx)

	loggedInUserID := r.Context().Value("userID")
	span.SetAttributes(attribute.String("user_id", loggedInUserID.(string)))

	if app.db == nil {
		app.logger.Printf("Database connection is nil")
		http.Error(w, "Server Error", http.StatusInternalServerError)
		return
	}

	idStr := r.URL.Path[len(todoIDRoute):]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var updateTodoDto domain.UpdateTodoDto
	err = json.NewDecoder(r.Body).Decode(&updateTodoDto)
	if err != nil {
		app.logger.Printf("Error decoding update todo request: %v", err)
		span.AddEvent("Error decoding update todo request")
		span.SetStatus(codes.Error, err.Error())
		http.Error(w, "Server Error", http.StatusInternalServerError)
		return
	}

	var todoRepo repositories.ITodoRepository = repositories.NewTodoRepository(app.db, app.rabbitConn)

	err = todoRepo.Update(id, updateTodoDto.Title, updateTodoDto.Description)
	if err != nil {
		app.logger.Printf("Error updating todo: %v", err)

		if err.Error() == "todo not found" {
			http.Error(w, "Todo not found", http.StatusNotFound)
		} else {
			span.AddEvent("Error updating todo")
			span.SetStatus(codes.Error, err.Error())
			http.Error(w, "Server error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusNoContent)
}

func (app *application) deleteTodo(w http.ResponseWriter, r *http.Request) {
	tracer := otel.Tracer("DeleteTodoHandler")
	ctx, span := tracer.Start(r.Context(), "DeleteTodoHandler")
	defer span.End()
	r = r.WithContext(ctx)

	loggedInUserID := r.Context().Value("userID")
	span.SetAttributes(attribute.String("user_id", loggedInUserID.(string)))

	if app.db == nil {
		app.logger.Printf("Database connection is nil")
		http.Error(w, "Server Error", http.StatusInternalServerError)
		return
	}

	idStr := r.URL.Path[len(todoIDRoute):]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		app.logger.Printf("Error parsing ID: %v", err)
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var todoRepo repositories.ITodoRepository = repositories.NewTodoRepository(app.db, app.rabbitConn)

	err = todoRepo.Delete(id)
	if err != nil {
		app.logger.Printf("Error deleting todo: %v", err)

		if err.Error() == "todo not found" {
			http.Error(w, "Todo not found", http.StatusNotFound)
		} else {
			span.AddEvent("Error deleting todo")
			span.SetStatus(codes.Error, err.Error())
			http.Error(w, "Server error", http.StatusInternalServerError)
		}

		return
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusNoContent)
}
