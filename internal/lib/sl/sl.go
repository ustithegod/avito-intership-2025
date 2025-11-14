package sl

import (
	"avito-intership-2025/internal/http/api"
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Err позволяет передавать в атрибуты slog-логов ошибку как она есть (error type)
func Err(err error) slog.Attr {
	return slog.Attr{
		Key:   "error",
		Value: slog.StringValue(err.Error()),
	}
}

func NewLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func DecodeErrorResponse(t *testing.T, body *bytes.Buffer) api.ErrorResponse {
	var resp api.ErrorResponse
	err := json.NewDecoder(body).Decode(&resp)
	assert.NoError(t, err)
	return resp
}
