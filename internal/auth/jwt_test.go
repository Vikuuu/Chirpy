package auth

import (
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

const (
	validSecret   = "mySecret"
	invalidSecret = "wrongSecret"
	validExpiry   = time.Minute * 5
	expiredExpiry = -time.Minute * 5
)

// TestMakeJWT tests the creation of a valid JWT token
func TestMakeJWT(t *testing.T) {
	userID := uuid.New()

	token, err := MakeJWT(userID, validSecret, validExpiry)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
}

// TestValidateJWT tests the validation of a correctly signed JWT token
func TestValidateJWT(t *testing.T) {
	userID := uuid.New()

	token, err := MakeJWT(userID, validSecret, validExpiry)
	assert.NoError(t, err)

	parsedID, err := ValidateJWT(token, validSecret)
	assert.NoError(t, err)
	assert.Equal(t, userID, parsedID)
}

// TestExpiredJWT tests the rejection of an expired JWT token
func TestExpiredJWT(t *testing.T) {
	userID := uuid.New()

	token, err := MakeJWT(userID, validSecret, expiredExpiry)
	assert.NoError(t, err)

	_, err = ValidateJWT(token, validSecret)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "token is expired") // customize error message check as needed
}

// TestJWTWithWrongSecret tests that a JWT signed with the wrong secret is rejected
func TestJWTWithWrongSecret(t *testing.T) {
	userID := uuid.New()

	token, err := MakeJWT(userID, validSecret, validExpiry)
	assert.NoError(t, err)

	_, err = ValidateJWT(token, invalidSecret)
	assert.Error(t, err)
	assert.Contains(
		t,
		err.Error(),
		"signature is invalid",
	) // customize error message check as needed
}

func TestGetBearerToken(t *testing.T) {
	tests := []struct {
		name          string
		header        http.Header
		expectedToken string
		expectError   bool
		errorMessage  string
	}{
		{
			name:          "No Authorization Header",
			header:        http.Header{},
			expectedToken: "",
			expectError:   true,
			errorMessage:  "No Bearer token provided",
		},
		{
			name: "Empty Authorization Header",
			header: http.Header{
				"Authorization": []string{""},
			},
			expectedToken: "",
			expectError:   true,
			errorMessage:  "No Bearer token provided",
		},
		{
			name: "Authorization Header without Bearer",
			header: http.Header{
				"Authorization": []string{"Basic abc123"},
			},
			expectedToken: "",
			expectError:   true,
			errorMessage:  "No Bearer token provided",
		},
		{
			name: "Valid Bearer Token",
			header: http.Header{
				"Authorization": []string{"Bearer abc123xyz"},
			},
			expectedToken: "abc123xyz",
			expectError:   false,
		},
		{
			name: "Bearer Token with Extra Spaces",
			header: http.Header{
				"Authorization": []string{"Bearer    abc123xyz   "},
			},
			expectedToken: "abc123xyz",
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := GetBearerToken(tt.header)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMessage)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedToken, token)
			}
		})
	}
}
