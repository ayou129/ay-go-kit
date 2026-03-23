package cryptox

import (
	"testing"
)

func TestHashPassword_ReturnsNonEmpty(t *testing.T) {
	hash, err := HashPassword("test123")
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}
	if hash == "" {
		t.Error("HashPassword returned empty string")
	}
}

func TestHashPassword_DifferentEachTime(t *testing.T) {
	h1, _ := HashPassword("test123")
	h2, _ := HashPassword("test123")
	if h1 == h2 {
		t.Error("HashPassword should return different hashes due to random salt")
	}
}

func TestCheckPassword_CorrectPassword(t *testing.T) {
	hash, _ := HashPassword("correct")
	if !CheckPassword("correct", hash) {
		t.Error("CheckPassword should return true for correct password")
	}
}

func TestCheckPassword_WrongPassword(t *testing.T) {
	hash, _ := HashPassword("correct")
	if CheckPassword("wrong", hash) {
		t.Error("CheckPassword should return false for wrong password")
	}
}

func TestCheckPassword_EmptyPassword(t *testing.T) {
	hash, _ := HashPassword("")
	if !CheckPassword("", hash) {
		t.Error("CheckPassword should handle empty password")
	}
}
