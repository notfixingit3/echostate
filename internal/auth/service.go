package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/notfixingit3/echostate/internal/config"
	"github.com/notfixingit3/echostate/internal/db"
)

const (
	PurposeInitial  = "initial"
	PurposeDevice   = "device"
	PurposeRecovery = "recovery"

	ChallengeRegister = "register"
	ChallengeLogin    = "login"

	SessionCookieName   = "echostate_session"
	EnrollCookieName    = "echostate_enroll"
	enrollCookieMinutes = 15
)

var (
	ErrInvalidCode      = errors.New("invalid or expired enrollment code")
	ErrTooManyAttempts  = errors.New("too many code attempts")
	ErrAuthDisabled     = errors.New("authentication is disabled")
	ErrUserDisabled     = errors.New("user is disabled")
	ErrNoCredentials    = errors.New("no passkey enrolled")
	ErrBootstrapExists  = errors.New("admin user already exists")
	ErrBreakGlassSecret = errors.New("break-glass secret required")
)

type User struct {
	ID          uuid.UUID `json:"id"`
	DisplayName string    `json:"display_name"`
	Role        string    `json:"role"`
	Disabled    bool      `json:"disabled"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Credential struct {
	ID         uuid.UUID  `json:"id"`
	Nickname   string     `json:"nickname"`
	CreatedAt  time.Time  `json:"created_at"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
}

type EnrollmentCodeResult struct {
	Code      string    `json:"code"`
	ExpiresAt time.Time `json:"expires_at"`
	Purpose   string    `json:"purpose"`
	UserID    uuid.UUID `json:"user_id"`
}

type Service struct {
	db           *db.DB
	frontendURL  string
	webauthn     *webauthn.WebAuthn
	webauthnOnce func() (*webauthn.WebAuthn, error)
}

func NewService(database *db.DB, frontendURL string) *Service {
	return &Service{db: database, frontendURL: frontendURL}
}

func (s *Service) Settings() Settings {
	return SettingsFromConfig(config.GetSettings(), s.frontendURL)
}

func (s *Service) AuthRequired(ctx context.Context) (bool, error) {
	if !s.Settings().AuthEnabled {
		return false, nil
	}
	count, err := s.UserCount(ctx)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (s *Service) UserCount(ctx context.Context) (int, error) {
	var count int
	err := s.db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM users`).Scan(&count)
	return count, err
}

func (s *Service) getWebAuthn() (*webauthn.WebAuthn, error) {
	if s.webauthn != nil {
		return s.webauthn, nil
	}
	if s.webauthnOnce != nil {
		w, err := s.webauthnOnce()
		if err != nil {
			return nil, err
		}
		s.webauthn = w
		return w, nil
	}
	settings := s.Settings()
	w, err := webauthn.New(&webauthn.Config{
		RPDisplayName: "EchoState",
		RPID:          settings.WebAuthnRPID,
		RPOrigins:     []string{settings.WebAuthnRPOrigin},
	})
	if err != nil {
		return nil, err
	}
	s.webauthn = w
	return w, nil
}

// WebAuthnUser implements webauthn.User for ceremony flows.
type WebAuthnUser struct {
	User
	CredentialList []webauthn.Credential
}

func (u WebAuthnUser) WebAuthnID() []byte {
	return u.ID[:]
}

func (u WebAuthnUser) WebAuthnName() string {
	return u.DisplayName
}

func (u WebAuthnUser) WebAuthnDisplayName() string {
	return u.DisplayName
}

func (u WebAuthnUser) WebAuthnCredentials() []webauthn.Credential {
	return u.CredentialList
}

func (s *Service) LoadWebAuthnUser(ctx context.Context, userID uuid.UUID) (*WebAuthnUser, error) {
	user, err := s.GetUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	creds, err := s.loadWebAuthnCredentials(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &WebAuthnUser{User: *user, CredentialList: creds}, nil
}

func (s *Service) GetUser(ctx context.Context, userID uuid.UUID) (*User, error) {
	var user User
	err := s.db.Pool.QueryRow(ctx, `
		SELECT id, display_name, role, disabled, created_at, updated_at
		FROM users WHERE id = $1
	`, userID).Scan(&user.ID, &user.DisplayName, &user.Role, &user.Disabled, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *Service) ListUsers(ctx context.Context) ([]User, error) {
	rows, err := s.db.Pool.Query(ctx, `
		SELECT id, display_name, role, disabled, created_at, updated_at
		FROM users ORDER BY created_at ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.ID, &user.DisplayName, &user.Role, &user.Disabled, &user.CreatedAt, &user.UpdatedAt); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, rows.Err()
}

