package main

import (
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5"

	repo "github.com/yourorg/backend-go/internal/adapters/postgresql/sqlc"
	"github.com/yourorg/backend-go/internal/products"
	"github.com/yourorg/backend-go/internal/orders"
)

// mount
func (app *application) mount() http.Handler {
	r := chi.NewRouter()

	// middleware
	r.Use(middleware.RequestID)	// important for rate limiting
	r.Use(middleware.RealIP) 	// import for rate limiting, analytics and tracing
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)	// recover from crashes

	// set a timeout value on the request context (ctx),
	// that will signal through ctx.Done() that the
	// request has timed out and further processing should be stopped.
	r.Use(middleware.Timeout(60 * time.Second)) // 60 seconds

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("all good"))
	})

	// products routes
	productService := products.NewService(repo.New(app.db))
	productHandler := products.NewHandler(productService)
	r.Get("/products", productHandler.ListProducts)

	// orders routes
	orderService := orders.NewService(repo.New(app.db), app.db)
	ordersHandler := orders.NewHandler(orderService)
	r.Post("/orders", ordersHandler.PlaceOrder)
	r.Post("/orders", ordersHandler.GetOrders)

	return r 
}

// run
func (app *application) run(h http.Handler) error {
	srv := &http.Server{
		Addr: 			app.config.addr,
		Handler: 		h,
		WriteTimeout: 	time.Second * 30,
		ReadTimeout: 	time.Second * 10,
		IdleTimeout: 	time.Minute,
	}

	log.Printf("Server has started at addr %s", app.config.addr)

	return srv.ListenAndServe()
}


type application struct {
	config config
	
	// db driver
	db *pgx.Conn
}

type config struct {
	addr string
	db dbConfig
}

type dbConfig struct {
	dsn string
}