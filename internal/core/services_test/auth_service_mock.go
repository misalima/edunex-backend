package services_test

import (
	"github.com/stretchr/testify/mock"
)

// MockAuthRepo centralizes mocked authentication repository methods.
type MockAuthRepo struct {
	mock.Mock
}

func (m *MockAuthRepo) SomeMethod() error {
	// Add mocked implementation here
	return nil
}