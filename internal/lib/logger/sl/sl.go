package sl

import "log/slog"

func Err(err error) slog.Attr { // helper to log errors
	return slog.Attr{
		Key:   "error",
		Value: slog.StringValue(err.Error()),
	}
}
