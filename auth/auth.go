package auth

import (
	"crypto/rand"
	"errors"
	"time"

	"github.com/Parsa-Sh-Y/book-manager-service/db"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrIncorrectPassword = errors.New("Incorrect Password")
)

type UserCredentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type claims struct {
	jwt.MapClaims
	Username string `json:"username"`
	Password string `json:"password"`
}

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

// returns an empty string when there is an error
func (a *Auth) Login(cred *UserCredentials) (string, error) {

	// get the user from the database
	user, err := a.db.GetUserByUsername(cred.Username)
	if err != nil {
		return "", err
	}

	// check if password is correct
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(cred.Password)); err != nil {
		return "", ErrIncorrectPassword
	}

	// Create the JWT token
	expirationTime := time.Now().Add(a.jwtExpirationDuration)
	tokenJWT := jwt.NewWithClaims(jwt.SigningMethodHS256, &claims{
		Username: cred.Username,
		MapClaims: jwt.MapClaims{
			"expired_at": expirationTime.Unix(),
		},
	})

	tokenString, err := tokenJWT.SignedString(a.jwtSecretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
