package dto

import "github.com/MiladJlz/dekamond-task/internal/types"

// RequestOTPResponse is the response for OTP request endpoint
// @Description Response for OTP request
type RequestOTPResponse struct {
	Message string `json:"message" example:"OTP sent" description:"Success message"`
}

// VerifyOTPResponse is the response for OTP verification endpoint
// @Description Response for OTP verification
type VerifyOTPResponse struct {
	Message string `json:"message" example:"Login success" description:"Success message"`
	Token   string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." description:"JWT authentication token"`
}

// UserListResponse is the response for user list endpoint
// @Description Response containing paginated list of users
type UserListResponse struct {
	Users  []types.User `json:"users" description:"List of users"`
	Total  int          `json:"total" example:"100" description:"Total number of users"`
	Offset int          `json:"offset" example:"0" description:"Current offset for pagination"`
	Limit  int          `json:"limit" example:"10" description:"Number of users per page"`
}

// ComponentHealth describes the health of a single dependency (e.g., Postgres, Redis)
// @Description Health details for a single dependency
type ComponentHealth struct {
	Status string `json:"status" example:"up" description:"Component status (up/down)"`
}

// HealthCheckResponse is the response for health check endpoint
// @Description Response for health check
type HealthCheckResponse struct {
	Status   string          `json:"status" example:"healthy" description:"Overall service status"`
	Postgres ComponentHealth `json:"postgres" description:"PostgreSQL status information"`
	Redis    ComponentHealth `json:"redis" description:"Redis status information"`
}

// ErrorResponse is the standard error response format
// @Description Standard error response format
type ErrorResponse struct {
	Error string `json:"error" example:"Error message" description:"Error description"`
}
