package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"rent-app/docs"
	"rent-app/internal/adapter"
	"rent-app/internal/config"
	"rent-app/internal/database"
	apartmentHandler "rent-app/internal/handler/apartment"
	authHandler "rent-app/internal/handler/auth"
	userHandler "rent-app/internal/handler/user"
	slogpretty "rent-app/internal/lib/logger/handlers/slogpretty"
	mwAuth "rent-app/internal/middleware/auth"
	mwLogger "rent-app/internal/middleware/logger"
	apartmentRepo "rent-app/internal/repository/apartment"
	authRepo "rent-app/internal/repository/auth"
	userRepo "rent-app/internal/repository/user"
	apartmentService "rent-app/internal/service/apartment"
	authService "rent-app/internal/service/auth"
	userService "rent-app/internal/service/user"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger"
)

// @title           Rent App API
// @version         1.0
// @description     API для управления арендой квартир
// @termsOfService  http://swagger.io/terms/

// @contact.name Yuri Oschepkov
// @contact.email genda5656@gmail.com

// @license.name MIT License
// @license.url https://mit-license.org/

// @host      localhost:8080
// @BasePath  /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
// @schemes http https

const (
	dbTimeout          = 3 * time.Second
	shutdownTimeout    = 5 * time.Second
	tokenCleanupPeriod = 1 * time.Hour
	envLocal           = "local"
	envDev             = "development"
	envProd            = "production"
)

func main() {
	cfg, err := config.ConfigLoad()
	log := setupLogger(cfg.Env)
	log.Info("starting server", slog.String("env", cfg.Env))
	log.Debug("debug message", slog.String("env", cfg.Env))
	if err != nil {
		log.Error("failed to load config", slog.String("env", cfg.Env), slog.String("err", err.Error()))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	db, err := database.Connect(ctx, cfg.DSN)
	if err != nil {
		log.Error("failed to connect to database", slog.String("env", cfg.Env), slog.String("err", err.Error()))
		return
	}
	defer db.Close()

	log.Info("successfully connected to database")

	userRepository := userRepo.NewPostgresRepo(db)
	authRepository := authRepo.NewPostgresRepo(db)
	apartmentRepository := apartmentRepo.NewPostgresRepo(db)

	userServiceInstance := userService.NewService(userRepository)
	authServiceInstance := authService.NewService(authRepository)
	apartmentServiceInstance := apartmentService.NewService(apartmentRepository)

	userAuthenticator := adapter.NewUserAuthenticatorAdapter(userServiceInstance)

	userHandlerInstance := userHandler.NewHandler(userServiceInstance)
	authHandlerInstance := authHandler.NewHandler(authServiceInstance, userAuthenticator)
	apartmentHandlerInstance := apartmentHandler.NewHandler(apartmentServiceInstance)

	mux := chi.NewRouter()

	mux.Use(middleware.RequestID)
	mux.Use(mwLogger.New(log))

	docs.SwaggerInfo.Host = "localhost:" + cfg.Port
	docs.SwaggerInfo.BasePath = "/"
	docs.SwaggerInfo.Schemes = []string{"http", "https"}

	mux.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("http://localhost:"+cfg.Port+"/swagger/doc.json"),
	))

	mux.Get("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("welcome to the rent app!"))
	})

	mux.Route("/api/auth", func(r chi.Router) {
		r.Post("/login", authHandlerInstance.Login)
		r.Post("/refresh", authHandlerInstance.Refresh)
	})

	mux.Route("/api/users", func(r chi.Router) {
		r.With(mwAuth.OptionalAuthMiddleware(authServiceInstance)).Post("/", userHandlerInstance.CreateUser)
		r.With(mwAuth.AuthMiddleware(authServiceInstance)).Route("/", func(r chi.Router) {
			r.Method("GET", "/", mwAuth.RequireAdmin(http.HandlerFunc(userHandlerInstance.GetAllUsers)))
			r.Method("GET", "/{id}", mwAuth.RequireAdmin(http.HandlerFunc(userHandlerInstance.GetUserByID)))
			r.Method("GET", "/email/{email}", mwAuth.RequireAdmin(http.HandlerFunc(userHandlerInstance.GetUserByEmail)))
			r.Put("/{id}", userHandlerInstance.UpdateUser)
			r.Put("/{id}/reset-password", userHandlerInstance.ResetPassword)
			r.Delete("/{id}", userHandlerInstance.DeleteUser)
		})
	})

	mux.Route("/api/apartments", func(r chi.Router) {
		r.With(mwAuth.AuthMiddleware(authServiceInstance)).Route("/", func(r chi.Router) {
			r.Post("/", apartmentHandlerInstance.CreateApartment)
			r.Get("/", apartmentHandlerInstance.GetAllApartments)
			r.Get("/{id}", apartmentHandlerInstance.GetApartmentByID)
			r.Get("/owner/{ownerID}", apartmentHandlerInstance.GetApartmentsByOwnerID)
			r.Put("/{id}", apartmentHandlerInstance.UpdateApartment)
			r.Delete("/{id}", apartmentHandlerInstance.DeleteApartment)
		})
	})

	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: mux,
	}

	bgCtx, bgCancel := context.WithCancel(context.Background())
	defer bgCancel()

	// убираем токены из блеклиста
	go func() {
		ticker := time.NewTicker(tokenCleanupPeriod)
		defer ticker.Stop()

		if err := authRepository.CleanupExpiredTokens(); err != nil {
			log.Error("failed to cleanup expired tokens", slog.String("env", cfg.Env), slog.String("err", err.Error()))
		}

		for {
			select {
			case <-ticker.C:
				if err := authRepository.CleanupExpiredTokens(); err != nil {
					log.Error("failed to cleanup expired tokens", slog.String("env", cfg.Env), slog.String("err", err.Error()))
				}
			case <-bgCtx.Done():
				return
			}
		}
	}()

	go func() {
		log.Info("starting server", slog.String("env", cfg.Env), slog.String("port", cfg.Port))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("failed to start server", slog.String("env", cfg.Env), slog.String("err", err.Error()))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")

	ctx, cancel = context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Error("server forced to shutdown", slog.String("env", cfg.Env), slog.String("err", err.Error()))
	}

	log.Info("Server exited")
}

func setupLogger(env string) *slog.Logger {
	var logger *slog.Logger

	switch env {
	case envLocal:
		logger = setupPrettySlog()
	case envDev:
		logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}))
	case envProd:
		logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}))
	default:
		logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}))
	}
	return logger
}

func setupPrettySlog() *slog.Logger {
	opts := slogpretty.PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	}

	handler := opts.NewPrettyHandler(os.Stdout)

	return slog.New(handler)
}
