package krouter

import (
	"net/http"
)

// @Author KHighness
// @Update 2022-11-09

// GetParam returns route param stored in http.Request.
func GetParam(r *http.Request, key string) string {
	return GetAllParams(r)[key]
}

// contextKeyType defines a type which is used for
// storing values in context.Context.
type contextKeyType struct {
}

// contextKey is the key which is used to store values
// in context for each http.Request.
var contextKey = contextKeyType{}

// paramsMapType defines a type which is used to
// store route params.
type paramsMapType map[string]string

// GetAllParams returns all route params sored in http.Request.
func GetAllParams(r *http.Request) paramsMapType {
	if values, ok := r.Context().Value(contextKey).(paramsMapType); ok {
		return values
	}

	return nil
}
