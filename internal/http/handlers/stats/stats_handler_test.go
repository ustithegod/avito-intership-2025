package stats_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"avito-intership-2025/internal/http/api"
	"avito-intership-2025/internal/http/handlers"
	"avito-intership-2025/internal/http/handlers/mocks"
	"avito-intership-2025/internal/http/handlers/stats"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestStatsHandler_GetStatistics_DefaultSort_Success(t *testing.T) {
	mockService := mocks.NewMockStatsService(t)
	h := stats.NewStatsHandler(handlers.NewLogger(), mockService)

	expected := api.StatsResponse{
		Pr: api.PrStats{
			PrCount:   10,
			OpenPrs:   7,
			MergedPrs: 3,
		},
		User: []api.UserStats{
			{UserID: "u1", Username: "Alice", AssignmentCount: 5},
		},
	}

	// Без параметра sort -> должно быть "desc" по умолчанию
	mockService.On("GetStatistics", mock.Anything, "desc").Return(expected, nil).Once()

	req := httptest.NewRequest(http.MethodGet, "/statistics", nil)
	w := httptest.NewRecorder()

	h.GetStatistics(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var got api.StatsResponse
	err := jsonNewDecoder(w.Body).Decode(&got)
	assert.NoError(t, err)
	assert.Equal(t, expected, got)
}

func TestStatsHandler_GetStatistics_SortAsc_Success(t *testing.T) {
	mockService := mocks.NewMockStatsService(t)
	h := stats.NewStatsHandler(handlers.NewLogger(), mockService)

	expected := api.StatsResponse{
		Pr: api.PrStats{
			PrCount:   2,
			OpenPrs:   1,
			MergedPrs: 1,
		},
		User: []api.UserStats{
			{UserID: "u2", Username: "Bob", AssignmentCount: 3},
		},
	}

	mockService.On("GetStatistics", mock.Anything, "asc").Return(expected, nil).Once()

	req := httptest.NewRequest(http.MethodGet, "/statistics?sort=asc", nil)
	w := httptest.NewRecorder()

	h.GetStatistics(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var got api.StatsResponse
	err := jsonNewDecoder(w.Body).Decode(&got)
	assert.NoError(t, err)
	assert.Equal(t, expected, got)
}

func TestStatsHandler_GetStatistics_BadSortParam(t *testing.T) {
	mockService := mocks.NewMockStatsService(t)
	h := stats.NewStatsHandler(handlers.NewLogger(), mockService)

	// Неверный параметр sort -> 400 BAD_REQUEST
	req := httptest.NewRequest(http.MethodGet, "/statistics?sort=zzz", nil)
	w := httptest.NewRecorder()

	h.GetStatistics(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	resp := handlers.DecodeErrorResponse(t, w.Body)
	assert.Equal(t, api.ErrBadRequest, resp.Error.Code)
	// Можно дополнительно проверить текст:
	// assert.Contains(t, resp.Error.Message, "sort must be 'desc' or 'asc'")
}

func TestStatsHandler_GetStatistics_ServiceError(t *testing.T) {
	mockService := mocks.NewMockStatsService(t)
	h := stats.NewStatsHandler(handlers.NewLogger(), mockService)

	mockService.
		On("GetStatistics", mock.Anything, "desc").
		Return(api.StatsResponse{}, errors.New("db error")).
		Once()

	req := httptest.NewRequest(http.MethodGet, "/statistics", nil)
	w := httptest.NewRecorder()

	h.GetStatistics(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	resp := handlers.DecodeErrorResponse(t, w.Body)
	assert.Equal(t, api.ErrInternalErr, resp.Error.Code)
}

// helper to avoid importing encoding/json in each test function
type decoder interface{ Decode(v any) error }

func jsonNewDecoder(b *bytes.Buffer) decoder {
	return json.NewDecoder(b)
}
