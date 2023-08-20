package auth

import (
	"crypto/rand"
	"errors"
	"time"

	"github.com/Parsa-Sh-Y/book-manager-service/db"
)

type Auth struct {
	db *db.GormDB
	// jwtSecretKey is the JWT secret key. Each time the server starts, new key is generated.
	jwtSecretKey          []byte
	jwtExpirationDuration time.Duration
}

// NewAuth creates new instance of Auth for authenticating user accounts.
func NewAuth(authDB *db.GormDB, jwtExpirationInMinutes int64) (*Auth, error) {
	secretKey, err := generateRandomKey()
	if err != nil {
		return nil, err
	}

	// Check the authDB
	if authDB == nil {
		return nil, errors.New("the authenticate database is essential")
	}

	return &Auth{
		db:                    authDB,
		jwtSecretKey:          secretKey,
		jwtExpirationDuration: time.Duration(int64(time.Minute) * jwtExpirationInMinutes),
	}, nil
}

// generateRandomKey
// Each time that Auth is initialized, generateRandomKey is called to
// generate another key
func generateRandomKey() ([]byte, error) {
	jwtKey := make([]byte, 32)
	if _, err := rand.Read(jwtKey); err != nil {
		return nil, err
	}

	return jwtKey, nil
}
