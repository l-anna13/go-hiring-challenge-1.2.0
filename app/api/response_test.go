package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// unmarshalable is a helper struct containing a channel, which JSON cannot marshal.
type unmarshalable struct {
	C chan int
}

func TestRespondWithJSON(t *testing.T) {
	tests := []struct {
		name               string
		code               int
		payload            interface{}
		expectMarshalError bool
		expectedBody       string
		expectedStatus     int
	}{
		{
			name: "Success_StandardData_201",
			code: http.StatusCreated,
			payload: struct {
				ID      int
				Message string
			}{ID: 1, Message: "created"},
			expectMarshalError: false,
			expectedBody:       `{"ID":1,"Message":"created"}`,
			expectedStatus:     http.StatusCreated,
		},
		{
			name:               "Success_EmptyPayload_200",
			code:               http.StatusOK,
			payload:            nil,
			expectMarshalError: false,
			expectedBody:       `null`,
			expectedStatus:     http.StatusOK,
		},
		{
			name:               "Failure_MarshalError_500",
			code:               http.StatusOK, // The intended code is irrelevant as it's overwritten by the 500
			payload:            unmarshalable{C: make(chan int)},
			expectMarshalError: true,
			expectedBody:       "Internal server error formatting response.",
			expectedStatus:     http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()

			// Call the function under test
			RespondWithJSON(rr, tt.code, tt.payload)

			// 1. Check Status Code
			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status code %d, got %d. Body: %s", tt.expectedStatus, rr.Code, rr.Body.String())
			}

			// 2. Check Content-Type Header: Only successful responses should have application/json.
			if !tt.expectMarshalError {
				if rr.Header().Get("Content-Type") != "application/json" {
					t.Errorf("Expected Content-Type 'application/json', got '%s'", rr.Header().Get("Content-Type"))
				}
			} else {
				// Error path uses http.Error, which sets "text/plain; charset=utf-8"
				if rr.Header().Get("Content-Type") != "text/plain; charset=utf-8" {
					t.Errorf("Expected error Content-Type 'text/plain; charset=utf-8', got '%s'", rr.Header().Get("Content-Type"))
				}
			}

			// 3. Check Body Content
			body := strings.TrimSpace(rr.Body.String())
			if !strings.Contains(body, tt.expectedBody) {
				t.Errorf("Expected body to contain '%s', but got '%s'", tt.expectedBody, body)
			}
		})
	}
}
