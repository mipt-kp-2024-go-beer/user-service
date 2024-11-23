package users

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"sync"

	"github.com/mipt-kp-2024-go-beer/user-service/internal/server"
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
	server.InitPublic(h.public)
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

const TokenLen int = 64

func GenerateSecureToken(length int) string {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}

// Login handler for user login
func (h *Handler) loginHandler(w http.ResponseWriter, r *http.Request) {
	var creds struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}
	println("we got here 1")

	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	println("we got here")

	// Check credentials (mock implementation)
	ctx := context.Background()
	checked, err := h.service.CheckUser(ctx, User{"", creds.Login, creds.Password, 0})
	if err != nil || !checked {
		http.Error(w, "There is no user with such credentials", http.StatusBadRequest)
		return
	}

	token, err := h.service.GetUniqueToken(ctx)

	if err != nil {
		http.Error(w, "Error in generating unique token", http.StatusNotAcceptable)
		return
	}

	w.WriteHeader(http.StatusFound)

	json.NewEncoder(w).Encode(token)
	return

	http.Error(w, "Unauthorized", http.StatusUnauthorized)
}

// Create user handler
func createUserHandler(w http.ResponseWriter, r *http.Request) {
	var newUser HandleUser
	if err := json.NewDecoder(r.Body).Decode(&newUser); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	usersMutex.Lock()
	defer usersMutex.Unlock()

	newUser.ID = string(nextID)
	nextID++

	users[newUser.ID] = newUser
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"id": newUser.ID})
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

// Main function to start the server
func (h *Handler) InitPublic(m *http.ServeMux) {
	m.HandleFunc("/user/login", h.loginHandler)
	m.HandleFunc("/user/create", createUserHandler)
	m.HandleFunc("/user/delete", deleteUserHandler)
	m.HandleFunc("/user/info", getUserInfoHandler)
	m.HandleFunc("/user/edit", editUserHandler)
}

func (h *Handler) postProducts(w http.ResponseWriter, r *http.Request) {

}
