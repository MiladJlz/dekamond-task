package main

import (
	"net/http"

	_ "github.com/MiladJlz/dekamond-task/docs"
	"github.com/MiladJlz/dekamond-task/internal/api"
	_ "github.com/MiladJlz/dekamond-task/internal/api/dto"
	"github.com/MiladJlz/dekamond-task/internal/config"
	"github.com/MiladJlz/dekamond-task/internal/db"
	"github.com/MiladJlz/dekamond-task/internal/otp"
	_ "github.com/MiladJlz/dekamond-task/internal/types"
	"github.com/go-chi/chi/v5"
	httpSwagger "github.com/swaggo/http-swagger/v2"
	"go.uber.org/zap"
)

// @title OTP Authentication API
// @version 1.0
// @description OTP-based authentication and user management service
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

// @host localhost:8080
// @BasePath /v1
func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		panic("failed to init zap logger: " + err.Error())
	}
	defer logger.Sync()
	sugar := logger.Sugar()

	cfg := config.LoadConfig(logger)

	db, err := db.NewPostgresDB(cfg.PostgresDSN)
	if err != nil {
		sugar.Fatalw("cannot connect to postgres", "error", err)
	}
	if pingErr := db.DB.Ping(); pingErr != nil {
		sugar.Fatalw("postgres ping failed", "error", pingErr)
	}

	redisClient := otp.NewRedisClient(cfg.RedisAddr, cfg.OTPTTL, cfg.RateLimit, cfg.RateLimitWindow)
	if err := redisClient.PingRedis(); err != nil {
		sugar.Fatalw("redis not reachable", "error", err)
	}

	h := api.NewHandler(db, redisClient, cfg.JWTSecret, sugar)

	r := chi.NewRouter()

	// versioned API
	v1 := chi.NewRouter()

	// Health check
	v1.Get("/health", h.HealthCheck)

	// Auth routes
	v1.Post("/request-otp", h.RequestOTP)
	v1.Post("/verify-otp", h.VerifyOTP)

	// User management routes (protected with JWT)
	v1.Get("/users", h.JWTAuthMiddleware(h.GetUsers))
	v1.Get("/users/{id}", h.JWTAuthMiddleware(h.GetUser))

	// Swagger documentation (versioned)
	v1.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("http://localhost:8080/v1/swagger/doc.json"),
	))

	r.Mount("/v1", v1)

	sugar.Infow("server starting", "port", cfg.AppPort)
	sugar.Infow("otp configuration",
		"otp_ttl", cfg.OTPTTL.String(),
		"rate_limit", cfg.RateLimit,
		"rate_limit_window", cfg.RateLimitWindow.String())
	sugar.Fatalw("server failed", "error", http.ListenAndServe(":"+cfg.AppPort, r))
}
