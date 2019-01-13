package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// RespondwithJSON write json response format
func RespondwithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	fmt.Println(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

// RespondWithError return error message
func RespondWithError(w http.ResponseWriter, code int, msg string) {
	RespondwithJSON(w, code, map[string]string{"message": msg})
}

// UnmarshalJSON Function
func UnmarshalJSON(data []byte, target interface{}) error {
	err := json.Unmarshal(data, &target)

	if err != nil {
		return err
	}

	return nil
}
