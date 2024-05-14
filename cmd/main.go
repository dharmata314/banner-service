package main

import (
	errMsg "banner-serivce/internal/api/err"
	"banner-serivce/internal/auth/jwt"
	"banner-serivce/internal/config"
	"banner-serivce/internal/crud"
	"banner-serivce/internal/db/postgresql"
	bannerhandlers "banner-serivce/internal/handlers/banner_handlers"
	featurehandlers "banner-serivce/internal/handlers/feature_handlers"
	taghandlers "banner-serivce/internal/handlers/tag_handlers"
	userhandlers "banner-serivce/internal/handlers/user_handlers"
	"context"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log/slog"
	"net/http"
	"os"
)

const (
	LocalEnv = "local"
	DevEnv   = "dev"
	ProdEnv  = "prod"
)

func main() {
	cfg := config.MustLoad()
	fmt.Println(cfg)
	fmt.Println(cfg.Env)
	log := setupLogger(cfg.Env)
	log.Debug("debug messages are active")
	pg, err := connectToPostgres(cfg, log)
	if err != nil {
		log.Error("failed to create postgres db", errMsg.Err(err))
		os.Exit(1)
	}
	fmt.Println("connecting to postgres...")
	defer pg.Close()

	if pg == nil {
		log.Error("failed to connect to postgres")
		os.Exit(1)
	}
	if err := pg.Ping(context.Background()); err != nil {
		log.Error("failed to ping postgres db", errMsg.Err(err))
		os.Exit(1)
	} else {
		log.Info("postgres db connected successfully")
	}
	log.Info("application started", slog.String("env", cfg.Env))

	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	fr := crud.NewFeatureRepository(pg.Db, log)
	tr := crud.NewTagRepository(pg.Db, log)
	ur := crud.NewUserRepository(pg.Db, log)
	br := crud.NewBannerRepository(pg.Db, log)
	btr := crud.NewBannerTagRepository(pg.Db, log)
	jwtManager := jwt.NewJWTManager(cfg.JWT.Secret, log)

	router.Post("/users", userhandlers.New(log, ur))
	router.Post("/login", userhandlers.LoginFunc(log, ur, jwtManager))

	router.With(func(next http.Handler) http.Handler {
		return jwt.TokenAuthMiddleware(jwtManager, next)
	}).Get("/user_banner", bannerhandlers.NewGetBannerHandler(log, br))

	router.With(func(next http.Handler) http.Handler {
		return jwt.TokenAuthMiddleware(jwtManager, next)
	}).Post("/tags", taghandlers.New(log, tr))

	router.With(func(next http.Handler) http.Handler {
		return jwt.TokenAuthMiddleware(jwtManager, next)
	}).Post("/banners", bannerhandlers.New(log, br, btr))

	router.With(func(next http.Handler) http.Handler {
		return jwt.TokenAuthMiddleware(jwtManager, next)
	}).Post("/features", featurehandlers.New(log, fr))

	router.With(func(next http.Handler) http.Handler {
		return jwt.TokenAuthMiddleware(jwtManager, next)
	}).Delete("/banner/{id}", bannerhandlers.NewDeleteBannerHandler(log, br))

	router.With(func(next http.Handler) http.Handler {
		return jwt.TokenAuthMiddleware(jwtManager, next)
	}).Get("/banner", bannerhandlers.NewGetBannersHandler(br, log))

	router.With(func(next http.Handler) http.Handler {
		return jwt.TokenAuthMiddleware(jwtManager, next)
	}).Patch("/banner/{id}", bannerhandlers.NewUpdateBannerHandler(br, log))

	log.Info("starting server", slog.String("addr", cfg.HTTPServer.Addr))
	server := &http.Server{
		Addr:              cfg.HTTPServer.Addr,
		Handler:           router,
		ReadHeaderTimeout: cfg.HTTPServer.Timeout,
		WriteTimeout:      cfg.HTTPServer.Timeout,
		IdleTimeout:       cfg.HTTPServer.IdleTimeout,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Error("failed to start server", errMsg.Err(err))
	}

}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger
	switch env {
	case LocalEnv:
		log = slog.New(slog.NewTextHandler(os.Stdout,
			&slog.HandlerOptions{Level: slog.LevelDebug}))
	case DevEnv:
		log = slog.New(slog.NewJSONHandler(os.Stdout,
			&slog.HandlerOptions{Level: slog.LevelDebug}))
	case ProdEnv:
		log = slog.New(slog.NewJSONHandler(os.Stdout,
			&slog.HandlerOptions{Level: slog.LevelInfo}))
	}
	return log
}

func connectToPostgres(cfg *config.Config, log *slog.Logger) (*postgresql.Postgres, error) {
	connString := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s",
		cfg.Database.Host, cfg.Database.Port, cfg.Database.User, cfg.Database.Password, cfg.Database.DBName)
	pg, err := postgresql.NewPG(context.Background(), connString, log, cfg)
	return pg, err
}
