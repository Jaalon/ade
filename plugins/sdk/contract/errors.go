package contract

import "errors"

var (
	ErrPluginNotFound      = errors.New("plugin not found")
	ErrPluginAlreadyExists = errors.New("plugin already registered")
	ErrInvalidDescriptor   = errors.New("invalid plugin descriptor")
	ErrRegistrationFailed  = errors.New("plugin registration failed")
)
