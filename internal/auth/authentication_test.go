package auth

import (
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestHashPassword(t *testing.T) {
	password := "mysecurepassword"

	// Call the HashPassword function
	funcHashPassword, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword returned an error: %v", err)
	}

	// Check that the hashed password is non-empty
	if funcHashPassword == "" {
		t.Fatal("Expected hashed password to be non-empty")
	}

	// Verify the hash matches the password
	err = bcrypt.CompareHashAndPassword([]byte(funcHashPassword), []byte(password))
	if err != nil {
		t.Fatal("The hashed password does not match the original password")
	}
}

func TestCheckPasswordHash(t *testing.T) {
	password := "mysecurepassword"

	// Generate a hash using bcrypt directly
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("Error generating bcrypt hash: %v", err)
	}

	// Call the CheckPasswordHash function to verify the password against the hash
	err = CheckPasswordHash(password, string(hash))
	if err != nil {
		t.Fatal("CheckPasswordHash returned an error, the password does not match the hash")
	}
}
