package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthHanler(t *testing.T) {
	handler := NewHealthHandler()

	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatal(err)
	}

	responseRecorder := httptest.NewRecorder()

	handler.ServeHTTP(responseRecorder, req)

	// Check status code
	if status := responseRecorder.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %vv", status, http.StatusOK)
	}

	// Check content type
	expectedContentType := "application/json"
	if ct := responseRecorder.Header().Get("Content-Type"); ct != expectedContentType {
		t.Errorf("Handler returned wrong content type: got %v want %v", ct, expectedContentType)
	}

	// Check body contains expected fields
	body := responseRecorder.Body.String()
	if !contains(body, "status") {
		t.Error("response body does not container 'status' field")
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 &&
		(s == substr || len(s) >= len(substr) &&
			containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
