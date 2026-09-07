package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"awebo/app/handlers"
	"awebo/app/infrastructure/config"
	"awebo/app/infrastructure/database"
	"awebo/app/infrastructure/server"
	"awebo/app/repositories"
	"awebo/app/services"
)

const migrationsDir = "app/infrastructure/database/migrations"

func main() {
	_ = godotenv.Load()

	cfg := config.Load()

	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	if err := database.Migrate(db, migrationsDir); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	mailer := services.NewMailer(cfg)
	uploadService := services.NewUploadService(cfg)

	authRepo := repositories.NewAuthRepository(db)
	userRepo := repositories.NewUserRepository(db)
	articleRepo := repositories.NewArticleRepository(db)
	categoryRepo := repositories.NewCategoryRepository(db)
	contactRepo := repositories.NewContactRepository(db)
	newsletterRepo := repositories.NewNewsletterRepository(db)
	forumRepo := repositories.NewForumRepository(db)
	creatorRequestRepo := repositories.NewCreatorRequestRepository(db)

	handler := handlers.NewHandler(handlers.Services{
		Auth:           services.NewAuthService(authRepo, cfg, mailer),
		User:           services.NewUserService(userRepo),
		Article:        services.NewArticleService(articleRepo, categoryRepo),
		Category:       services.NewCategoryService(categoryRepo),
		Contact:        services.NewContactService(contactRepo),
		Newsletter:     services.NewNewsletterService(newsletterRepo),
		Forum:          services.NewForumService(forumRepo),
		Search:         services.NewSearchService(articleRepo, forumRepo),
		CreatorRequest: services.NewCreatorRequestService(creatorRequestRepo, userRepo),
		Avatar:         services.NewAvatarService(userRepo, uploadService),
		Upload:         uploadService,
	}, cfg)

	srv := new(infrastructure.Server)

	go func() {
		log.Printf("awebo API listening on :%s%s", cfg.ServicePort, cfg.APIPrefix)
		if err := srv.Run(cfg.ServicePort, handler.InitRoutes()); err != nil && err != http.ErrServerClosed {
			log.Fatalf("failed to start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("failed to shutdown server gracefully: %v", err)
	}
	log.Println("server stopped")
}