func (s *Service) CreateUser(ctx context.Context, displayName, role string) (*User, error) {
	displayName = stringsTrim(displayName)
	if displayName == "" || !RoleIsValid(role) {
		return nil, fmt.Errorf("invalid user")
	}
	var user User
	err := s.db.Pool.QueryRow(ctx, `
		INSERT INTO users (display_name, role)
		VALUES ($1, $2)
		RETURNING id, display_name, role, disabled, created_at, updated_at
	`, displayName, role).Scan(&user.ID, &user.DisplayName, &user.Role, &user.Disabled, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *Service) BootstrapAdmin(ctx context.Context, displayName string) (*User, *EnrollmentCodeResult, error) {
	count, err := s.UserCount(ctx)
	if err != nil {
		return nil, nil, err
	}
	if count > 0 {
		return nil, nil, ErrBootstrapExists
	}
	if stringsTrim(displayName) == "" {
		displayName = "Admin"
	}
	user, err := s.CreateUser(ctx, displayName, RoleAdmin)
	if err != nil {
		return nil, nil, err
	}
	code, err := s.IssueEnrollmentCode(ctx, user.ID, nil, PurposeInitial, false)
	if err != nil {
		return nil, nil, err
	}
	return user, code, nil
}

func (s *Service) IssueEnrollmentCode(ctx context.Context, userID uuid.UUID, createdBy *uuid.UUID, purpose string, recovery bool) (*EnrollmentCodeResult, error) {
	settings := s.Settings()
	length := settings.EnrollmentCodeLength
	if recovery {
		length = settings.RecoveryCodeLength
	}
	var code string
	var err error
	if recovery {
		code, err = GenerateRecoveryCode(length)
	} else {
		code, err = GenerateNumericCode(length)
	}
	if err != nil {
		return nil, err
	}
	expires := time.Now().Add(time.Duration(settings.EnrollmentCodeTTLHours) * time.Hour)
	var createdByVal *uuid.UUID
	if createdBy != nil {
		createdByVal = createdBy
	}
	_, err = s.db.Pool.Exec(ctx, `
		INSERT INTO enrollment_codes (user_id, code_hash, purpose, expires_at, created_by, attempts)
		VALUES ($1, $2, $3, $4, $5, 0)
	`, userID, HashCode(code), purpose, expires, createdByVal)
	if err != nil {
		return nil, err
	}
	return &EnrollmentCodeResult{
		Code:      code,
		ExpiresAt: expires,
		Purpose:   purpose,
		UserID:    userID,
	}, nil
}

func (s *Service) VerifyEnrollmentCode(ctx context.Context, code, clientIP string) (*User, error) {
	settings := s.Settings()
	normalized := NormalizeEnrollmentInput(code)
	if normalized == "" {
		return nil, ErrInvalidCode
	}
	hash := HashCode(normalized)

	var (
		codeID    uuid.UUID
		userID    uuid.UUID
		attempts  int
		firstTry  time.Time
		expiresAt time.Time
		usedAt    *time.Time
	)
	err := s.db.Pool.QueryRow(ctx, `
		SELECT id, user_id, attempts, created_at, expires_at, used_at
		FROM enrollment_codes
		WHERE code_hash = $1
		ORDER BY created_at DESC
		LIMIT 1
	`, hash).Scan(&codeID, &userID, &attempts, &firstTry, &expiresAt, &usedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			_ = s.recordFailedAttempt(ctx, nil, clientIP)
			return nil, ErrInvalidCode
		}
		return nil, err
	}
	if usedAt != nil || time.Now().After(expiresAt) {
		return nil, ErrInvalidCode
	}
	windowStart := time.Now().Add(-time.Duration(settings.CodeAttemptWindowMinutes) * time.Minute)
	if attempts >= settings.MaxCodeAttempts && firstTry.After(windowStart) {
		return nil, ErrTooManyAttempts
	}

	user, err := s.GetUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user.Disabled {
		return nil, ErrUserDisabled
	}

	_, err = s.db.Pool.Exec(ctx, `
		UPDATE enrollment_codes
		SET used_at = NOW(), attempts = attempts + 1
		WHERE id = $1 AND used_at IS NULL
	`, codeID)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *Service) recordFailedAttempt(ctx context.Context, codeID *uuid.UUID, clientIP string) error {
	if codeID != nil {
		_, err := s.db.Pool.Exec(ctx, `UPDATE enrollment_codes SET attempts = attempts + 1 WHERE id = $1`, *codeID)
		return err
	}
	return nil
}

