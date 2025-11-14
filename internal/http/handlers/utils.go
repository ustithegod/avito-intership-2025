package handlers

import (
	"avito-intership-2025/internal/http/api"
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

func NewLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func DecodeErrorResponse(t *testing.T, body *bytes.Buffer) api.ErrorResponse {
	var resp api.ErrorResponse
	err := json.NewDecoder(body).Decode(&resp)
	assert.NoError(t, err)
	return resp
}
