package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/sashithaf16/peekalo/analyzer"
	"github.com/sashithaf16/peekalo/config"
	"github.com/sashithaf16/peekalo/logger"
	"github.com/sashithaf16/peekalo/metrics"
)

var validate = validator.New()

type UrlAnalyzeRequest struct {
	URL string `json:"url" validate:"required,url"`
}

type AnalyzeURLHandlerParams struct {
	cfg        *config.Config
	logger     logger.Logger
	httpClient analyzer.HttpClientInterface
}

func NewAnalyzeUrlHandler(cfg *config.Config, logger logger.Logger, httpClient analyzer.HttpClientInterface) *AnalyzeURLHandlerParams {
	return &AnalyzeURLHandlerParams{
		cfg:        cfg,
		logger:     logger,
		httpClient: httpClient,
	}
}

func (a *AnalyzeURLHandlerParams) AnalyzeURLHandler(w http.ResponseWriter, r *http.Request) {

	a.logger.Info().Msg("Received request to analyze URL")

	var req UrlAnalyzeRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		a.logger.Error().Err(err).Msg("Failed to decode request body")
		metrics.RequestInvalidCount.Inc()
		a.respondJSON(w, http.StatusBadRequest, APIResponse{Success: false, Error: "Invalid request payload"})
		return
	}

	err := validate.Struct(req)
	if err != nil {
		a.logger.Error().Err(err).Msg("Validation failed for request")
		metrics.RequestInvalidCount.Inc()
		a.respondJSON(w, http.StatusBadRequest, APIResponse{Success: false, Error: "Validation failed: " + err.Error()})
		return
	}

	metrics.RequestReceivedSuccessCount.Inc()

	analyzer := analyzer.NewAnalyzer(a.logger, a.cfg, a.httpClient)

	pageInfo, err := analyzer.AnalyzeURL(r.Context(), req.URL) // context from the request is propagated to the analyzer function
	if err != nil {
		a.logger.Error().Err(err).Msg("Failed to analyze URL")
		metrics.RequestAnalyzerFailureCount.Inc()
		a.respondJSON(w, http.StatusInternalServerError, APIResponse{Success: false, Error: "Failed to analyze URL: " + err.Error()})
		return
	}
	a.logger.Info().Msgf("Successfully analyzed URL: %s", req.URL)
	metrics.RequestAnalyzerSuccessCount.Inc()
	a.respondJSON(w, http.StatusOK, APIResponse{Success: true, Data: pageInfo})
}

type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

func (a *AnalyzeURLHandlerParams) respondJSON(w http.ResponseWriter, statusCode int, resp APIResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(resp)
}
