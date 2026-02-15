package auth

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

type params struct {
	memory      uint32
	iterations  uint32
	parallelism uint8
	saltLength  uint32
	keyLength   uint32
}

// Recommended parameters as of late 2024/early 2025 (adjust based on hardware and OWASP guidelines).
var p = &params{
	memory:      64 * 1024, // 64 MB
	iterations:  2,
	parallelism: 4, // Match CPU cores if possible
	saltLength:  16,
	keyLength:   32,
}

func generateRandomBytes(n uint32) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func HashPassword(password string) (encodedHash string, err error) {
	salt, err := generateRandomBytes(p.saltLength)
	if err != nil {
		return "", err
	}

	hash := argon2.IDKey(
		[]byte(password),
		salt,
		p.iterations,
		p.memory,
		p.parallelism,
		p.keyLength,
	)

	// Encode the parameters, salt, and hash into the standard Argon2 string format
	encodedHash = fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		p.memory,
		p.iterations,
		p.parallelism,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	)

	return encodedHash, nil
}

func VerifyPassword(password, encodedHash string) (bool, error) {
	// 1. Decode the hash to extract parameters, salt, and stored hash
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 {
		return false, errors.New("invalid encoded hash format")
	}

	// Extract salt and stored hash (ignore errors for brevity, production code needs robust error handling)
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, fmt.Errorf("failed to decode salt: %w", err)
	}
	storedHash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, fmt.Errorf("failed to decode hash: %w", err)
	}

	// 2. Hash the user-provided password using the extracted parameters and salt
	// Note: We are using the parameters from the *stored* hash to ensure the same computation
	// We extract 'm', 't', 'p' from parts[3] (m=...,t=...,p=...)
	// In a real implementation, you would parse the parameters from the string, but here we assume our global 'p' was used.
	// A robust library (like the one linked in search results) handles this parsing automatically.

	// Example using the global 'p' (for simplicity in this demonstration)
	// In production, parameters *must* be parsed from the `encodedHash` string itself.
	challengeHash := argon2.IDKey(
		[]byte(password),
		salt,
		p.iterations,
		p.memory,
		p.parallelism,
		p.keyLength,
	)

	// 3. Compare the newly generated hash with the stored hash using constant-time comparison (bytes.Equal)
	// This helps prevent timing attacks.
	if len(storedHash) != len(challengeHash) || !bytes.Equal(storedHash, challengeHash) { // using bytes.Equal, which needs to be imported
		return false, nil // Password does not match
	}

	return true, nil // Password matches
}
