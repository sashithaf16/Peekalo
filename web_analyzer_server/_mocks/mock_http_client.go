package mocks

import (
	"net/http"

	"github.com/stretchr/testify/mock"
)

// --- Mock HTTP Client ---
type MockHTTPClient struct {
	mock.Mock
}

func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	args := m.Called(req)
	resp, _ := args.Get(0).(*http.Response)
	return resp, args.Error(1)
}
