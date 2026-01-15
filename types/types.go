// Package types its for all types in project
package types

type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

type UserResponse struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Surname     string  `json:"surname"`
	PhoneNumber string  `json:"phone_number"`
	Balance     float64 `json:"balance"`
}

type ResponseForAuth struct {
	Success bool    `json:"success"`
	Message string  `json:"message,omitempty"`
	UserID  int64   `json:"user_id,omitempty"`
	Session string  `json:"session_id,omitempty"`
	Balance float64 `json:"balance,omitempty"`
	Error   string  `json:"error,omitempty"`
}

var BcryptSalt = 12

var DefaultSession = 32