package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"restaurant/pkg/models"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var baseURL = "http://localhost:8080"

func doPost(t *testing.T, path string, body interface{}, cookie *http.Cookie) *http.Response {
	t.Helper()

	var reqBody io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		require.NoError(t, err, "failed to marshal request body")
		reqBody = bytes.NewBuffer(b)
	}

	req, err := http.NewRequest(http.MethodPost, baseURL+path, reqBody)
	require.NoError(t, err, "failed to create request")

	req.Header.Set("Content-Type", "application/json")
	if cookie != nil {
		req.AddCookie(cookie)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	require.NoError(t, err, "request failed")

	return resp
}

func readJSON(t *testing.T, resp *http.Response, v interface{}) {
	t.Helper()
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err, "failed to read response body")

	err = json.Unmarshal(body, v)
	require.NoError(t, err, "failed to unmarshal response: %s", string(body))
}

func extractCookie(t *testing.T, resp *http.Response, name string) *http.Cookie {
	t.Helper()

	for _, c := range resp.Cookies() {
		if c.Name == name {
			return c
		}
	}
	t.Fatalf("cookie %q not found", name)
	return nil
}

func TestUserFlow_HappyPath(t *testing.T) {
	username := fmt.Sprintf("testuser_%s", uuid.New().String()[:8])
	password := uuid.New().String()[:16]
	var refreshCookie *http.Cookie

	t.Run("Register", func(t *testing.T) {
		reqBody := models.RegisterRequest{Name: username, Password: password}
		resp := doPost(t, "/register", reqBody, nil)

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var got models.RegisterResponse
		readJSON(t, resp, &got)

		assert.NotEmpty(t, got.UserID)
	})

	t.Run("Login", func(t *testing.T) {
		reqBody := models.LoginRequest{Name: username, Password: password}
		resp := doPost(t, "/login", reqBody, nil)

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var got models.LoginResponse
		readJSON(t, resp, &got)

		assert.NotEmpty(t, got.AccessToken)
		assert.NotEmpty(t, got.RefreshToken)
		assert.NotEqual(t, got.AccessToken, got.RefreshToken)

		refreshCookie = extractCookie(t, resp, "refresh_token")
	})

	t.Run("Refresh", func(t *testing.T) {
		require.NotNil(t, refreshCookie)

		resp := doPost(t, "/refresh", nil, refreshCookie)

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var got models.RefreshResponse
		readJSON(t, resp, &got)

		assert.NotEmpty(t, got.AccessToken)

		refreshCookie = extractCookie(t, resp, "refresh_token")
	})

	t.Run("Logout", func(t *testing.T) {
		require.NotNil(t, refreshCookie)

		resp := doPost(t, "/logout", nil, refreshCookie)

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}

func TestUserFlow_Errors(t *testing.T) {
	username := fmt.Sprintf("testuser_%s", uuid.New().String()[:8])
	password := uuid.New().String()[:16]
	t.Run("Register with weak password", func(t *testing.T) {
		reqBody := models.RegisterRequest{Name: username, Password: "1"}
		resp := doPost(t, "/register", reqBody, nil)

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("Register with short name", func(t *testing.T) {
		reqBody := models.RegisterRequest{Name: "ab", Password: password}
		resp := doPost(t, "/register", reqBody, nil)

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	reqBody := models.RegisterRequest{Name: username, Password: password}
	resp := doPost(t, "/register", reqBody, nil)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	t.Run("Register duplicate user", func(t *testing.T) {
		reqBody := models.RegisterRequest{Name: username, Password: password}
		resp := doPost(t, "/register", reqBody, nil)

		assert.Equal(t, http.StatusConflict, resp.StatusCode)
	})

	t.Run("Login with wrong password", func(t *testing.T) {
		reqBody := models.LoginRequest{Name: username, Password: "wrongpassword"}
		resp := doPost(t, "/login", reqBody, nil)

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("Login with non-existent user", func(t *testing.T) {
		reqBody := models.LoginRequest{Name: "nonexistent_user", Password: password}
		resp := doPost(t, "/login", reqBody, nil)

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	loginBody := models.LoginRequest{Name: username, Password: password}
	loginResp := doPost(t, "/login", loginBody, nil)
	require.Equal(t, http.StatusOK, loginResp.StatusCode)
	refreshCookie := extractCookie(t, loginResp, "refresh_token")

	t.Run("Refresh without cookie", func(t *testing.T) {
		resp := doPost(t, "/refresh", nil, nil)

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("Refresh with fake token", func(t *testing.T) {
		fakeCookie := &http.Cookie{Name: "refresh_token", Value: "fake.token.here"}
		resp := doPost(t, "/refresh", nil, fakeCookie)

		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})

	t.Run("Refresh with empty token", func(t *testing.T) {
		emptyCookie := &http.Cookie{Name: "refresh_token", Value: ""}
		resp := doPost(t, "/refresh", nil, emptyCookie)

		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})

	t.Run("Logout without cookie", func(t *testing.T) {
		resp := doPost(t, "/logout", nil, nil)

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("Logout with fake token", func(t *testing.T) {
		fakeCookie := &http.Cookie{Name: "refresh_token", Value: "fake.token.here"}
		resp := doPost(t, "/logout", nil, fakeCookie)

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("Logout valid", func(t *testing.T) {
		resp := doPost(t, "/logout", nil, refreshCookie)

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}
