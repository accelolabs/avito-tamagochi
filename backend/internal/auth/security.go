package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	minPasswordLength = 8
	maxPasswordLength = 128
	argonMemory       = 64 * 1024
	argonIterations   = 1
	argonParallelism  = 4
	argonKeyLength    = 32
	argonSaltLength   = 16
)

func validPassword(password string) bool {
	return len(password) >= minPasswordLength && len(password) <= maxPasswordLength
}

func hashPassword(password string) (string, error) {
	salt := make([]byte, argonSaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("generate password salt: %w", err)
	}
	hash := argon2.IDKey([]byte(password), salt, argonIterations, argonMemory, argonParallelism, argonKeyLength)
	return fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		argonMemory,
		argonIterations,
		argonParallelism,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	), nil
}

func verifyPassword(password, encoded string) bool {
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 || parts[1] != "argon2id" {
		return false
	}

	version, ok := parseArgonParameter(parts[2], "v")
	if !ok || version != argon2.Version {
		return false
	}
	memory, ok := parseArgonParameter(parts[3], "m")
	if !ok || memory <= 0 || memory > 1024*1024 {
		return false
	}
	iterations, ok := parseArgonParameter(parts[3], "t")
	if !ok || iterations <= 0 || iterations > 10 {
		return false
	}
	parallelism, ok := parseArgonParameter(parts[3], "p")
	if !ok || parallelism <= 0 || parallelism > 255 {
		return false
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil || len(salt) < 8 {
		return false
	}
	expected, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil || len(expected) == 0 || len(expected) > 64 {
		return false
	}
	actual := argon2.IDKey([]byte(password), salt, uint32(iterations), uint32(memory), uint8(parallelism), uint32(len(expected)))
	return subtle.ConstantTimeCompare(actual, expected) == 1
}

func parseArgonParameter(parameters, name string) (int, bool) {
	for _, parameter := range strings.Split(parameters, ",") {
		key, value, ok := strings.Cut(parameter, "=")
		if !ok || key != name {
			continue
		}
		parsed, err := strconv.Atoi(value)
		return parsed, err == nil
	}
	return 0, false
}
