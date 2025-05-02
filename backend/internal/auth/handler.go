package auth

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/shekshuev/codeed/backend/internal/users"
)

type AuthHandler interface {
	GetTelegramCode(w http.ResponseWriter, r *http.Request)
	CheckTelegramCode(w http.ResponseWriter, r *http.Request)
}

type AuthHandlerImpl struct {
	userService users.UserService
}

func NewUserHandler(service users.UserService) *AuthHandlerImpl {
	return &AuthHandlerImpl{userService: service}
}

func (h *AuthHandlerImpl) RegisterRoutes(r chi.Router) {
	// r.Route("/users", func(r chi.Router) {
	// 	r.Post("/", h.createUser)
	// 	r.Get("/{id}", h.getUserByID)
	// })
}
