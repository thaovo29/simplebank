package token

import "time"

// Maker is an interface for managing tokens
type Maker interface{
	// Create Token creates a token for specific username and duration
	CreateToken(username string, role string, duration time.Duration) (string, *Payload, error)
	// VerifyToken checks if token is valid or not
	VerifyToken(token string) (*Payload, error)
} 