package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"strings"

	"github.com/iansebalt/sezzle-calculator/internal/apierror"
)

// Requests carry three short fields; anything larger is not a calculation.
const maxBodyBytes = 4 << 10

// Pointers distinguish "absent" from "sent as zero", which the zero value alone cannot express.
type calculateRequest struct {
	Operation *string  `json:"operation"`
	A         *float64 `json:"a"`
	B         *float64 `json:"b"`
}

type calculateResponse struct {
	Operation string   `json:"operation"`
	A         float64  `json:"a"`
	B         *float64 `json:"b,omitempty"`
	Result    float64  `json:"result"`
}

func decodeCalculateRequest(w http.ResponseWriter, r *http.Request) (calculateRequest, *apierror.Error) {
	if apiErr := requireJSONContentType(r); apiErr != nil {
		return calculateRequest{}, apiErr
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	var req calculateRequest
	if err := decoder.Decode(&req); err != nil {
		return calculateRequest{}, decodeError(err)
	}
	if decoder.More() {
		return calculateRequest{}, apierror.MalformedJSON()
	}
	if req.Operation == nil {
		return calculateRequest{}, apierror.MissingOperation()
	}

	return req, nil
}

func requireJSONContentType(r *http.Request) *apierror.Error {
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/json" {
		return apierror.UnsupportedMediaType()
	}
	return nil
}

func decodeError(err error) *apierror.Error {
	var maxBytesErr *http.MaxBytesError
	var typeErr *json.UnmarshalTypeError
	var syntaxErr *json.SyntaxError

	switch {
	case errors.As(err, &maxBytesErr):
		return apierror.PayloadTooLarge()

	case errors.As(err, &typeErr):
		switch typeErr.Field {
		// An empty field means the mismatch is the body itself, not one of its members.
		case "":
			return apierror.MalformedJSON()
		case "operation":
			return apierror.InvalidOperation()
		default:
			return apierror.InvalidOperand(typeErr.Field)
		}

	case errors.As(err, &syntaxErr),
		errors.Is(err, io.EOF),
		errors.Is(err, io.ErrUnexpectedEOF):
		return apierror.MalformedJSON()

	// encoding/json reports an unknown field as a bare string with no typed error to match on.
	case strings.HasPrefix(err.Error(), "json: unknown field "):
		return apierror.UnknownField()

	default:
		return apierror.MalformedJSON()
	}
}
