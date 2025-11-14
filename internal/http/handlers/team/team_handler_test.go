package team_test

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
	"avito-intership-2025/internal/http/handlers/team"
	repo "avito-intership-2025/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Add

func TestTeamHandler_Add_Success(t *testing.T) {
	mockService := mocks.NewMockTeamService(t)
	h := team.NewTeamHandler(handlers.NewLogger(), mockService)

	reqBody := team.TeamAddRequest{
		TeamName: "team1",
		Members: []api.TeamMember{
			{UserID: "u1", Username: "User1", IsActive: true},
		},
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/team", bytes.NewReader(body))
	w := httptest.NewRecorder()

	expectedTeam := &api.TeamSchema{
		TeamName: "team1",
		Members:  reqBody.Members,
	}
	mockService.On("Add", mock.Anything, "team1", reqBody.Members).Return(expectedTeam, nil)

	h.Add(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	var resp api.TeamResponse
	err := json.NewDecoder(w.Body).Decode(&resp)
	assert.NoError(t, err)
	assert.Equal(t, *expectedTeam, resp.Team)
}

func TestTeamHandler_Add_BadJSON(t *testing.T) {
	mockService := mocks.NewMockTeamService(t)
	h := team.NewTeamHandler(handlers.NewLogger(), mockService)

	req := httptest.NewRequest(http.MethodPost, "/team", bytes.NewReader([]byte("{invalid json")))
	w := httptest.NewRecorder()

	h.Add(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	resp := handlers.DecodeErrorResponse(t, w.Body)
	assert.Equal(t, api.ErrBadRequest, resp.Error.Code)
}

func TestTeamHandler_Add_ValidationError(t *testing.T) {
	mockService := mocks.NewMockTeamService(t)
	h := team.NewTeamHandler(handlers.NewLogger(), mockService)

	reqBody := team.TeamAddRequest{
		TeamName: "", // trigger validation error
		Members: []api.TeamMember{
			{UserID: "u1", Username: "User1", IsActive: true},
		},
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/team", bytes.NewReader(body))
	w := httptest.NewRecorder()

	h.Add(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp api.ErrorResponse
	err := json.NewDecoder(w.Body).Decode(&resp)
	assert.NoError(t, err)
	assert.Equal(t, api.ErrValidationErr, resp.Error.Code)
}

func TestTeamHandler_Add_TeamExists(t *testing.T) {
	mockService := mocks.NewMockTeamService(t)
	h := team.NewTeamHandler(handlers.NewLogger(), mockService)

	reqBody := team.TeamAddRequest{
		TeamName: "team1",
		Members: []api.TeamMember{
			{UserID: "u1", Username: "User1", IsActive: true},
		},
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/team", bytes.NewReader(body))
	w := httptest.NewRecorder()

	mockService.On("Add", mock.Anything, "team1", reqBody.Members).Return(nil, repo.ErrTeamExists)

	h.Add(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	resp := handlers.DecodeErrorResponse(t, w.Body)
	assert.Equal(t, api.ErrCodeTeamExists, resp.Error.Code)
}

func TestTeamHandler_Add_InternalError(t *testing.T) {
	mockService := mocks.NewMockTeamService(t)
	h := team.NewTeamHandler(handlers.NewLogger(), mockService)

	reqBody := team.TeamAddRequest{
		TeamName: "team1",
		Members: []api.TeamMember{
			{UserID: "u1", Username: "User1", IsActive: true},
		},
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/team", bytes.NewReader(body))
	w := httptest.NewRecorder()

	mockService.On("Add", mock.Anything, "team1", reqBody.Members).Return(nil, errors.New("db error"))

	h.Add(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	resp := handlers.DecodeErrorResponse(t, w.Body)
	assert.Equal(t, api.ErrInternalErr, resp.Error.Code)
}

// Get

func TestTeamHandler_Get_Success(t *testing.T) {
	mockService := mocks.NewMockTeamService(t)
	h := team.NewTeamHandler(handlers.NewLogger(), mockService)

	req := httptest.NewRequest(http.MethodGet, "/team?team_name=team1", nil)
	w := httptest.NewRecorder()

	expectedTeam := &api.TeamSchema{
		TeamName: "team1",
		Members: []api.TeamMember{
			{UserID: "u1", Username: "User1", IsActive: true},
		},
	}
	mockService.On("Get", mock.Anything, "team1").Return(expectedTeam, nil)

	h.Get(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp api.TeamSchema
	err := json.NewDecoder(w.Body).Decode(&resp)
	assert.NoError(t, err)
	assert.Equal(t, *expectedTeam, resp)
}

func TestTeamHandler_Get_MissingTeamName(t *testing.T) {
	mockService := mocks.NewMockTeamService(t)
	h := team.NewTeamHandler(handlers.NewLogger(), mockService)

	req := httptest.NewRequest(http.MethodGet, "/team", nil)
	w := httptest.NewRecorder()

	h.Get(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	resp := handlers.DecodeErrorResponse(t, w.Body)
	assert.Equal(t, api.ErrBadRequest, resp.Error.Code)
}

func TestTeamHandler_Get_NotFound(t *testing.T) {
	mockService := mocks.NewMockTeamService(t)
	h := team.NewTeamHandler(handlers.NewLogger(), mockService)

	req := httptest.NewRequest(http.MethodGet, "/team?team_name=team1", nil)
	w := httptest.NewRecorder()

	mockService.On("Get", mock.Anything, "team1").Return(nil, repo.ErrNotFound)

	h.Get(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	resp := handlers.DecodeErrorResponse(t, w.Body)
	assert.Equal(t, api.ErrCodeNotFound, resp.Error.Code)
}

func TestTeamHandler_Get_InternalError(t *testing.T) {
	mockService := mocks.NewMockTeamService(t)
	h := team.NewTeamHandler(handlers.NewLogger(), mockService)

	req := httptest.NewRequest(http.MethodGet, "/team?team_name=team1", nil)
	w := httptest.NewRecorder()

	mockService.On("Get", mock.Anything, "team1").Return(nil, errors.New("db error"))

	h.Get(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	resp := handlers.DecodeErrorResponse(t, w.Body)
	assert.Equal(t, api.ErrInternalErr, resp.Error.Code)
}
