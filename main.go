package main

var (
// en         env.Provider      = env.NewEnv()
// db         database.Provider = database.NewPG()
// mainRouter http.Router = http.NewMuxRouter()
)

func main() {
	initPublic()
	//router := mux.NewRouter()

	//userRepo := repository.NewUserRepository()
	//userService := services.NewUserService(userRepo)

	// Open APIs
	//router.HandleFunc("/api/user/login", handlers.UserLogin(userService)).Methods("POST")
	//router.HandleFunc("/api/user/create", middlewares.AuthMiddleware, handlers.UserCreate(userService)).Methods("POST")
	//router.HandleFunc("/api/user/delete/{id}", middlewares.AuthMiddleware, handlers.UserDelete(userService)).Methods("DELETE")

	// Public user info endpoint
	//router.HandleFunc("/api/user/info/{id}", middlewares.AuthMiddleware, handlers.UserInfo(userService)).Methods("GET")

	//log.Fatal(http.ListenAndServe(":8080", router))
}
