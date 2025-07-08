package analyzer

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"testing"

	mocks "github.com/sashithaf16/peekalo/_mocks"
	"github.com/sashithaf16/peekalo/config"
	"github.com/sashithaf16/peekalo/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAnalyzeURL_Success(t *testing.T) {
	mockHTML := `
		<!DOCTYPE html>
		<html>
		<head>
			<title>Test Page</title>
		</head>
		<body>
			<h1>Main Heading</h1>
			<h2>Subheading</h2>
			<a href="https://external.com">External</a>
			<a href="/internal">Internal</a>
			<a href="mailto:someone@example.com">Email</a>
			<form>
				<input type="text" name="user"/>
				<input type="password" name="pass"/>
			</form>
		</body>
		</html>
	`

	// Setup mock HTTP client
	mockClient := new(mocks.MockHTTPClient)
	mockResp := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewBufferString(mockHTML)),
	}
	mockClient.On("Do", mock.Anything).Return(mockResp, nil)

	// Initialize analyzer with mocks
	cfg := &config.Config{
		LogLevel: "debug",
	}
	logger := logger.CreateLogger(cfg.LogLevel)
	an := NewAnalyzer(logger, cfg, mockClient)

	// Run AnalyzeURL
	ctx := context.Background()
	result, err := an.AnalyzeURL(ctx, "http://example.com")

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, "HTML 5", result.HTMLVersion)
	assert.Equal(t, "Test Page", result.Title)
	assert.Equal(t, map[string]int{"h1": 1, "h2": 1, "h3": 0, "h4": 0, "h5": 0, "h6": 0}, result.Headings)
	assert.Equal(t, 1, result.Links.Internal)
	assert.Equal(t, 1, result.Links.External)
	assert.Equal(t, 1, result.Links.Inaccessible)
	assert.True(t, result.HasLogin)

	mockClient.AssertExpectations(t)
}

func TestAnalyzeURL_ErrorFetchingURL(t *testing.T) {
	mockCfg := &config.Config{
		LogLevel: "debug",
	}
	logger := logger.CreateLogger(mockCfg.LogLevel)
	mockClient := new(mocks.MockHTTPClient)

	mockClient.On("Do", mock.Anything).Return((*http.Response)(nil), errors.New("mocked network error"))

	an := NewAnalyzer(logger, mockCfg, mockClient)

	ctx := context.Background()
	_, err := an.AnalyzeURL(ctx, "http://example.com")

	expectedErr := "failed to fetch URL: mocked network error"
	assert.EqualError(t, err, expectedErr, "unexpected error message")
}

func TestAnalyzeURL_SocialLoginDetected(t *testing.T) {
	mockHTML := `
		<!DOCTYPE html>
		<html>
		<head>
			<title>Social Login Test</title>
		</head>
		<body>
			<form action="/login">
				<a href="#" class="btn google">Login with Google</a>
			</form>
		</body>
		</html>
	`

	// Setup mock HTTP client
	mockClient := new(mocks.MockHTTPClient)
	mockResp := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewBufferString(mockHTML)),
	}
	mockClient.On("Do", mock.Anything).Return(mockResp, nil)

	// Initialize analyzer with mocks
	cfg := &config.Config{
		LogLevel: "debug",
	}
	log := logger.CreateLogger(cfg.LogLevel)
	an := NewAnalyzer(log, cfg, mockClient)

	// Run AnalyzeURL
	ctx := context.Background()
	result, err := an.AnalyzeURL(ctx, "http://example.com")

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, "Social Login Test", result.Title)
	assert.True(t, result.HasLogin, "Expected HasLogin to be true for social login anchor")

	mockClient.AssertExpectations(t)
}
