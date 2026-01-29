package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/yourorg/backend-go/internal/env"
	appmiddleware "github.com/yourorg/backend-go/internal/middleware"
)

func main() {
ctx := context.Background()

	cfg := config{
		addr: ":8080",
		db: dbConfig{
			dsn: env.GetString("GOOSE_DBSTRING", "host=localhost user=postgres password=admin dbname=InventoryDB sslmode=disable"),
		},
	}

	// logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// database
	conn, err := pgx.Connect(ctx, cfg.db.dsn)
	if err != nil {
		panic(err)
	}
	defer conn.Close(ctx)

	logger.Info("Database connected", "dsn", cfg.db.dsn)
	
	// Initiate JWT Secret
    appmiddleware.InitJWT(env.GetString("JWT_SECRET", "supersecret"))


	api := application{
		config: cfg,
		db: conn,
	}
	
	if err := api.run(api.mount()); err != nil {
		slog.Error("Server failed to start", "error", err)
		os.Exit(1)
	}
	
}
