package yourservice

import (
	"context"
	"dev/yourservice.git/business/i"
)

// Constants
const ()

// Service encapsulates core yourservice functionality
type Service struct {
	Log   i.Logger
	Store Store
}

// Store encapsulates third-party dependencies
type Store interface {
	Create(ctx context.Context) error
}
