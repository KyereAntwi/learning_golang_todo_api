package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"todoapi.com/m/src/domain"
	"todoapi.com/m/src/repositories"
)

// @Summary Sign Up
// @Description Sign up a new user
// @ID sign-up
// @Accept json
// @Produce json
// @Param signUpDto body domain.SignUpDto true "Sign Up DTO"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {string} string "Bad Request"
// @Failure 409 {string} string "Username already taken"
// @Failure 500 {string} string "Server Error"
// @Router /signup [post]
func (app *application) signUpRouteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if app.db == nil {
		app.logger.Printf("Database connection is nil")
		http.Error(w, "Server Error", http.StatusInternalServerError)
		return
	}

	var signUpDto domain.SignUpDto
	err := json.NewDecoder(r.Body).Decode(&signUpDto)
	if err != nil {
		app.logger.Printf("Error decoding request body: %v", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	var userRepo repositories.IUserRepository = repositories.NewUserRepository(app.db)

	isTaken, err := userRepo.IsUsernameTaken(signUpDto.Username)
	if err != nil {
		app.logger.Printf("Error checking if username is taken: %v", err)
		http.Error(w, "Server Error", http.StatusInternalServerError)
		return
	}

	if isTaken {
		http.Error(w, "Username already taken", http.StatusConflict)
		return
	}

	id, err := userRepo.Create(signUpDto.Username, signUpDto.Password, signUpDto.Email, signUpDto.PrimaryPhone)
	if err != nil {
		app.logger.Printf("Error creating user: %v", err)
		http.Error(w, "Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Location", "/users/"+id.String())
	w.WriteHeader(http.StatusCreated)

	response := map[string]interface{}{
		"id": id,
	}

	responseJson, err := json.Marshal(response)
	if err != nil {
		app.logger.Printf("Error marshaling response: %v", err)
		http.Error(w, "Server Error", http.StatusInternalServerError)
		return
	}

	w.Write(responseJson)
}

// @Summary Sign In
// @Description Sign in a user
// @ID sign-in
// @Accept json
// @Produce json
// @Param loginDto body domain.LoginDto true "Login DTO"
// @Success 200 {object} domain.LoginResponse
// @Failure 400 {string} string "Bad Request"
// @Failure 401 {string} string "Invalid username or password"
// @Failure 500 {string} string "Server Error"
// @Router /signin [post]
func (app *application) signInRouteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if app.db == nil {
		app.logger.Printf("Database connection is nil")
		http.Error(w, "Server Error", http.StatusInternalServerError)
		return
	}

	var signInDto domain.LoginDto
	err := json.NewDecoder(r.Body).Decode(&signInDto)
	if err != nil {
		app.logger.Printf("Error decoding request body: %v", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	var userRepo repositories.IUserRepository = repositories.NewUserRepository(app.db)

	userId, err := userRepo.Authenticate(signInDto.Username, signInDto.Password)

	if err != nil {
		app.logger.Printf("Error authenticating user: %v", err)
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	accessToken, err := app.jwtManager.GenerateAccessToken(userId)

	if err != nil {
		app.logger.Printf("Error generating access token: %v", err)
		http.Error(w, "Server Error", http.StatusInternalServerError)
		return
	}

	refreshToken, expiresAt, err := app.jwtManager.GenerateRefreshToken(userId)

	if err != nil {
		app.logger.Printf("Error generating refresh token: %v", err)
		http.Error(w, "Server Error", http.StatusInternalServerError)
		return
	}

	hashedRefreshToken, err := app.jwtManager.HashRefreshToken(refreshToken)
	if err != nil {
		app.logger.Printf("Error hashing refresh token: %v", err)
		http.Error(w, "Server Error", http.StatusInternalServerError)
		return
	}

	// Store the refresh token in the database
	refreshTokenDomain := domain.RefreshToken{
		UserId:      uuid.MustParse(userId),
		HashedToken: hashedRefreshToken,
		ExpiresAt:   expiresAt,
		CreatedAt:   time.Now(),
	}
	err = userRepo.StoreRefreshToken(refreshTokenDomain)
	if err != nil {
		app.logger.Printf("Error storing refresh token: %v", err)
		http.Error(w, "Server Error", http.StatusInternalServerError)
		return
	}

	response := domain.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	responseJson, err := json.Marshal(response)
	if err != nil {
		app.logger.Printf("Error marshaling response: %v", err)
		http.Error(w, "Server Error", http.StatusInternalServerError)
		return
	}

	w.Write(responseJson)
}

// @Summary Refresh Token
// @Description Refresh a user's access token
// @ID refresh-token
// @Accept json
// @Produce json
// @Param refreshTokenDto body domain.RefreshTokenDto true "Refresh Token DTO"
// @Success 200 {object} domain.LoginResponse
// @Failure 400 {string} string "Bad Request"
// @Failure 401 {string} string "Invalid refresh token"
// @Failure 500 {string} string "Server Error"
// @Router /refresh-token [post]
func (app *application) refreshTokenRouteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if app.db == nil {
		app.logger.Printf("Database connection is nil")
		http.Error(w, "Server Error", http.StatusInternalServerError)
		return
	}

	var requestDto domain.RefreshTokenDto
	err := json.NewDecoder(r.Body).Decode(&requestDto)
	if err != nil {
		app.logger.Printf("Error decoding request body: %v", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	if requestDto.RefreshToken == "" {
		http.Error(w, "Refresh token is required", http.StatusBadRequest)
		return
	}

	var userRepo repositories.IUserRepository = repositories.NewUserRepository(app.db)

	userId, err := app.jwtManager.GetUserIDFromToken(requestDto.RefreshToken)
	if err != nil {
		app.logger.Printf("Error validating refresh token: %v", err)
		http.Error(w, "Invalid refresh token", http.StatusUnauthorized)
		return
	}

	currentRefreshToken, err := userRepo.GetRefreshTokenByUserID(uuid.MustParse(userId))
	if err != nil {
		app.logger.Printf("Error retrieving refresh token from database: %v", err)
		http.Error(w, "Invalid refresh token", http.StatusUnauthorized)
		return
	}

	// check if the refresh token is expired
	if time.Now().After(currentRefreshToken.ExpiresAt) {
		app.logger.Printf("Refresh token expired for user %s", userId)
		http.Error(w, "Refresh token expired", http.StatusUnauthorized)
		return
	}

	err = app.jwtManager.CompareRefreshTokens(requestDto.RefreshToken, currentRefreshToken.HashedToken)
	if err != nil {
		app.logger.Printf("Error comparing refresh tokens: %v", err)
		http.Error(w, "Invalid refresh token", http.StatusUnauthorized)
		return
	}

	newAccessToken, err := app.jwtManager.GenerateAccessToken(userId)
	if err != nil {
		app.logger.Printf("Error generating access token: %v", err)
		http.Error(w, "Server Error", http.StatusInternalServerError)
		return
	}

	newRefreshToken, expiresAt, err := app.jwtManager.GenerateRefreshToken(userId)
	if err != nil {
		app.logger.Printf("Error generating refresh token: %v", err)
		http.Error(w, "Server Error", http.StatusInternalServerError)
		return
	}

	hashedRefreshToken, err := app.jwtManager.HashRefreshToken(newRefreshToken)
	if err != nil {
		app.logger.Printf("Error hashing refresh token: %v", err)
		http.Error(w, "Server Error", http.StatusInternalServerError)
		return
	}

	// remove the old refresh token from the database
	err = userRepo.DeleteRefreshToken(uuid.MustParse(userId))
	if err != nil {
		app.logger.Printf("Error deleting old refresh token from database: %v", err)
		http.Error(w, "Server Error", http.StatusInternalServerError)
		return
	}

	// Store the new refresh token in the database
	refreshTokenDomain := domain.RefreshToken{
		UserId:      uuid.MustParse(userId),
		HashedToken: hashedRefreshToken,
		ExpiresAt:   expiresAt,
	}

	err = userRepo.StoreRefreshToken(refreshTokenDomain)
	if err != nil {
		app.logger.Printf("Error storing refresh token in database: %v", err)
		http.Error(w, "Server Error", http.StatusInternalServerError)
		return
	}

	response := domain.LoginResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
		TokenType:    "Bearer",
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	responseJson, err := json.Marshal(response)
	if err != nil {
		app.logger.Printf("Error marshaling response: %v", err)
		http.Error(w, "Server Error", http.StatusInternalServerError)
		return
	}

	w.Write(responseJson)
}
