package utils

import (
    "encoding/json"
    "net/http"
)

type JSONResponse struct {
    Success bool        `json:"success"`
    Message string      `json:"message,omitempty"`
    Data    interface{} `json:"data,omitempty"`
    Error   string      `json:"error,omitempty"`
}

func RespondWithJSON(w http.ResponseWriter, status int, payload interface{}) {
    response, err := json.Marshal(payload)
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(JSONResponse{
            Success: false,
            Error:   "Failed to encode response",
        })
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    w.Write(response)
}

func RespondWithError(w http.ResponseWriter, code int, message string) {
    RespondWithJSON(w, code, JSONResponse{
        Success: false,
        Error:   message,
    })
}

func RespondWithSuccess(w http.ResponseWriter, code int, data interface{}, message string) {
    RespondWithJSON(w, code, JSONResponse{
        Success: true,
        Data:    data,
        Message: message,
    })
}