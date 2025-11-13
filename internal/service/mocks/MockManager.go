package mocks

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
)

type MockManager struct {
	mock.Mock
}

func (m *MockManager) Do(ctx context.Context, fn func(context.Context) error) error {
	args := m.Called(ctx, fn)
	return args.Error(0)
}
