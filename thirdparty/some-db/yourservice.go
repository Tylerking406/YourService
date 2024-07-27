package some_db

import (
	"context"
)

// Create ...
func (s *SomeDB) Create(ctx context.Context) error {

	// Create and return the entity
	println("Wrote entity to SomeDB")
	return nil

}
