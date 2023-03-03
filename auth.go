package stupid

import (
	"crypto/rand"
	"encoding/base64"

	"golang.org/x/crypto/argon2"
)

func hashPassword(password, salt string) string {
	return base64.RawStdEncoding.EncodeToString(argon2.IDKey([]byte(password), []byte(salt), 1, 64*1024, 4, 32))
}

func generateSalt() (string, error) {
	hold := make([]byte, 16)
	_, err := rand.Read(hold)
	return base64.RawStdEncoding.EncodeToString(hold), err
}
