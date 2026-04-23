package auth_middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestValidateJIDFieldsNormalizesArrayFields(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.POST(
		"/group/participant",
		NewJIDValidationMiddleware().ValidateJIDFields("number", "participants"),
		func(c *gin.Context) {
			var payload struct {
				Number       string   `json:"number"`
				Participants []string `json:"participants"`
			}
			if err := c.ShouldBindJSON(&payload); err != nil {
				t.Fatalf("expected JSON to bind after middleware, got error: %v", err)
			}
			c.JSON(http.StatusOK, payload)
		},
	)

	req := httptest.NewRequest(http.MethodPost, "/group/participant", strings.NewReader(`{
		"number": "120363426688645359",
		"participants": ["554998272782", "554998272783@s.whatsapp.net"]
	}`))
	req.Header.Set("Content-Type", "application/json")

	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d with body %s", res.Code, res.Body.String())
	}

	var payload struct {
		Number       string   `json:"number"`
		Participants []string `json:"participants"`
	}
	if err := json.Unmarshal(res.Body.Bytes(), &payload); err != nil {
		t.Fatalf("expected JSON response, got error: %v", err)
	}

	if payload.Number != "120363426688645359@g.us" {
		t.Fatalf("expected normalized group JID, got %q", payload.Number)
	}
	if len(payload.Participants) != 2 {
		t.Fatalf("expected 2 participants, got %d", len(payload.Participants))
	}
	if payload.Participants[0] != "554998272782@s.whatsapp.net" {
		t.Fatalf("expected first participant normalized, got %q", payload.Participants[0])
	}
	if payload.Participants[1] != "554998272783@s.whatsapp.net" {
		t.Fatalf("expected second participant preserved, got %q", payload.Participants[1])
	}
}

func TestValidateJIDFieldsRejectsEmptyArrayFields(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.POST(
		"/group/participant",
		NewJIDValidationMiddleware().ValidateJIDFields("participants"),
		func(c *gin.Context) {
			c.Status(http.StatusOK)
		},
	)

	req := httptest.NewRequest(http.MethodPost, "/group/participant", strings.NewReader(`{
		"participants": []
	}`))
	req.Header.Set("Content-Type", "application/json")

	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	if res.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d with body %s", res.Code, res.Body.String())
	}
	if !strings.Contains(res.Body.String(), "participants is required and cannot be empty") {
		t.Fatalf("expected empty participants error, got %s", res.Body.String())
	}
}
