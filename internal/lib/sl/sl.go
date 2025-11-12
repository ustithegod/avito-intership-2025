package sl

import "log/slog"

// Err позволяет передавать в атрибуты slog-логов ошибку как она есть (error type)
func Err(err error) slog.Attr {
	return slog.Attr{
		Key:   "error",
		Value: slog.StringValue(err.Error()),
	}
}
