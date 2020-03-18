package datasource

import "github.com/stretchr/testify/mock"

type MockPropertyHandler struct {
	mock.Mock
}

func (m *MockPropertyHandler) isPropertyConsistent(src interface{}) bool {
	args := m.Called(src)
	return args.Bool(0)
}

func (m *MockPropertyHandler) Handle(src []byte) error {
	args := m.Called(src)
	return args.Error(0)
}
