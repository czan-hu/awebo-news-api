package database

import (
	"fmt"

	gormpg "gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"awebo/app/infrastructure/config"
)

func dsn(cfg *config.Config) string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName,
	)
}

func Connect(cfg *config.Config) (*gorm.DB, error) {
	db, err := gorm.Open(gormpg.Open(dsn(cfg)), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, fmt.Errorf("connect to database: %w", err)
	}
	return db, nil
}

// Migrate runs every pending SQL migration found in migrationsDir using the
// same connection pool GORM already opened.
func Migrate(db *gorm.DB, migrationsDir string) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("get underlying sql.DB: %w", err)
	}
	return RunMigrations(sqlDB, migrationsDir)
}
