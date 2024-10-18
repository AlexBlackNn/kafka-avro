package dto

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
	jsoniter "github.com/json-iterator/go"
)

// Input

type UserLoyalty struct {
	UUID      string `json:"uuid" validate:"uuid"`
	Operation string `json:"operation" validate:"required"`
	Comment   string `json:"comment" validate:"required"`
	Balance   int    `json:"balance" validate:"required"`
}

// Output

type Response struct {
	Status  string `json:"status"`
	Error   string `json:"error,omitempty"`
	UUID    string `json:"uuid,omitempty"`
	Balance int    `json:"balance,omitempty"`
}

const StatusError = "Error"
const StatusSuccess = "Success"

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func sendJSON(w http.ResponseWriter, statusCode int, data []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(data)
}

// Errors.

func ResponseErrorNotFound(
	w http.ResponseWriter,
	message string,
) {
	dataMarshal, _ := json.Marshal(Response{
		Status: StatusError,
		Error:  message,
	})
	sendJSON(w, http.StatusNotFound, dataMarshal)
}

func ResponseErrorInternal(
	w http.ResponseWriter,
	message string,
) {
	dataMarshal, _ := json.Marshal(Response{
		Status: StatusError,
		Error:  message,
	})
	sendJSON(w, http.StatusInternalServerError, dataMarshal)
}

func ResponseErrorNowAllowed(
	w http.ResponseWriter,
	message string,
) {
	dataMarshal, _ := json.Marshal(Response{
		Status: StatusError,
		Error:  message,
	})
	sendJSON(w, http.StatusMethodNotAllowed, dataMarshal)
}

func ResponseErrorStatusConflict(
	w http.ResponseWriter,
	message string,
) {
	dataMarshal, _ := json.Marshal(Response{
		Status: StatusError,
		Error:  message,
	})
	sendJSON(w, http.StatusMethodNotAllowed, dataMarshal)
}

func ResponseErrorBadRequest(
	w http.ResponseWriter,
	message string,
) {
	dataMarshal, _ := json.Marshal(Response{
		Status: StatusError,
		Error:  message,
	})
	sendJSON(w, http.StatusBadRequest, dataMarshal)
}

// OK.

func ResponseOK(w http.ResponseWriter) {
	dataMarshal, _ := json.Marshal(
		Response{
			Status: StatusSuccess,
		},
	)
	sendJSON(w, http.StatusOK, dataMarshal)
}

func ResponseOKLoyalty(
	w http.ResponseWriter,
	uuid string,
	value int,
) {
	dataMarshal, _ := json.Marshal(
		Response{
			Status:  StatusSuccess,
			UUID:    uuid,
			Balance: value,
		},
	)
	sendJSON(w, http.StatusOK, dataMarshal)
}

// Validation error.

func ValidationError(errs validator.ValidationErrors) string {
	var errMsgs []string

	for _, err := range errs {
		switch err.ActualTag() {
		case "required":
			errMsgs = append(
				errMsgs, fmt.Sprintf("field %s is a required field", err.Field()),
			)
		default:
			errMsgs = append(
				errMsgs, fmt.Sprintf("field %s is not valid", err.Field()),
			)
		}
	}
	return strings.Join(errMsgs, ", ")
}
