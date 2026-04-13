package response

import (
    "encoding/json"
    "net/http"
)

type APIResponse struct {
    Success bool        `json:"success"`
    Message string      `json:"message,omitempty"`
    Data    interface{} `json:"data,omitempty"`
    Error   string      `json:"error,omitempty"`
}

func JSON(w http.ResponseWriter, status int, data interface{}) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(data)
}

func Success(w http.ResponseWriter, data interface{}) {
    JSON(w, http.StatusOK, APIResponse{Success: true, Data: data})
}

func Created(w http.ResponseWriter, data interface{}) {
    JSON(w, http.StatusCreated, APIResponse{Success: true, Data: data})
}

func Error(w http.ResponseWriter, status int, msg string) {
    JSON(w, status, APIResponse{Success: false, Error: msg})
}

func Unauthorized(w http.ResponseWriter) {
    Error(w, http.StatusUnauthorized, "Unauthorized")
}