package dto

// RequestOTPRequest is the request body for OTP request.
// @Description Request body for OTP request
type RequestOTPRequest struct {
	Phone string `json:"phone" example:"+1234567890" binding:"required" description:"User's phone number"`
}

// VerifyOTPRequest is the request body for OTP verification.
// @Description Request body for OTP verification
type VerifyOTPRequest struct {
	Phone string `json:"phone" example:"+1234567890" binding:"required" description:"User's phone number"`
	Code  string `json:"code" example:"123456" binding:"required" description:"6-digit OTP code"`
}
