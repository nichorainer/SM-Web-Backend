package main

import (
	// "context"
	"log/slog"
	"os"
	// "log"

	// "github.com/jackc/pgx/v5"
	"github.com/yourorg/backend-go/internal/env"
	"github.com/yourorg/backend-go/internal/config"
	// appmiddleware "github.com/yourorg/backend-go/internal/middleware"
)

func main() {
	
	cfg := configStruct{
		addr: ":8080",
		db: dbConfig{
			dsn: env.GetString("GOOSE_DBSTRING", "host=localhost user=postgres password=admin dbname=InventoryDB sslmode=disable"),
		},
	}

	// logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// database
	// conn, err := pgx.Connect(ctx, cfg.db.dsn)
	// if err != nil {
	// 	panic(err)
	// }
	// defer conn.Close(ctx)

	// logger.Info("Database connected", "dsn", cfg.db.dsn)

	// init database pool via config.InitDB()
	config.InitDB()
	defer config.GetDB().Close()
	// log init db
	logger.Info("Database pool initialized", "dsn", cfg.db.dsn)

	// application pakai pool dari config.GetDB()
	api := application{
		config: cfg,
		db:     config.GetDB(),
	}
	
	// // Initiate JWT Secret
	// appmiddleware.InitJWT(env.GetString("JWT_SECRET", "supersecret"))

	if err := api.run(api.mount()); err != nil {
		slog.Error("Server failed to start", "error", err)
		os.Exit(1)
	}
}
