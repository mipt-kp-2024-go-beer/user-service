package users

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

type Handler struct {
	service Service
	public  *http.ServeMux
	private *http.ServeMux
}

func NewHandler(service Service, public *http.ServeMux, private *http.ServeMux) *Handler {
	return &Handler{
		service: service,
		public:  public,
		private: private,
	}
}

func (h *Handler) Register() {
	h.InitPublic(h.public)
	h.InitPrivate(h.private)
	//h.public.Group(func(r chi.Router) {
	//	r.Get("/api/v1/products", h.getProducts)
	//	r.Post("/api/v1/products", h.postProducts)
	//})
}

func (h *Handler) getProducts(w http.ResponseWriter, r *http.Request) {
	// validate r
	//data, err := h.service.Products(r.Context())
	//if err != nil {
	//	http.Error(w, err.Error(), http.StatusInternalServerError)
	//	return
	//}

	//fmt.Fprintf(w, "", data)
}

// User structure to hold user data
type HandleUser struct {
	ID          string `json:"id"`
	Login       string `json:"login"`
	Password    string `json:"password"`
	Permissions int8   `json:"permissions"`
}

// In-memory storage for demonstration purposes
var (
	usersMutex sync.Mutex
	users      = make(map[string]HandleUser) // user mock
	nextID     = 1
)

// Login handler for user login
func (h *Handler) loginHandler(w http.ResponseWriter, r *http.Request) {
	var creds struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Check credentials (mock implementation)
	ctx := context.Background()
	token, err := h.service.CreateToken(ctx, creds.Login, creds.Password)
	if err != nil {
		http.Error(w, "Error getting token", http.StatusBadRequest)
	}

	w.WriteHeader(http.StatusFound)
	json.NewEncoder(w).Encode(token)
}

// Create user handler
func (h *Handler) createUserHandler(w http.ResponseWriter, r *http.Request) {
	var creds struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	id, err := h.service.NewUser(ctx, User{"", creds.Login, creds.Password, 0})

	if err != nil {
		w.WriteHeader(http.StatusConflict)
		http.Error(w, "User exists", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"id": id})
}

// Delete user handler
func deleteUserHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	usersMutex.Lock()
	defer usersMutex.Unlock()

	// Check for existence and delete
	if _, exists := users[id]; exists {
		delete(users, id)
		w.WriteHeader(http.StatusNoContent)
		return
	}

	http.Error(w, "User not found", http.StatusNotFound)
}

// Get user info handler
func getUserInfoHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	usersMutex.Lock()
	defer usersMutex.Unlock()

	if user, exists := users[id]; exists {
		json.NewEncoder(w).Encode(user)
		return
	}

	http.Error(w, "User not found", http.StatusNotFound)
}

// Edit user handler
func editUserHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	var updatedUser HandleUser
	if err := json.NewDecoder(r.Body).Decode(&updatedUser); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	usersMutex.Lock()
	defer usersMutex.Unlock()

	if user, exists := users[id]; exists {
		user.Login = updatedUser.Login
		user.Password = updatedUser.Password
		user.Permissions = updatedUser.Permissions
		users[id] = user
		json.NewEncoder(w).Encode(user)
		return
	}

	http.Error(w, "User not found", http.StatusNotFound)
}

func (h *Handler) getID(w http.ResponseWriter, r *http.Request) {
	var token struct {
		Access string `json:"token"`
	}

	if err := json.NewDecoder(r.Body).Decode(&token); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	ID, err := h.service.GetIDByToken(ctx, token.Access)

	if err != nil {
		fmt.Errorf("%w", err)
	}

	if err != nil {
		http.Error(w, "Token incorrect", http.StatusBadRequest)
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"id": ID})
}

// Main function to start the server
func (h *Handler) InitPublic(m *http.ServeMux) {
	m.HandleFunc("/user/login", h.loginHandler)
	m.HandleFunc("/user/create", h.createUserHandler)
	// not impemented yet
	m.HandleFunc("/user/delete", deleteUserHandler)
	m.HandleFunc("/user/info", getUserInfoHandler)
	m.HandleFunc("/user/edit", editUserHandler)
}

func (h *Handler) InitPrivate(m *http.ServeMux) {
	// not impemented yet
	m.HandleFunc("/user/id", h.getID)
	m.HandleFunc("/user/permissions", h.createUserHandler)
}
