package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/sashithaf16/peekalo/config"

	mocks "github.com/sashithaf16/peekalo/_mocks"
	"github.com/sashithaf16/peekalo/logger"
)

// Helper to create a HTTP response for mock
func newHTTPResponse(body string, statusCode int) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     make(http.Header),
	}
}

func TestAnalyzeURLHandler_WithMockHTTPClient(t *testing.T) {
	// Setup real config and logger (you might want to customize these)
	cfg := &config.Config{
		LogLevel: "debug",
	}
	log := logger.CreateLogger(cfg.LogLevel)
	mockHTTPClient := new(mocks.MockHTTPClient)

	h := NewAnalyzeUrlHandler(cfg, log, mockHTTPClient)

	t.Run("successful analysis", func(t *testing.T) {
		urlToAnalyze := "https://example.com"
		reqBody := `{"url":"` + urlToAnalyze + `"}`

		// Mock HTTP client to return some HTML for analyzer to parse
		mockResponseBody := `<html><head><title>Example Domain</title></head><body>Example Description</body></html>`
		mockHTTPClient.On("Do", mock.AnythingOfType("*http.Request")).Return(
			newHTTPResponse(mockResponseBody, 200),
			nil,
		).Once()

		req := httptest.NewRequest(http.MethodPost, "/analyze", bytes.NewBufferString(reqBody))
		w := httptest.NewRecorder()

		h.AnalyzeURLHandler(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var apiResp APIResponse
		err := json.NewDecoder(resp.Body).Decode(&apiResp)
		assert.NoError(t, err)
		assert.True(t, apiResp.Success)
		assert.NotNil(t, apiResp.Data)

		mockHTTPClient.AssertExpectations(t)
	})

	t.Run("invalid json request", func(t *testing.T) {
		invalidBody := `{"url": "https://example.com",` // invalid JSON

		req := httptest.NewRequest(http.MethodPost, "/analyze", bytes.NewBufferString(invalidBody))
		w := httptest.NewRecorder()

		h.AnalyzeURLHandler(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

		var apiResp APIResponse
		err := json.NewDecoder(resp.Body).Decode(&apiResp)
		assert.NoError(t, err)
		assert.False(t, apiResp.Success)
		assert.Contains(t, apiResp.Error, "Invalid request payload")
	})

	t.Run("validation failure - bad url", func(t *testing.T) {
		reqBody := `{"url":"invalid-url"}`

		req := httptest.NewRequest(http.MethodPost, "/analyze", bytes.NewBufferString(reqBody))
		w := httptest.NewRecorder()

		h.AnalyzeURLHandler(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

		var apiResp APIResponse
		err := json.NewDecoder(resp.Body).Decode(&apiResp)
		assert.NoError(t, err)
		assert.False(t, apiResp.Success)
		assert.Contains(t, apiResp.Error, "Validation failed")
	})

	t.Run("http client returns error", func(t *testing.T) {
		urlToAnalyze := "https://example.com"
		reqBody := `{"url":"` + urlToAnalyze + `"}`

		mockHTTPClient.On("Do", mock.AnythingOfType("*http.Request")).Return(
			nil,
			errors.New("mocked network error"),
		).Once()

		req := httptest.NewRequest(http.MethodPost, "/analyze", bytes.NewBufferString(reqBody))
		w := httptest.NewRecorder()

		h.AnalyzeURLHandler(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

		var apiResp APIResponse
		err := json.NewDecoder(resp.Body).Decode(&apiResp)
		assert.NoError(t, err)
		assert.False(t, apiResp.Success)
		assert.Contains(t, apiResp.Error, "Failed to analyze URL")

		mockHTTPClient.AssertExpectations(t)
	})
}
