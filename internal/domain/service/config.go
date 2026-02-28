package service

// Config defines the interface for application configuration
// This interface allows other layers to access configuration
// without coupling to the infrastructure implementation
type Config interface {
	GetAppCode() string
}
