package httphandlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/terabiome/terabiome/internal/contracts"
	"go.yaml.in/yaml/v4"
)

var globalValidator = validator.New()

type ContentType string

const (
	ContentTypeJSON ContentType = "application/json"
	ContentTypeYAML ContentType = "application/yaml"
)

// GenericResponse is a standard API response structure
type GenericResponse struct {
	Message string `json:"message"`
	Body    any    `json:"body,omitempty"`
	Error   error  `json:"error,omitempty"`
}

// responseCallback is a function type for error handling callbacks
type responseCallback func()

// ParseBodyAndHandleError parses the request body and handles errors
func ParseBodyAndHandleError(writer http.ResponseWriter, request *http.Request, target any, requireBody bool) (responseCallback, error) {
	var err error
	if requireBody {
		// decode and check
		acceptType := request.Header.Get("Accept")
		switch acceptType {
		case string(ContentTypeJSON):
			err = json.NewDecoder(request.Body).Decode(target)
		case string(ContentTypeYAML):
			err = yaml.NewDecoder(request.Body).Decode(target)
		default:
			err = fmt.Errorf("unsupported HTTP header Accept: %s", acceptType)
		}
		// read validator tag
		if v, ok := target.(contracts.Validatable); err == nil && ok {
			err = v.Validate(globalValidator)
		}
		if err != nil {
			return func() {
				WriteResult(writer, http.StatusBadRequest, ContentTypeJSON, GenericResponse{
					Body:    nil,
					Message: "invalid request body",
					Error:   err,
				})
			}, err
		}
	}
	return func() {}, nil
}

// WriteResult writes a JSON response with the given status code
func WriteResult(writer http.ResponseWriter, statusCode int, contentType ContentType, response GenericResponse) {
	writer.Header().Set("Content-Type", string(contentType))
	writer.WriteHeader(statusCode)
	json.NewEncoder(writer).Encode(response)
}

// WriteBytes writes raw bytes with the given status code
func WriteBytes(writer http.ResponseWriter, statusCode int, contentType ContentType, data []byte) {
	writer.Header().Set("Content-Type", string(contentType))
	writer.WriteHeader(statusCode)
	writer.Write(data)
}
