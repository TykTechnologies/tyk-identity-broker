package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleAPIOK(t *testing.T) {
	tests := []struct {
		name     string
		data     interface{}
		id       string
		code     int
		wantID   string
		wantCode int
	}{
		{
			name:     "200 with data and id",
			data:     map[string]string{"key": "value"},
			id:       "profile-1",
			code:     http.StatusOK,
			wantID:   "profile-1",
			wantCode: http.StatusOK,
		},
		{
			name:     "201 created with empty id",
			data:     nil,
			id:       "",
			code:     http.StatusCreated,
			wantID:   "",
			wantCode: http.StatusCreated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()

			HandleAPIOK(tt.data, tt.id, tt.code, w, nil)

			assert.Equal(t, tt.wantCode, w.Code)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

			var got APIOKMessage
			require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
			assert.Equal(t, "ok", got.Status)
			assert.Equal(t, tt.wantID, got.ID)
		})
	}
}
