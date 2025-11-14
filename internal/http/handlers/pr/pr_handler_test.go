package pr_test

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
	"avito-intership-2025/internal/http/handlers/pr"
	repo "avito-intership-2025/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Create
func TestPrHandler_Create_Success(t *testing.T) {
	mockService := mocks.NewMockPrService(t)
	h := pr.NewPrHandler(handlers.NewLogger(), mockService)

	reqBody := pr.CreateRequest{
		PrID:     "pr1",
		PrName:   "My PR",
		AuthorId: "u1",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/pr/create", bytes.NewReader(body))
	w := httptest.NewRecorder()

	expectedPR := &api.PullRequestSchema{
		ID:       "pr1",
		Name:     "My PR",
		AuthorID: "u1",
		Status:   "open",
	}
	mockService.On("Create", mock.Anything, "pr1", "My PR", "u1").Return(expectedPR, nil)

	h.Create(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	var resp api.PrResponse
	err := json.NewDecoder(w.Body).Decode(&resp)
	assert.NoError(t, err)
	assert.Equal(t, *expectedPR, resp.PullRequest)
}

func TestPrHandler_Create_BadJSON(t *testing.T) {
	mockService := mocks.NewMockPrService(t)
	h := pr.NewPrHandler(handlers.NewLogger(), mockService)

	req := httptest.NewRequest(http.MethodPost, "/pr/create", bytes.NewReader([]byte("{invalid json")))
	w := httptest.NewRecorder()

	h.Create(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	resp := handlers.DecodeErrorResponse(t, w.Body)
	assert.Equal(t, api.ErrBadRequest, resp.Error.Code)
}

func TestPrHandler_Create_ValidationError(t *testing.T) {
	mockService := mocks.NewMockPrService(t)
	h := pr.NewPrHandler(handlers.NewLogger(), mockService)

	reqBody := pr.CreateRequest{
		PrID:     "",    // required
		PrName:   "abc", // min 5
		AuthorId: "",    // required
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/pr/create", bytes.NewReader(body))
	w := httptest.NewRecorder()

	h.Create(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	resp := handlers.DecodeErrorResponse(t, w.Body)
	assert.Equal(t, api.ErrValidationErr, resp.Error.Code)
}

func TestPrHandler_Create_PRExists(t *testing.T) {
	mockService := mocks.NewMockPrService(t)
	h := pr.NewPrHandler(handlers.NewLogger(), mockService)

	reqBody := pr.CreateRequest{PrID: "pr1", PrName: "My PR", AuthorId: "u1"}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/pr/create", bytes.NewReader(body))
	w := httptest.NewRecorder()

	mockService.On("Create", mock.Anything, "pr1", "My PR", "u1").Return(nil, repo.ErrPRExists)

	h.Create(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
	resp := handlers.DecodeErrorResponse(t, w.Body)
	assert.Equal(t, api.ErrCodePRExists, resp.Error.Code)
}

func TestPrHandler_Create_NotFound(t *testing.T) {
	mockService := mocks.NewMockPrService(t)
	h := pr.NewPrHandler(handlers.NewLogger(), mockService)

	reqBody := pr.CreateRequest{PrID: "pr1", PrName: "My PR", AuthorId: "u1"}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/pr/create", bytes.NewReader(body))
	w := httptest.NewRecorder()

	mockService.On("Create", mock.Anything, "pr1", "My PR", "u1").Return(nil, repo.ErrNotFound)

	h.Create(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	resp := handlers.DecodeErrorResponse(t, w.Body)
	assert.Equal(t, api.ErrCodeNotFound, resp.Error.Code)
}

func TestPrHandler_Create_InternalError(t *testing.T) {
	mockService := mocks.NewMockPrService(t)
	h := pr.NewPrHandler(handlers.NewLogger(), mockService)

	reqBody := pr.CreateRequest{PrID: "pr1", PrName: "My PR", AuthorId: "u1"}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/pr/create", bytes.NewReader(body))
	w := httptest.NewRecorder()

	mockService.On("Create", mock.Anything, "pr1", "My PR", "u1").Return(nil, errors.New("db error"))

	h.Create(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	resp := handlers.DecodeErrorResponse(t, w.Body)
	assert.Equal(t, api.ErrInternalErr, resp.Error.Code)
}

// ----------------- Merge -----------------
func TestPrHandler_Merge_Success(t *testing.T) {
	mockService := mocks.NewMockPrService(t)
	h := pr.NewPrHandler(handlers.NewLogger(), mockService)

	reqBody := pr.MergeRequest{PrID: "pr1"}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/pr/merge", bytes.NewReader(body))
	w := httptest.NewRecorder()

	expectedPR := &api.PullRequestSchema{ID: "pr1", Name: "My PR", AuthorID: "u1", Status: "merged"}
	mockService.On("Merge", mock.Anything, "pr1").Return(expectedPR, nil)

	h.Merge(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp api.PrResponse
	err := json.NewDecoder(w.Body).Decode(&resp)
	assert.NoError(t, err)
	assert.Equal(t, *expectedPR, resp.PullRequest)
}

func TestPrHandler_Merge_BadJSON(t *testing.T) {
	mockService := mocks.NewMockPrService(t)
	h := pr.NewPrHandler(handlers.NewLogger(), mockService)

	req := httptest.NewRequest(http.MethodPost, "/pr/merge", bytes.NewReader([]byte("{invalid")))
	w := httptest.NewRecorder()

	h.Merge(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	resp := handlers.DecodeErrorResponse(t, w.Body)
	assert.Equal(t, api.ErrBadRequest, resp.Error.Code)
}

func TestPrHandler_Merge_ValidationError(t *testing.T) {
	mockService := mocks.NewMockPrService(t)
	h := pr.NewPrHandler(handlers.NewLogger(), mockService)

	reqBody := pr.MergeRequest{PrID: ""}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/pr/merge", bytes.NewReader(body))
	w := httptest.NewRecorder()

	h.Merge(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	resp := handlers.DecodeErrorResponse(t, w.Body)
	assert.Equal(t, api.ErrValidationErr, resp.Error.Code)
}

func TestPrHandler_Merge_NotFound(t *testing.T) {
	mockService := mocks.NewMockPrService(t)
	h := pr.NewPrHandler(handlers.NewLogger(), mockService)

	reqBody := pr.MergeRequest{PrID: "pr1"}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/pr/merge", bytes.NewReader(body))
	w := httptest.NewRecorder()

	mockService.On("Merge", mock.Anything, "pr1").Return(nil, repo.ErrNotFound)

	h.Merge(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	resp := handlers.DecodeErrorResponse(t, w.Body)
	assert.Equal(t, api.ErrCodeNotFound, resp.Error.Code)
}

func TestPrHandler_Merge_InternalError(t *testing.T) {
	mockService := mocks.NewMockPrService(t)
	h := pr.NewPrHandler(handlers.NewLogger(), mockService)

	reqBody := pr.MergeRequest{PrID: "pr1"}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/pr/merge", bytes.NewReader(body))
	w := httptest.NewRecorder()

	mockService.On("Merge", mock.Anything, "pr1").Return(nil, errors.New("db error"))

	h.Merge(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	resp := handlers.DecodeErrorResponse(t, w.Body)
	assert.Equal(t, api.ErrInternalErr, resp.Error.Code)
}

// ----------------- Reassign -----------------
func TestPrHandler_Reassign_Success(t *testing.T) {
	mockService := mocks.NewMockPrService(t)
	h := pr.NewPrHandler(handlers.NewLogger(), mockService)

	reqBody := pr.ReassignRequest{PrID: "pr1", OldReviewerID: "u1"}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/pr/reassign", bytes.NewReader(body))
	w := httptest.NewRecorder()

	expectedResp := &api.ReassignResponse{PullRequest: api.PullRequestSchema{ID: "pr1"}, ReplacedBy: "u2"}
	mockService.On("Reassign", mock.Anything, "pr1", "u1").Return(expectedResp, nil)

	h.Reassign(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp api.ReassignResponse
	err := json.NewDecoder(w.Body).Decode(&resp)
	assert.NoError(t, err)
	assert.Equal(t, expectedResp, &resp)
}

// Остальные кейсы Reassign: BadJSON, ValidationError, NotFound, NoCandidate, PRMerged, NotAssigned, InternalError
func TestPrHandler_Reassign_Errors(t *testing.T) {
	mockService := mocks.NewMockPrService(t)
	h := pr.NewPrHandler(handlers.NewLogger(), mockService)

	tests := []struct {
		name        string
		mockErr     error
		wantStatus  int
		wantErrCode string
	}{
		{"NotFound", repo.ErrNotFound, http.StatusNotFound, api.ErrCodeNotFound},
		{"NoCandidate", repo.ErrNoCandidate, http.StatusConflict, api.ErrCodeNoCandidate},
		{"PRMerged", repo.ErrPRMerged, http.StatusConflict, api.ErrCodePRMerged},
		{"NotAssigned", repo.ErrNotAssigned, http.StatusConflict, api.ErrCodeNotAssigned},
		{"InternalError", errors.New("db error"), http.StatusInternalServerError, api.ErrInternalErr},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqBody := pr.ReassignRequest{PrID: "pr1", OldReviewerID: "u1"}
			body, _ := json.Marshal(reqBody)
			req := httptest.NewRequest(http.MethodPost, "/pr/reassign", bytes.NewReader(body))
			w := httptest.NewRecorder()

			mockService.On("Reassign", mock.Anything, "pr1", "u1").Return(nil, tt.mockErr).Once()

			h.Reassign(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			resp := handlers.DecodeErrorResponse(t, w.Body)
			assert.Equal(t, tt.wantErrCode, resp.Error.Code)
		})
	}
}