func (s *Service) CreateEnrollmentSession(ctx context.Context, userID uuid.UUID) (string, time.Time, error) {
	token, err := NewSessionToken()
	if err != nil {
		return "", time.Time{}, err
	}
	expires := time.Now().Add(enrollCookieMinutes * time.Minute)
	_, err = s.db.Pool.Exec(ctx, `
		INSERT INTO enrollment_sessions (token_hash, user_id, expires_at)
		VALUES ($1, $2, $3)
	`, HashCode(token), userID, expires)
	if err != nil {
		return "", time.Time{}, err
	}
	return token, expires, nil
}

func (s *Service) EnrollmentUser(ctx context.Context, token string) (*User, error) {
	var userID uuid.UUID
	err := s.db.Pool.QueryRow(ctx, `
		SELECT user_id FROM enrollment_sessions
		WHERE token_hash = $1 AND expires_at > NOW()
	`, HashCode(stringsTrim(token))).Scan(&userID)
	if err != nil {
		return nil, err
	}
	return s.GetUser(ctx, userID)
}

func (s *Service) DeleteEnrollmentSession(ctx context.Context, token string) error {
	_, err := s.db.Pool.Exec(ctx, `DELETE FROM enrollment_sessions WHERE token_hash = $1`, HashCode(stringsTrim(token)))
	return err
}

func (s *Service) CreateSession(ctx context.Context, userID uuid.UUID, userAgent, ip string) (string, time.Time, error) {
	token, err := NewSessionToken()
	if err != nil {
		return "", time.Time{}, err
	}
	expires := time.Now().Add(time.Duration(s.Settings().SessionTTLHours) * time.Hour)
	_, err = s.db.Pool.Exec(ctx, `
		INSERT INTO sessions (user_id, token_hash, expires_at, user_agent, ip)
		VALUES ($1, $2, $3, $4, $5)
	`, userID, HashCode(token), expires, userAgent, ip)
	if err != nil {
		return "", time.Time{}, err
	}
	return token, expires, nil
}

