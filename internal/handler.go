package users

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
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
		return
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
func (h *Handler) deleteUserHandler(w http.ResponseWriter, r *http.Request) {
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
		http.Error(w, "Token incorrect", http.StatusBadRequest)
	}

	err = h.service.DeleteUser(ctx, ID)

	if err != nil {
		http.Error(w, "There is no user", http.StatusBadRequest)
	}

	w.WriteHeader(http.StatusOK)
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
		http.Error(w, "Token incorrect", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"id": ID})
}

func (h *Handler) getPermissions(w http.ResponseWriter, r *http.Request) {
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
		http.Error(w, "Token incorrect", http.StatusBadRequest)
		return
	}

	Info, infoerr := h.service.UserInfo(ctx, ID)

	if infoerr != nil {
		http.Error(w, "User incorrect", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"permissios": strconv.Itoa(int(Info.Permissions))})
}

func (h *Handler) editUserHandler(w http.ResponseWriter, r *http.Request) {
	var editor struct {
		Access   string `json:"token"`
		ID       string `json:"id"`
		Login    string `json:"newLogin"`
		Password string `json:"newPassword"`
	}

	if err := json.NewDecoder(r.Body).Decode(&editor); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	_, err := h.service.EditUser(ctx, editor.Access, User{Login: editor.Login, Password: editor.Password, ID: editor.ID, Permissions: 0})
	if err != nil {
		http.Error(w, "Error editing tiken", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) givePermissionHandler(w http.ResponseWriter, r *http.Request) {
	var editor struct {
		Access     string `json:"token"`
		ID         string `json:"id"`
		Permission uint   `json:"permission"`
	}

	if err := json.NewDecoder(r.Body).Decode(&editor); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	err := h.service.GivePermission(ctx, editor.Access, editor.ID, editor.Permission)

	if err != nil {
		http.Error(w, "Cannot give permissions", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// Login handler for user login
func (h *Handler) RefreshHandler(w http.ResponseWriter, r *http.Request) {
	var tokens struct {
		Acess   string `json:"access"`
		Refresh string `json:"refresh"`
	}

	if err := json.NewDecoder(r.Body).Decode(&tokens); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Check credentials (mock implementation)
	ctx := context.Background()
	token, err := h.service.RefreshToken(ctx, tokens.Acess, tokens.Refresh)
	if err != nil {
		http.Error(w, "Error getting token", http.StatusBadRequest)
	}

	w.WriteHeader(http.StatusFound)
	json.NewEncoder(w).Encode(token)
}

// Main function to start the server
func (h *Handler) InitPublic(m *http.ServeMux) {
	m.HandleFunc("/user/login", h.loginHandler)
	m.HandleFunc("/user/create", h.createUserHandler)
	m.HandleFunc("/user/delete", h.deleteUserHandler)
	m.HandleFunc("/user/edit", h.editUserHandler)
	m.HandleFunc("/user/give", h.givePermissionHandler)
	m.HandleFunc("/user/refresh", h.RefreshHandler)
}

func (h *Handler) InitPrivate(m *http.ServeMux) {
	m.HandleFunc("/user/id", h.getID)
	m.HandleFunc("/user/permissions", h.getPermissions)
}
