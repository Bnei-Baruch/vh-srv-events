package cmd

import (
	"log/slog"
	"os"

	"github.com/spf13/cobra"

	"gitlab.bbdev.team/vh/vh-srv-events/repo"
)

func init() {
	rootCmd.AddCommand(devMigrateCmd)
}

var devMigrateCmd = &cobra.Command{
	Use:   "dev-migrate",
	Short: "Run database migrations",
	Long:  "Run database migrations for the events service",
	Run:   devMigrateFn,
}

func devMigrateFn(cmd *cobra.Command, args []string) {
	slog.Info("Starting migration process")

	// Run database migrations
	slog.Info("Running database migrations")
	if err := repo.SyncDBStructInsertionAndMigrations(); err != nil {
		slog.Error("Database migration failed", slog.Any("error", err))
		os.Exit(1)
	}
	slog.Info("Database migrations completed successfully")

	slog.Info("Migration process completed successfully")
}
