package handler

import (
	"context"
	"encoding/json"
	"errors"
	sqlcdb "flight-booking/db/sqlc"
	"flight-booking/internal/domain"
	"flight-booking/internal/service"
	"net/http"
	"strings"
)

type RegisterRequest struct {
	Name  	 		string `json:"name" binding:"required"`
    Email    		string `json:"email" binding:"required,email"`
	Password 		string `json:"password" binding:"required,min=6"`
	PhoneNumber   	string `json:"phone_number" binding:"required"`
}

type RegisterResponse struct {
	ID       		string `json:"id"`
	Name     		string `json:"name"`
	Email    		string `json:"email"`
	PhoneNumber 	string `json:"phone_number"`
}
type AuthHandler struct {
    svc AuthService
}
type AuthService interface {
    RegisterUser(ctx context.Context, req service.RegisterUserParams) (sqlcdb.User, error)
    LoginUser(ctx context.Context, email, password string) (string, error)
}
func NewAuthHandler(svc AuthService) *AuthHandler {
    return &AuthHandler{
        svc: svc,
    }
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	AccessToken string `json:"access_token"`
	TokenType  string `json:"token_type"`
	ExpiresIn int64 `json:"expires_in"`
}


func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	// check http.Post 
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}


	// Normalizer email, name
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	req.Name = strings.TrimSpace(req.Name)
	req.PhoneNumber = strings.TrimSpace(req.PhoneNumber)
	if req.Email == "" || req.Name == "" || req.Password == "" || req.PhoneNumber == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}
	user, err := h.svc.RegisterUser(r.Context(), service.RegisterUserParams{
		Name:        req.Name,
		Email:       req.Email,
		Password:    req.Password,
		PhoneNumber: req.PhoneNumber,
	})
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrUserAlreadyExists):
			http.Error(w, "User already exists", http.StatusConflict)
		case errors.Is(err, domain.ErrInValidPassword):
			http.Error(w, "Password must be at least 6 characters long", http.StatusBadRequest)
		default:
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	resp := RegisterResponse{
		ID:          user.ID.String(),
		Name:        user.Name,
		Email:       user.Email,
		PhoneNumber: user.PhoneNumber,
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

