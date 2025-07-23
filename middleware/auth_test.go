package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAuthMiddleware(t *testing.T) {
	const secrectKey = "test-secret-key"

	t.Setenv("API_SECRET_KEY", secrectKey)
	testCases := []struct {
		name           string
		apiKeyHeader   string
		expectedStatus int
		expectedBody   string
		isJSON         bool
	}{
		{
			name:           "Success - Correct Key",
			apiKeyHeader:   secrectKey,
			expectedStatus: http.StatusOK,
			expectedBody:   "next called",
			isJSON:         false,
		},
		{
			name:           "Failure - No Key",
			apiKeyHeader:   "incorrect key",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"Unauthorized"}`,
			isJSON:         true,
		},
		{
			name:           "Failure - Without Key",
			apiKeyHeader:   "",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"Unauthorized"}`,
			isJSON:         true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			app := fiber.New()

			app.Use(AuthMiddleware())

			app.Get("/test", func(c *fiber.Ctx) error {
				return c.SendString("next called")
			})

			req := httptest.NewRequest("GET", "/test", nil)
			req.Header.Set("X-API-Key", tc.apiKeyHeader)

			resp, err := app.Test(req)
			assert.NoError(t, err, "app.Test should run no errors")
			defer func() {
				if err := resp.Body.Close(); err != nil {
					log.Printf("Failed to close response body: %v", err)
				}
			}()

			assert.Equal(t, tc.expectedStatus, resp.StatusCode, "The statusCode isn't expected")

			body, err := io.ReadAll(resp.Body)
			assert.NoError(t, err, "Reading the response body should not fail")

			if tc.isJSON {
				assert.JSONEq(t, tc.expectedBody, string(body), "The JSON response body is not what is expected")
			} else {
				assert.Equal(t, tc.expectedBody, string(body), "The text response body is not what is expected")
			}
		})
	}
}
