package yourservice

import (
	"context"
)

// Create ...
func (s *Service) Create(ctx context.Context) error {

	// Create
	err := s.Store.Create(ctx)
	if err != nil {
		return err
	}
	return nil

}
