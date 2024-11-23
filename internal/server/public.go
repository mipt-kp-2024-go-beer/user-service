package server

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"

	"sync"
)

// User structure to hold user data
type User struct {
	ID          string `json:"id"`
	Login       string `json:"login"`
	Password    string `json:"password"`
	Permissions int8   `json:"permissions"`
}

// In-memory storage for demonstration purposes
var (
	usersMutex sync.Mutex
	users      = make(map[string]User) // user mock
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
func loginHandler(w http.ResponseWriter, r *http.Request) {
	var creds struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Check credentials (mock implementation)
	for _, user := range users {
		if user.Login == creds.Login && user.Password == creds.Password {
			token := GenerateSecureToken(TokenLen)
			json.NewEncoder(w).Encode(map[string]string{"token": token})
			return
		}
	}

	http.Error(w, "Unauthorized", http.StatusUnauthorized)
}

// Create user handler
func createUserHandler(w http.ResponseWriter, r *http.Request) {
	var newUser User
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
	var updatedUser User
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
func InitPublic(m *http.ServeMux) {
	m.HandleFunc("/user/login", loginHandler)
	m.HandleFunc("/user/create", createUserHandler)
	m.HandleFunc("/user/delete", deleteUserHandler)
	m.HandleFunc("/user/info", getUserInfoHandler)
	m.HandleFunc("/user/edit", editUserHandler)
}
