package apperrors

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
)

func ErrorHandler(w http.ResponseWriter, req *http.Request, err error) {
	var appErr *Error
	if !errors.As(err, &appErr) {
		appErr = Unknown.Wrap(err, "internal process failed").(*Error)
	}

	statusCode := statusCodeFor(ErrCode(appErr.ErrCode))
	log.Printf("error occurred: code=%s method=%s path=%s status=%d message=%s cause=%v", appErr.ErrCode, req.Method, req.URL.Path, statusCode, appErr.Message, appErr.Err)

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(appErr)
}

func statusCodeFor(code ErrCode) int {
	switch code {
	case BadParam, ReqBodyDecodeFailed, RequestBodyTooLarge:
		return http.StatusBadRequest
	case NotFound:
		return http.StatusNotFound
	case DependencyUnavailable:
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}
