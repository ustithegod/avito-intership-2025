package user_test

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
	"avito-intership-2025/internal/http/handlers/user"
	repo "avito-intership-2025/internal/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// SetIsActive

func TestUserHandler_SetIsActive_Success(t *testing.T) {
	mockService := mocks.NewMockUserService(t)
	h := user.NewUserHandler(handlers.NewLogger(), mockService)

	reqBody := user.SetIsActiveRequest{
		UserID:   "u1",
		IsActive: true,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/user/set_active", bytes.NewReader(body))
	w := httptest.NewRecorder()

	expectedUser := &api.UserSchema{
		UserID:   "u1",
		Username: "User1",
		TeamName: "team1",
		IsActive: true,
	}
	mockService.On("SetIsActive", mock.Anything, "u1", true).Return(expectedUser, nil)

	h.SetIsActive(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp api.UserResponse
	err := json.NewDecoder(w.Body).Decode(&resp)
	assert.NoError(t, err)
	assert.Equal(t, *expectedUser, resp.User)
}

func TestUserHandler_SetIsActive_BadJSON(t *testing.T) {
	mockService := mocks.NewMockUserService(t)
	h := user.NewUserHandler(handlers.NewLogger(), mockService)

	req := httptest.NewRequest(http.MethodPost, "/user/set_active", bytes.NewReader([]byte("{invalid json")))
	w := httptest.NewRecorder()

	h.SetIsActive(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	resp := handlers.DecodeErrorResponse(t, w.Body)
	assert.Equal(t, api.ErrBadRequest, resp.Error.Code)
}

func TestUserHandler_SetIsActive_ValidationError(t *testing.T) {
	mockService := mocks.NewMockUserService(t)
	h := user.NewUserHandler(handlers.NewLogger(), mockService)

	reqBody := user.SetIsActiveRequest{
		UserID:   "", // missing required field
		IsActive: true,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/user/set_active", bytes.NewReader(body))
	w := httptest.NewRecorder()

	h.SetIsActive(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp api.ErrorResponse
	err := json.NewDecoder(w.Body).Decode(&resp)
	assert.NoError(t, err)
	assert.Equal(t, api.ErrValidationErr, resp.Error.Code)
}

func TestUserHandler_SetIsActive_NotFound(t *testing.T) {
	mockService := mocks.NewMockUserService(t)
	h := user.NewUserHandler(handlers.NewLogger(), mockService)

	reqBody := user.SetIsActiveRequest{
		UserID:   "u1",
		IsActive: true,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/user/set_active", bytes.NewReader(body))
	w := httptest.NewRecorder()

	mockService.On("SetIsActive", mock.Anything, "u1", true).Return(nil, repo.ErrNotFound)

	h.SetIsActive(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	resp := handlers.DecodeErrorResponse(t, w.Body)
	assert.Equal(t, api.ErrCodeNotFound, resp.Error.Code)
}

func TestUserHandler_SetIsActive_InternalError(t *testing.T) {
	mockService := mocks.NewMockUserService(t)
	h := user.NewUserHandler(handlers.NewLogger(), mockService)

	reqBody := user.SetIsActiveRequest{
		UserID:   "u1",
		IsActive: true,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/user/set_active", bytes.NewReader(body))
	w := httptest.NewRecorder()

	mockService.On("SetIsActive", mock.Anything, "u1", true).Return(nil, errors.New("db error"))

	h.SetIsActive(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	resp := handlers.DecodeErrorResponse(t, w.Body)
	assert.Equal(t, api.ErrInternalErr, resp.Error.Code)
}

// GetReview

func TestUserHandler_GetReview_Success(t *testing.T) {
	mockService := mocks.NewMockUserService(t)
	h := user.NewUserHandler(handlers.NewLogger(), mockService)

	req := httptest.NewRequest(http.MethodGet, "/user/get_review?team_name=u1", nil)
	w := httptest.NewRecorder()

	expectedReview := &api.GetReviewResponse{
		UserID: "u1",
		PullRequests: []api.PullRequestShort{
			{ID: "pr1", Name: "PR 1", AuthorID: "u1", Status: "open"},
		},
	}
	mockService.On("GetReview", mock.Anything, "u1").Return(expectedReview, nil)

	h.GetReview(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp api.GetReviewResponse
	err := json.NewDecoder(w.Body).Decode(&resp)
	assert.NoError(t, err)
	assert.Equal(t, expectedReview, &resp)
}

func TestUserHandler_GetReview_MissingUserID(t *testing.T) {
	mockService := mocks.NewMockUserService(t)
	h := user.NewUserHandler(handlers.NewLogger(), mockService)

	req := httptest.NewRequest(http.MethodGet, "/user/get_review", nil)
	w := httptest.NewRecorder()

	h.GetReview(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	resp := handlers.DecodeErrorResponse(t, w.Body)
	assert.Equal(t, api.ErrBadRequest, resp.Error.Code)
}

func TestUserHandler_GetReview_NotFound(t *testing.T) {
	mockService := mocks.NewMockUserService(t)
	h := user.NewUserHandler(handlers.NewLogger(), mockService)

	req := httptest.NewRequest(http.MethodGet, "/user/get_review?team_name=u1", nil)
	w := httptest.NewRecorder()

	mockService.On("GetReview", mock.Anything, "u1").Return(nil, repo.ErrNotFound)

	h.GetReview(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	resp := handlers.DecodeErrorResponse(t, w.Body)
	assert.Equal(t, api.ErrCodeNotFound, resp.Error.Code)
}

func TestUserHandler_GetReview_InternalError(t *testing.T) {
	mockService := mocks.NewMockUserService(t)
	h := user.NewUserHandler(handlers.NewLogger(), mockService)

	req := httptest.NewRequest(http.MethodGet, "/user/get_review?team_name=u1", nil)
	w := httptest.NewRecorder()

	mockService.On("GetReview", mock.Anything, "u1").Return(nil, errors.New("db error"))

	h.GetReview(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	resp := handlers.DecodeErrorResponse(t, w.Body)
	assert.Equal(t, api.ErrInternalErr, resp.Error.Code)
}
