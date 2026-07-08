package auth

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/notfixingit3/echostate/internal/db"
)

func TestRecoveryRegistrationReplacesExistingCredentials(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Skip("DATABASE_URL not set")
	}

	database, err := db.Connect(databaseURL)
	if err != nil {
		t.Fatalf("connect db: %v", err)
	}
	defer database.Close()

	if err := db.Migrate(database); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	ctx := context.Background()
	svc := NewService(database, "http://localhost:3001")

	user, err := svc.CreateUser(ctx, "Recovery Test", RoleAdmin)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	oldCredID := []byte("old-credential-id")
	newCredID := []byte("new-credential-id")
	for _, cred := range [][]byte{oldCredID, newCredID} {
		_, err = database.Pool.Exec(ctx, `
			INSERT INTO webauthn_credentials (user_id, credential_id, public_key, nickname)
			VALUES ($1, $2, $3, 'device')
		`, user.ID, cred, []byte(`{"id":"device"}`))
		if err != nil {
			t.Fatalf("insert credential: %v", err)
		}
	}

	token, _, err := svc.CreateEnrollmentSession(ctx, user.ID, PurposeRecovery)
	if err != nil {
		t.Fatalf("create enrollment session: %v", err)
	}

	session, err := svc.LookupEnrollmentSession(ctx, token)
	if err != nil {
		t.Fatalf("lookup enrollment session: %v", err)
	}
	if session.Purpose != PurposeRecovery {
		t.Fatalf("expected recovery purpose, got %q", session.Purpose)
	}

	if err := svc.DeleteAllCredentialsExcept(ctx, user.ID, newCredID); err != nil {
		t.Fatalf("delete old credentials: %v", err)
	}

	var count int
	if err := database.Pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM webauthn_credentials WHERE user_id = $1
	`, user.ID).Scan(&count); err != nil {
		t.Fatalf("count credentials: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 credential after recovery replace, got %d", count)
	}

	var remaining []byte
	if err := database.Pool.QueryRow(ctx, `
		SELECT credential_id FROM webauthn_credentials WHERE user_id = $1
	`, user.ID).Scan(&remaining); err != nil {
		t.Fatalf("load remaining credential: %v", err)
	}
	if string(remaining) != string(newCredID) {
		t.Fatalf("expected remaining credential %q, got %q", newCredID, remaining)
	}
}

func TestVerifyEnrollmentCodeTracksIPFailures(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Skip("DATABASE_URL not set")
	}

	database, err := db.Connect(databaseURL)
	if err != nil {
		t.Fatalf("connect db: %v", err)
	}
	defer database.Close()

	if err := db.Migrate(database); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	ctx := context.Background()
	svc := NewService(database, "http://localhost:3001")
	clientIP := "203.0.113.50"
	if _, err := database.Pool.Exec(ctx, `DELETE FROM enrollment_verify_attempts WHERE ip = $1`, clientIP); err != nil {
		t.Fatalf("clear prior attempts: %v", err)
	}
	maxAttempts := svc.Settings().MaxCodeAttempts

	for i := 0; i < maxAttempts-1; i++ {
		_, err := svc.VerifyEnrollmentCode(ctx, fmt.Sprintf("0000000%d", i), clientIP)
		if err != ErrInvalidCode {
			t.Fatalf("attempt %d: expected ErrInvalidCode, got %v", i+1, err)
		}
	}

	_, err = svc.VerifyEnrollmentCode(ctx, "00000009", clientIP)
	if err != ErrTooManyAttempts {
		t.Fatalf("expected ErrTooManyAttempts after repeated failures, got %v", err)
	}
}
