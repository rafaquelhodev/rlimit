package main

import (
	"testing"

	"github.com/rafaquelhodev/rlimit/internal/tokens"
)

func TestExecuteINITRequest(t *testing.T) {
	tokenManager := tokens.NewUserTokenBucketManager()

	got, err := executeRequest(tokenManager, []string{"INIT", "req-id", "bucket-id", "2", "1000"})

	if err != nil {
		t.Fatalf(`Got error executing request, got: %+v`, err)
	}

	if got != "req-id: DONE" {
		t.Fatalf(`Got wrong output, got: %+v`, got)
	}
}

func TestExecuteCHECKRequest(t *testing.T) {
	tokenManager := tokens.NewUserTokenBucketManager()
	tokenManager.CreateTokenBucket("bucket-id", 1, 1)

	got, err := executeRequest(tokenManager, []string{"CHECK", "req-id", "bucket-id"})

	if err != nil {
		t.Fatalf(`Got error executing request, got: %+v`, err)
	}

	if got != "req-id: AVAILABLE" {
		t.Fatalf(`Got wrong output, got: %+v`, got)
	}
}

func TestExecuteCHECKRequestReturnsErrorWhenNotFound(t *testing.T) {
	tokenManager := tokens.NewUserTokenBucketManager()

	_, err := executeRequest(tokenManager, []string{"CHECK", "req-id", "bucket-id"})

	if err.Error() != "bucket bucket-id not found" {
		t.Fatalf(`Error was expected, got: %+v`, err)
	}
}
