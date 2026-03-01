package main

import (
	"backend-go/config"
	sqliteeev "backend-go/internal/db/sqlite/generated"
	"backend-go/internal/handlers"
	"backend-go/internal/httpx"
	middleware2 "backend-go/internal/middleware"
	"backend-go/internal/s3"
	"backend-go/internal/services"
	"backend-go/internal/validation"
	"context"
	"database/sql"
	"log"
	"net/http"
	"strconv"

	_ "embed"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "modernc.org/sqlite"
)

func main() {
	appConfig := config.NewAppConfig()
	ctx := context.Background()
	log.Println("Opening SQLite DB at:", appConfig.DB.ConnectionString())
	db, err := sql.Open("sqlite", appConfig.DB.ConnectionString())
	if err != nil {
		log.Fatal(err)
	}

	if err := db.Ping(); err != nil {
		log.Fatal("failed to ping db:", err)
	}

	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(db)

	validation.Init()
	queries := sqliteeev.New(db)
	log.Println("Database connection established")

	r := chi.NewRouter()
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	//r.Use(middleware.RedirectSlashes)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Heartbeat("/alive"))

	s3Manager, err := s3.NewManager(ctx, s3.ManagerConfig{
		AccessKey:   appConfig.S3.AccessKey,
		SecretKey:   appConfig.S3.SecretKey,
		Region:      appConfig.S3.Region,
		Bucket:      appConfig.S3.Bucket,
		EndpointURL: appConfig.S3.EndpointURL,
	})

	shareableService := services.NewShareableService(queries, s3Manager)
	authService := services.NewAuthService(appConfig.OAuth, appConfig.Jwt, queries)
	userService := services.NewUserService(queries)

	shareableHandler := handlers.NewShareableHandler(shareableService)
	authHandler := handlers.NewAuthHandler(appConfig, authService, userService)
	meHandler := handlers.NewMeHandler(authService, userService)

	r.Group(func(r chi.Router) {
		r.Get("/share/{code}", shareableHandler.GetShareable)

		r.Get("/auth/google", authHandler.GoogleLogin)
		r.With(
			httpx.ValidateBody[handlers.GoogleCallbackRequest],
		).Post("/auth/google/callback", authHandler.GoogleCallback)
	})

	r.Group(func(r chi.Router) {
		r.Use(middleware2.AutoRefreshMiddleware(authService))

		r.Get("/me", meHandler.GetMe)
		r.Post("/share", httpx.ValidateBody[handlers.CreateShareableRequest](
			http.HandlerFunc(shareableHandler.CreateShareable),
		).ServeHTTP)
	})

	server := &http.Server{
		Addr:    ":" + strconv.Itoa(appConfig.Port),
		Handler: r,
	}

	log.Printf("Server started at port %d", appConfig.Port)

	err = server.ListenAndServe()
	if err != nil {
		log.Fatal("Server failed", "error", err)
	}
}
