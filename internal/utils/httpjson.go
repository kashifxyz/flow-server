package utils

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

const MaxJSONBody = 64 << 10

type ErrorPayload struct {
	Error ErrorDetail `json:"error"`
}

type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func JSONError(w http.ResponseWriter, status int, code, message string) {
	WriteJSON(w, status, ErrorPayload{Error: ErrorDetail{Code: code, Message: message}})
}

func NotImplemented(w http.ResponseWriter, _ *http.Request) {
	JSONError(w, http.StatusNotImplemented, "not_implemented", "Not implemented")
}

func DecodeJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, MaxJSONBody)
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(dst); err != nil {
		if errors.Is(err, io.EOF) {
			return err
		}
		return err
	}
	return nil
}