func (s *Service) SessionUser(ctx context.Context, token string) (*User, error) {
	if stringsTrim(token) == "" {
		return nil, pgx.ErrNoRows
	}
	var user User
	var sessionID uuid.UUID
	err := s.db.Pool.QueryRow(ctx, `
		SELECT s.id, u.id, u.display_name, u.role, u.disabled, u.created_at, u.updated_at
		FROM sessions s
		JOIN users u ON u.id = s.user_id
		WHERE s.token_hash = $1 AND s.expires_at > NOW() AND u.disabled = FALSE
	`, HashCode(token)).Scan(
		&sessionID, &user.ID, &user.DisplayName, &user.Role, &user.Disabled, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	_, _ = s.db.Pool.Exec(ctx, `UPDATE sessions SET last_seen_at = NOW() WHERE id = $1`, sessionID)
	return &user, nil
}

func (s *Service) DeleteSession(ctx context.Context, token string) error {
	_, err := s.db.Pool.Exec(ctx, `DELETE FROM sessions WHERE token_hash = $1`, HashCode(stringsTrim(token)))
	return err
}

func (s *Service) SaveChallenge(ctx context.Context, userID *uuid.UUID, kind string, data *webauthn.SessionData) (string, error) {
	id, err := NewChallengeID()
	if err != nil {
		return "", err
	}
	payload, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	expires := time.Now().Add(enrollCookieMinutes * time.Minute)
	_, err = s.db.Pool.Exec(ctx, `
		INSERT INTO webauthn_challenges (id, user_id, kind, challenge_data, expires_at)
		VALUES ($1, $2, $3, $4, $5)
	`, id, userID, kind, payload, expires)
	return id, err
}

func (s *Service) LoadChallenge(ctx context.Context, challengeID string, kind string) (*webauthn.SessionData, *uuid.UUID, error) {
	var payload []byte
	var userID *uuid.UUID
	err := s.db.Pool.QueryRow(ctx, `
		SELECT challenge_data, user_id
		FROM webauthn_challenges
		WHERE id = $1 AND kind = $2 AND expires_at > NOW()
	`, challengeID, kind).Scan(&payload, &userID)
	if err != nil {
		return nil, nil, err
	}
	var data webauthn.SessionData
	if err := json.Unmarshal(payload, &data); err != nil {
		return nil, nil, err
	}
	_, _ = s.db.Pool.Exec(ctx, `DELETE FROM webauthn_challenges WHERE id = $1`, challengeID)
	return &data, userID, nil
}

func (s *Service) StoreCredential(ctx context.Context, userID uuid.UUID, cred webauthn.Credential, nickname string) error {
	payload, err := json.Marshal(cred)
	if err != nil {
		return err
	}
	_, err = s.db.Pool.Exec(ctx, `
		INSERT INTO webauthn_credentials (user_id, credential_id, public_key, sign_count, nickname, transport)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (credential_id) DO UPDATE SET
			public_key = EXCLUDED.public_key,
			sign_count = EXCLUDED.sign_count,
			last_used_at = NOW()
	`, userID, cred.ID, payload, cred.Authenticator.SignCount, stringsTrim(nickname), "")
	return err
}

func (s *Service) ListCredentials(ctx context.Context, userID uuid.UUID) ([]Credential, error) {
	rows, err := s.db.Pool.Query(ctx, `
		SELECT id, nickname, created_at, last_used_at
		FROM webauthn_credentials
		WHERE user_id = $1
		ORDER BY created_at ASC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var creds []Credential
	for rows.Next() {
		var cred Credential
		if err := rows.Scan(&cred.ID, &cred.Nickname, &cred.CreatedAt, &cred.LastUsedAt); err != nil {
			return nil, err
		}
		creds = append(creds, cred)
	}
	return creds, rows.Err()
}

func (s *Service) DeleteCredential(ctx context.Context, userID, credID uuid.UUID) error {
	tag, err := s.db.Pool.Exec(ctx, `
		DELETE FROM webauthn_credentials WHERE id = $1 AND user_id = $2
	`, credID, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (s *Service) loadWebAuthnCredentials(ctx context.Context, userID uuid.UUID) ([]webauthn.Credential, error) {
	rows, err := s.db.Pool.Query(ctx, `
		SELECT public_key FROM webauthn_credentials WHERE user_id = $1
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var creds []webauthn.Credential
	for rows.Next() {
		var payload []byte
		if err := rows.Scan(&payload); err != nil {
			return nil, err
		}
		var cred webauthn.Credential
		if err := json.Unmarshal(payload, &cred); err != nil {
			return nil, err
		}
		creds = append(creds, cred)
	}
	return creds, rows.Err()
}

func (s *Service) UpdateCredentialSignCount(ctx context.Context, userID uuid.UUID, credID []byte, signCount uint32) error {
	_, err := s.db.Pool.Exec(ctx, `
		UPDATE webauthn_credentials
		SET sign_count = $1, last_used_at = NOW()
		WHERE user_id = $2 AND credential_id = $3
	`, signCount, userID, credID)
	return err
}

func (s *Service) FindAdminUser(ctx context.Context) (*User, error) {
	var user User
	err := s.db.Pool.QueryRow(ctx, `
		SELECT id, display_name, role, disabled, created_at, updated_at
		FROM users WHERE role = $1 AND disabled = FALSE
		ORDER BY created_at ASC LIMIT 1
	`, RoleAdmin).Scan(&user.ID, &user.DisplayName, &user.Role, &user.Disabled, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *Service) BeginRegistration(ctx context.Context, userID uuid.UUID) (*protocol.CredentialCreation, string, error) {
	wa, err := s.getWebAuthn()
	if err != nil {
		return nil, "", err
	}
	wuser, err := s.LoadWebAuthnUser(ctx, userID)
	if err != nil {
		return nil, "", err
	}
	options, session, err := wa.BeginRegistration(wuser)
	if err != nil {
		return nil, "", err
	}
	challengeID, err := s.SaveChallenge(ctx, &userID, ChallengeRegister, session)
	if err != nil {
		return nil, "", err
	}
	return options, challengeID, nil
}

func (s *Service) FinishRegistration(ctx context.Context, userID uuid.UUID, challengeID string, body []byte) (*webauthn.Credential, error) {
	wa, err := s.getWebAuthn()
	if err != nil {
		return nil, err
	}
	session, storedUser, err := s.LoadChallenge(ctx, challengeID, ChallengeRegister)
	if err != nil {
		return nil, err
	}
	if storedUser == nil || *storedUser != userID {
		return nil, fmt.Errorf("challenge user mismatch")
	}
	wuser, err := s.LoadWebAuthnUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	parsed, err := protocol.ParseCredentialCreationResponseBytes(body)
	if err != nil {
		return nil, err
	}
	return wa.CreateCredential(wuser, *session, parsed)
}

func (s *Service) BeginLogin(ctx context.Context, userID uuid.UUID) (*protocol.CredentialAssertion, string, error) {
	wa, err := s.getWebAuthn()
	if err != nil {
		return nil, "", err
	}
	wuser, err := s.LoadWebAuthnUser(ctx, userID)
	if err != nil {
		return nil, "", err
	}
	if len(wuser.CredentialList) == 0 {
		return nil, "", ErrNoCredentials
	}
	options, session, err := wa.BeginLogin(wuser)
	if err != nil {
		return nil, "", err
	}
	challengeID, err := s.SaveChallenge(ctx, &userID, ChallengeLogin, session)
	return options, challengeID, err
}

func (s *Service) FinishLogin(ctx context.Context, userID uuid.UUID, challengeID string, body []byte) (*webauthn.Credential, error) {
	wa, err := s.getWebAuthn()
	if err != nil {
		return nil, err
	}
	session, storedUser, err := s.LoadChallenge(ctx, challengeID, ChallengeLogin)
	if err != nil {
		return nil, err
	}
	if storedUser == nil || *storedUser != userID {
		return nil, fmt.Errorf("challenge user mismatch")
	}
	wuser, err := s.LoadWebAuthnUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	parsed, err := protocol.ParseCredentialRequestResponseBytes(body)
	if err != nil {
		return nil, err
	}
	cred, err := wa.ValidateLogin(wuser, *session, parsed)
	if err != nil {
		return nil, err
	}
	_ = s.UpdateCredentialSignCount(ctx, userID, cred.ID, cred.Authenticator.SignCount)
	return cred, nil
}

func stringsTrim(value string) string {
	return strings.TrimSpace(value)
}