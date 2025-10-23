package data

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base32"
	"time"
)

const (
	ScopeActivation = "activation"
	ScopeRefresh    = "refresh"
)

// Token represents an application token used for account activation or refresh flows.
// Plaintext is only available at creation time; Hash is stored in the database.
// Scope differentiates token usage (e.g., activation, refresh).
type Token struct {
	Plaintext string
	Hash      []byte
	UserID    int64
	Expiry    time.Time
	Scope     string
}

// generateToken creates a new Token with a random plaintext value and SHA-256 hash.
// The plaintext is Base32(no padding) encoded. The caller is responsible for persisting it.
func generateToken(userID int64, ttl time.Duration, scope string) (*Token, error) {
	token := &Token{
		UserID: userID,
		Expiry: time.Now().Add(ttl),
		Scope:  scope,
	}

	randomBytes := make([]byte, 16)

	_, err := rand.Read(randomBytes)
	if err != nil {
		return nil, err
	}

	token.Plaintext = base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randomBytes)

	hash := sha256.Sum256([]byte(token.Plaintext))
	token.Hash = hash[:]

	return token, nil
}

type tokenModel struct {
	DB *sql.DB
}

// New creates a Token for the given user and scope, persists it, and returns
// the token including its plaintext value for one-time presentation.
func (m tokenModel) New(userID int64, ttl time.Duration, scope string) (*Token, error) {
	token, err := generateToken(userID, ttl, scope)
	if err != nil {
		return nil, err
	}

	err = m.Insert(token)
	return token, err
}

// Insert persists a token hash with associated metadata.
func (m tokenModel) Insert(token *Token) error {
	query := `
	INSERT INTO tokens (hash, user_id, expiry, scope) 
	VALUES ($1, $2, $3, $4)`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, query, token.Hash, token.UserID, token.Expiry, token.Scope)
	return err
}

// DeleteAllForUser removes all tokens for a user within the specified scope.
func (m tokenModel) DeleteAllForUser(scope string, userID int64) error {
	query := `
	DELETE FROM tokens
	WHERE scope = $1 AND user_id = $2`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, query, scope, userID)
	return err
}
