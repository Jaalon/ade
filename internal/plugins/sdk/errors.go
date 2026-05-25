package sdk

import "errors"

var (
	ErrInvalidConfig    = errors.New("invalid plugin configuration")
	ErrServerStartup    = errors.New("server startup failed")
	ErrRegistration     = errors.New("registration failed")
	ErrServerNotRunning = errors.New("server is not running")
)
