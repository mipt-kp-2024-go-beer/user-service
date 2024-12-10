package users

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
)

// Handler is responsible for handling HTTP requests and routing them to the appropriate service.
type Handler struct {
	service Service        // Service interface to perform business logic
	public  *http.ServeMux // ServeMux for public routes
	private *http.ServeMux // ServeMux for private routes
}

// Handler constructor
func NewHandler(service Service, public *http.ServeMux, private *http.ServeMux) *Handler {
	return &Handler{
		service: service,
		public:  public,
		private: private,
	}
}

// Register sets up the public and private routes for the handler.
func (h *Handler) Register() {
	h.InitPublic(h.public)
	h.InitPrivate(h.private)
}

// loginHandler handles user login requests by validating credentials and generating a token.
// @param w http.ResponseWriter for returning the response to the client.
// @param r *http.Request containing the user's login credentials in the request body.
func (h *Handler) loginHandler(w http.ResponseWriter, r *http.Request) {
	// user credentials to be checked
	var creds struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Taking credentials, chcecking existance of user and generating access and refresh token
	ctx := context.Background()
	token, err := h.service.CreateToken(ctx, creds.Login, creds.Password)
	if err != nil {
		http.Error(w, "Error getting token", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusFound)
	json.NewEncoder(w).Encode(token)
}

// createUserHandler handles requests to create a new user.
// @param w http.ResponseWriter for returning the response to the client.
// @param r *http.Request containing the user's credentials in the request body.
func (h *Handler) createUserHandler(w http.ResponseWriter, r *http.Request) {
	// user credentials to be checked
	var creds struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Appending new user to the storage also checking existance of user
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

// deleteUserHandler handles requests to delete a user.
// @param w http.ResponseWriter for returning the response to the client.
// @param r *http.Request containing the access token in the request body.
func (h *Handler) deleteUserHandler(w http.ResponseWriter, r *http.Request) {
	// token of corresponding user
	var token struct {
		Access string `json:"token"`
	}

	if err := json.NewDecoder(r.Body).Decode(&token); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	ID, err := h.service.GetIDByToken(ctx, token.Access)

	// if token is not correct or token is expired quit
	if err != nil {
		http.Error(w, "Token incorrect", http.StatusBadRequest)
	}

	err = h.service.DeleteUser(ctx, ID)

	if err != nil {
		http.Error(w, "There is no user", http.StatusBadRequest)
	}

	w.WriteHeader(http.StatusOK)
}

// getting ID of user, private api
// @param w http.ResponseWriter for returning the response to the client.
// @param r *http.Request containing the access token in the request body.
func (h *Handler) getID(w http.ResponseWriter, r *http.Request) {
	// token of corresponding user
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

// getting Permissions of user, private api
// @param w http.ResponseWriter for returning the response to the client.
// @param r *http.Request containing the access token in the request body.
func (h *Handler) getPermissions(w http.ResponseWriter, r *http.Request) {
	// token of corresponding user
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

// edit user information by user with corresponding permissions
// @param w http.ResponseWriter for returning the response to the client.
// @param r *http.Request containing the access token in the request body.
func (h *Handler) editUserHandler(w http.ResponseWriter, r *http.Request) {
	// token is token of user with corresponding permissions, admin
	// id - id of user to be edited
	// newLogin, newPassword - newData for editing
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

	// user editing with checking token and user to be edited
	ctx := context.Background()
	_, err := h.service.EditUser(ctx, editor.Access, User{Login: editor.Login, Password: editor.Password, ID: editor.ID, Permissions: 0})
	if err != nil {
		http.Error(w, "Error editing tiken", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// give permissions to user by user with corresponding permissions
// @param w http.ResponseWriter for returning the response to the client.
// @param r *http.Request containing the access token in the request body.
func (h *Handler) givePermissionHandler(w http.ResponseWriter, r *http.Request) {
	// token is token of user with corresponding permissions, admin
	// id - id of user to be edited
	// Permission - permissions to be edited
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

// refresh access token for user
// @param w http.ResponseWriter for returning the response to the client.
// @param r *http.Request containing the access token in the request body.
func (h *Handler) RefreshHandler(w http.ResponseWriter, r *http.Request) {
	var tokens struct {
		Acess   string `json:"access"`
		Refresh string `json:"refresh"`
	}

	if err := json.NewDecoder(r.Body).Decode(&tokens); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	token, err := h.service.RefreshToken(ctx, tokens.Acess, tokens.Refresh)
	if err != nil {
		http.Error(w, "Error getting token", http.StatusBadRequest)
	}

	w.WriteHeader(http.StatusFound)
	json.NewEncoder(w).Encode(token)
}

// Initing public routing
// @param m http.ServeMux public mux
func (h *Handler) InitPublic(m *http.ServeMux) {
	m.HandleFunc("/user/login", h.loginHandler)
	m.HandleFunc("/user/create", h.createUserHandler)
	m.HandleFunc("/user/delete", h.deleteUserHandler)
	m.HandleFunc("/user/edit", h.editUserHandler)
	m.HandleFunc("/user/give", h.givePermissionHandler)
	m.HandleFunc("/user/refresh", h.RefreshHandler)
}

// Initing public routing
// @param m http.ServeMux private mux
func (h *Handler) InitPrivate(m *http.ServeMux) {
	m.HandleFunc("/user/id", h.getID)
	m.HandleFunc("/user/permissions", h.getPermissions)
}
