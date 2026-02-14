package main

import (
	"log/slog"
	"os"

	"github.com/nichorainer/backend-go/internal/env"
	"github.com/nichorainer/backend-go/internal/config"
)

func main() {
	cfg := configStruct{
		addr: ":8080",
		db: dbConfig{
			dsn: env.GetString("GOOSE_DBSTRING",
				"host=localhost user=postgres password=admin dbname=InventoryDB sslmode=disable"),
		},
	}

	// logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

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

	// run server
	if err := api.run(api.mount()); err != nil {
		slog.Error("Server failed to start", "error", err)
		os.Exit(1)
	}
}
