package types

import "time"

// User represents a user in the system
// @Description User entity with phone number and registration details
type User struct {
	ID        uint64    `json:"id" example:"1" description:"Unique user identifier"`
	Phone     string    `json:"phone" example:"+1234567890" description:"User's phone number"`
	CreatedAt time.Time `json:"created_at" example:"2025-08-19T12:00:00Z" description:"User registration timestamp"`
}
