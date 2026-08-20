package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

type APIResponse struct {
	Success bool            `json:"success"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
	Error   string          `json:"error"`
}

func JSONRequest(
	t testing.TB,
	router http.Handler,
	method string,
	path string,
	body interface{},
	token string,
	headers map[string]string,
) *httptest.ResponseRecorder {
	t.Helper()

	var payload []byte
	var err error
	if body != nil {
		payload, err = json.Marshal(body)
		if err != nil {
			t.Fatalf("json.Marshal() error = %v", err)
		}
	}

	req := httptest.NewRequest(method, path, bytes.NewReader(payload))
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	return recorder
}

func DecodeAPIResponse(t testing.TB, recorder *httptest.ResponseRecorder) APIResponse {
	t.Helper()

	var response APIResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("json.Unmarshal(APIResponse) error = %v\nbody=%s", err, recorder.Body.String())
	}

	return response
}

func DecodeData(t testing.TB, raw json.RawMessage, target interface{}) {
	t.Helper()

	if err := json.Unmarshal(raw, target); err != nil {
		t.Fatalf("json.Unmarshal(data) error = %v", err)
	}
}
