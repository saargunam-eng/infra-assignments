package domain

import "time"

// Config is the core entity managed by this service.
type Config struct {
	ID        string    `json:"id"`
	Host      string    `json:"host"`
	Port      int       `json:"port"`
	AppName   string    `json:"app_name"`
	LogLevel  string    `json:"log_level"`
	UpdatedAt time.Time `json:"updated_at"`
}
