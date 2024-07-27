package some_db

import (
	"dev/yourservice.git/business/i"
)

// Config is the required properties to use the database.
type Config struct {
	ProjectID    string
	EmulatorHost string
	Setting      int64
}

// SomeDB would be replaced by the actual client
type SomeDB struct {
	Log i.Logger
}

// Close will return dispose the client
func (s *SomeDB) Close() {

	// Close the client
	return

}

// NewClient will return the third party client
func NewClient(log i.Logger) (*SomeDB, error) {

	// Create the client
	return &SomeDB{Log: log}, nil

}
