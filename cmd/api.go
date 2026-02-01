package main

import (
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5/pgxpool"
	// "github.com/jackc/pgx/v5"

	repo "github.com/yourorg/backend-go/internal/adapters/postgresql/sqlc"
	"github.com/yourorg/backend-go/internal/handlers"
	/* appmiddleware "github.com/yourorg/backend-go/internal/middleware" */
)

// Mount Server
func (app *application) mount() http.Handler {
	r := chi.NewRouter()

	server := handlers.Server{
		Repo: repo.New(app.db),
	}

	// --- CORS middleware ---
	// Allow FE (React dev server) to call BE

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173"}, // FE dev server
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
	}))

	// --- End CORS ---

	// Global Middleware
	r.Use(chimiddleware.RequestID) // important for rate limiting
	r.Use(chimiddleware.RealIP)    // import for rate limiting, analytics and tracing
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)       // recover from crashes
	r.Use(chimiddleware.RedirectSlashes) // redirect slashes to no slash URL
	r.Use(chimiddleware.Timeout(60 * time.Second))

	// Health Check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("all good"))
	})

	// Auth Routes (Public)
	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", server.CreateUser)
		r.Post("/login", server.Login)
	})

	// // Products Routes
	r.Route("/products", func(r chi.Router) {
		r.Get("/", server.ListProducts)
		r.Post("/", server.CreateProduct)
		r.Get("/{id}", server.GetProductByID)
	})

	// Orders Routes
	r.Get("/orders", server.ListOrders)
	r.Post("/orders", server.CreateOrder)

	// Users
	r.Get("/users", server.ListUsers)
	r.Get("/users/{user_id}", server.GetUserByID)
	r.Put("/users/{id}", handlers.UpdateUser)
	r.Put("/users/me", handlers.UpdateUser)

	// Users (protected endpoints)
	// r.Group(func(r chi.Router) {
	// 	r.Use(appmiddleware.JWTMiddleware)
	// 	r.Post("/users/{id}", handlers.UpdateUser)
	// 	r.Post("/users/me", handlers.UpdateUser)
	// })

	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("MethodNotAllowed: %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	})
	// Log running routes
	chi.Walk(r, func(method string, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		log.Printf("[ROUTE] %s %s", method, route)
		return nil
	})

	return r
}

// run
func (app *application) run(h http.Handler) error {
	srv := &http.Server{
		Addr:         app.config.addr,
		Handler:      h,
		WriteTimeout: time.Second * 30,
		ReadTimeout:  time.Second * 10,
		IdleTimeout:  time.Minute,
	}

	log.Printf("Server has started at addr %s", app.config.addr)

	return srv.ListenAndServe()
}

type application struct {
	config configStruct

	// db driver
	// db *pgx.Conn

	// db driver using pool
	db *pgxpool.Pool
}

type configStruct struct {
	addr string
	db   dbConfig
}

type dbConfig struct {
	dsn string
}
