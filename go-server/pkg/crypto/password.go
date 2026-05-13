package crypto

import (
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	Argon2Memory      uint32 = 64 * 1024
	Argon2Iterations  uint32 = 3
	Argon2Parallelism uint8  = 4
	Argon2SaltLength         = 16
	Argon2KeyLength          = 32
)

// HashPassword hashes a password using argon2id.
func HashPassword(password string, salt []byte) (string, error) {
	hash := argon2.IDKey([]byte(password), salt, Argon2Iterations, Argon2Memory, Argon2Parallelism, Argon2KeyLength)
	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, Argon2Memory, Argon2Iterations, Argon2Parallelism,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash[:])), nil
}

// VerifyPassword checks if a password matches the given argon2id hash.
func VerifyPassword(hashWithSalt, password string) bool {
	parts := strings.Split(hashWithSalt, "$")
	if len(parts) != 6 {
		return false
	}

	if parts[1] != "argon2id" || parts[2] != fmt.Sprintf("v=%d", argon2.Version) {
		return false
	}

	var memory, iterations int
	var parallelism uint8
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &iterations, &parallelism); err != nil {
		return false
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil || len(salt) != Argon2SaltLength {
		return false
	}

	expectedHash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil || len(expectedHash) != Argon2KeyLength {
		return false
	}

	actualHash := argon2.IDKey([]byte(password), salt, uint32(iterations), uint32(memory), parallelism, Argon2KeyLength)
	return subtle.ConstantTimeCompare(expectedHash, actualHash) == 1
}
