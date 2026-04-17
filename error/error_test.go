package error

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleError(t *testing.T) {
	tests := []struct {
		name       string
		tag        string
		errorMsg   string
		rawErr     error
		code       int
		wantCode   int
		wantStatus string
		wantError  string
	}{
		{
			name:       "404 not found",
			tag:        "TEST",
			errorMsg:   "resource not found",
			rawErr:     errors.New("key missing"),
			code:       http.StatusNotFound,
			wantCode:   http.StatusNotFound,
			wantStatus: "error",
			wantError:  "resource not found",
		},
		{
			name:       "400 bad request",
			tag:        "TEST",
			errorMsg:   "invalid input",
			rawErr:     errors.New("validation failed"),
			code:       http.StatusBadRequest,
			wantCode:   http.StatusBadRequest,
			wantStatus: "error",
			wantError:  "invalid input",
		},
		{
			name:       "500 internal server error",
			tag:        "TEST",
			errorMsg:   "something went wrong",
			rawErr:     errors.New("internal error"),
			code:       http.StatusInternalServerError,
			wantCode:   http.StatusInternalServerError,
			wantStatus: "error",
			wantError:  "something went wrong",
		},
		{
			name:       "nil raw error",
			tag:        "TEST",
			errorMsg:   "no raw error",
			rawErr:     nil,
			code:       http.StatusUnauthorized,
			wantCode:   http.StatusUnauthorized,
			wantStatus: "error",
			wantError:  "no raw error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()

			HandleError(tt.tag, tt.errorMsg, tt.rawErr, tt.code, w, nil)

			assert.Equal(t, tt.wantCode, w.Code)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

			var got APIErrorMessage
			require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
			assert.Equal(t, tt.wantStatus, got.Status)
			assert.Equal(t, tt.wantError, got.Error)
		})
	}
}

func TestHandleError_ResponseBody(t *testing.T) {
	w := httptest.NewRecorder()

	HandleError("TAG", "some error", errors.New("raw"), http.StatusBadRequest, w, nil)

	body := w.Body.Bytes()
	assert.True(t, json.Valid(body), "response body should be valid JSON")
}
