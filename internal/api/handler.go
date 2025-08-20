package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/MiladJlz/dekamond-task/internal/api/dto"
	"github.com/MiladJlz/dekamond-task/internal/auth"
	"github.com/MiladJlz/dekamond-task/internal/db"
	"github.com/MiladJlz/dekamond-task/internal/otp"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type Handler struct {
	store     *db.Store
	otp       *otp.RedisOTP
	jwtSecret string
	logger    *zap.SugaredLogger
}

// NewHandler constructor
func NewHandler(s *db.Store, r *otp.RedisOTP, jwtSecret string, logger *zap.SugaredLogger) *Handler {
	return &Handler{store: s, otp: r, jwtSecret: jwtSecret, logger: logger}
}

func JSONError(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(dto.ErrorResponse{Error: message})
}

// JSONErrorWithLog logs the internal error and returns a safe public message
func (h *Handler) JSONErrorWithLog(w http.ResponseWriter, publicMessage string, code int, err error, context string, fields ...interface{}) {
	if err != nil {
		h.logger.Errorw(context, append([]any{"error", err}, fields...)...)
	} else {
		h.logger.Errorw(context, fields...)
	}
	JSONError(w, publicMessage, code)
}

// RequestOTP godoc
// @Summary Request OTP
// @Description Send OTP to phone (logged in server for now)
// @Tags auth
// @Accept  json
// @Produce  json
// @Param request body dto.RequestOTPRequest true "Request body"
// @Success 200 {object} dto.RequestOTPResponse
// @Failure 429 {object} dto.ErrorResponse
// @Failure 400 {object} dto.ErrorResponse
// @Router /request-otp [post]
func (h *Handler) RequestOTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	var req dto.RequestOTPRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.JSONErrorWithLog(w, "Invalid request body", http.StatusBadRequest, err, "decode request body failed")
		return
	}

	if req.Phone == "" {
		JSONError(w, "Phone number is required", http.StatusBadRequest)
		return
	}

	rateLimitStart := time.Now()
	allowed, rlErr := h.otp.RateLimit(req.Phone)
	if rlErr != nil {
		h.JSONErrorWithLog(w, "Temporary service issue. Please try again.", http.StatusServiceUnavailable, rlErr, "rate limit error", "phone", req.Phone)
		return
	}
	if !allowed {
		JSONError(w, "Rate limit exceeded", http.StatusTooManyRequests)
		return
	}
	rateLimitDuration := time.Since(rateLimitStart)

	generateStart := time.Now()
	code, genErr := h.otp.Generate(req.Phone)
	if genErr != nil {
		h.JSONErrorWithLog(w, "Temporary service issue. Please try again.", http.StatusServiceUnavailable, genErr, "generate otp error", "phone", req.Phone)
		return
	}
	generateDuration := time.Since(generateStart)

	h.logger.Infow("otp generated", "phone", req.Phone, "code", code)
	h.logger.Infow("otp request perf", "rate_limit_ms", rateLimitDuration.Milliseconds(), "generate_ms", generateDuration.Milliseconds(), "total_ms", time.Since(start).Milliseconds())

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(dto.RequestOTPResponse{Message: "OTP sent"})
}

// VerifyOTP godoc
// @Summary Verify OTP
// @Description Verify OTP and login/register user
// @Tags auth
// @Accept  json
// @Produce  json
// @Param body body dto.VerifyOTPRequest true "Request body for OTP verification"
// @Success 200 {object} dto.VerifyOTPResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /verify-otp [post]
func (h *Handler) VerifyOTP(w http.ResponseWriter, r *http.Request) {
	var req dto.VerifyOTPRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.JSONErrorWithLog(w, "Invalid request body", http.StatusBadRequest, err, "decode request body failed")
		return
	}

	if req.Phone == "" || req.Code == "" {
		JSONError(w, "Phone and code are required", http.StatusBadRequest)
		return
	}

	valid, valErr := h.otp.Validate(req.Phone, req.Code)
	if valErr != nil {
		h.JSONErrorWithLog(w, "Temporary service issue. Please try again.", http.StatusServiceUnavailable, valErr, "validate otp error", "phone", req.Phone)
		return
	}

	if !valid {
		JSONError(w, "Invalid OTP", http.StatusUnauthorized)
		return
	}

	exists, err := h.store.UserExists(req.Phone)
	if err != nil {
		h.JSONErrorWithLog(w, "Database temporarily unavailable. Please try again.", http.StatusInternalServerError, err, "user exists query failed", "phone", req.Phone)
		return
	}

	if !exists {
		if err := h.store.CreateUser(req.Phone); err != nil {
			h.JSONErrorWithLog(w, "Failed to create user", http.StatusInternalServerError, err, "create user failed", "phone", req.Phone)
			return
		}
		h.logger.Infow("user created", "phone", req.Phone)
	}

	token, err := auth.GenerateJWT(req.Phone, h.jwtSecret)
	if err != nil {
		h.JSONErrorWithLog(w, "Failed to issue token", http.StatusInternalServerError, err, "jwt sign failed", "phone", req.Phone)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(dto.VerifyOTPResponse{Message: "Login success", Token: token})
}

// GetUser godoc
// @Summary Get user by ID
// @Description Retrieve a single user by their ID
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Success 200 {object} types.User
// @Failure 401 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /users/{id} [get]
func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		JSONError(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	user, err := h.store.GetUserByID(id)
	if err != nil {
		h.JSONErrorWithLog(w, "User not found", http.StatusNotFound, err, "get user by id failed", "id", id)

		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// GetUsers godoc
// @Summary Get users list
// @Description Retrieve list of users with pagination and search
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param offset query int false "Offset for pagination (default: 0)"
// @Param limit query int false "Limit for pagination (default: 10, max: 100)"
// @Param search query string false "Search by phone number"
// @Success 200 {object} dto.UserListResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /users [get]
func (h *Handler) GetUsers(w http.ResponseWriter, r *http.Request) {
	offsetStr := r.URL.Query().Get("offset")
	limitStr := r.URL.Query().Get("limit")
	search := r.URL.Query().Get("search")

	offset := 0
	limit := 10

	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	users, err := h.store.GetUsers(limit, offset, search)
	if err != nil {
		h.JSONErrorWithLog(w, "Failed to fetch users", http.StatusInternalServerError, err, "list users failed", "offset", offset, "limit", limit, "search", search)
		return
	}

	total, err := h.store.GetUsersCount(search)
	if err != nil {
		h.JSONErrorWithLog(w, "Failed to get total count", http.StatusInternalServerError, err, "count users failed", "search", search)
		return
	}

	response := dto.UserListResponse{
		Users:  users,
		Total:  total,
		Offset: offset,
		Limit:  limit,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HealthCheck godoc
// @Summary Health check
// @Description Check service health for PostgreSQL and Redis
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} dto.HealthCheckResponse
// @Failure 503 {object} dto.ErrorResponse
// @Router /health [get]
func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	pgErr := h.store.DB.PingContext(ctx)
	if pgErr != nil {
		h.logger.Errorw("postgres health check failed", "error", pgErr)
	}

	redisErr := h.otp.PingRedis()
	if redisErr != nil {
		h.logger.Errorw("redis health check failed", "error", redisErr)
	}

	overallHealthy := pgErr == nil && redisErr == nil
	if !overallHealthy {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(dto.ErrorResponse{Error: "one or more dependencies are down"})
		return
	}

	resp := dto.HealthCheckResponse{
		Status:   "healthy",
		Postgres: dto.ComponentHealth{Status: "up"},
		Redis:    dto.ComponentHealth{Status: "up"},
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}
