package password

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"golang.org/x/crypto/argon2"
)

const (
	version     = 19
	memoryKiB   = 64 * 1024
	iterations  = 3
	parallelism = 2
	saltBytes   = 16
	keyBytes    = 32
)

func HashPassword(value string) (string, error) {
	salt := make([]byte, saltBytes)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("generate password salt: %w", err)
	}
	key := argon2.IDKey([]byte(value), salt, iterations, memoryKiB, parallelism, keyBytes)
	return fmt.Sprintf("argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		version,
		memoryKiB,
		iterations,
		parallelism,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(key),
	), nil
}

func VerifyPassword(value, encoded string) (bool, error) {
	params, salt, expected, err := parse(encoded)
	if err != nil {
		return false, err
	}
	actual := argon2.IDKey([]byte(value), salt, params.iterations, params.memoryKiB, params.parallelism, uint32(len(expected)))
	return subtle.ConstantTimeCompare(actual, expected) == 1, nil
}

func NeedsRehash(encoded string) bool {
	params, _, _, err := parse(encoded)
	if err != nil {
		return true
	}
	return params.memoryKiB != memoryKiB || params.iterations != iterations || params.parallelism != parallelism
}

var dummy struct {
	sync.Once
	hash string
}

func DummyHash() string {
	dummy.Do(func() {
		dummy.hash, _ = HashPassword("invalid-login-password")
	})
	return dummy.hash
}

type parameters struct {
	memoryKiB   uint32
	iterations  uint32
	parallelism uint8
}

func parse(encoded string) (parameters, []byte, []byte, error) {
	parts := strings.Split(encoded, "$")
	if len(parts) != 5 || parts[0] != "argon2id" {
		return parameters{}, nil, nil, errors.New("invalid Argon2id hash format")
	}
	if parts[1] != "v=19" {
		return parameters{}, nil, nil, errors.New("unsupported Argon2id version")
	}
	values := map[string]string{}
	for _, part := range strings.Split(parts[2], ",") {
		key, value, found := strings.Cut(part, "=")
		if !found {
			return parameters{}, nil, nil, errors.New("invalid Argon2id parameters")
		}
		values[key] = value
	}
	memory, err := strconv.ParseUint(values["m"], 10, 32)
	if err != nil || memory == 0 {
		return parameters{}, nil, nil, errors.New("invalid Argon2id memory parameter")
	}
	iter, err := strconv.ParseUint(values["t"], 10, 32)
	if err != nil || iter == 0 {
		return parameters{}, nil, nil, errors.New("invalid Argon2id iteration parameter")
	}
	parallel, err := strconv.ParseUint(values["p"], 10, 8)
	if err != nil || parallel == 0 {
		return parameters{}, nil, nil, errors.New("invalid Argon2id parallelism parameter")
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[3])
	if err != nil || len(salt) < 8 {
		return parameters{}, nil, nil, errors.New("invalid Argon2id salt")
	}
	expected, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil || len(expected) < 16 {
		return parameters{}, nil, nil, errors.New("invalid Argon2id key")
	}
	return parameters{memoryKiB: uint32(memory), iterations: uint32(iter), parallelism: uint8(parallel)}, salt, expected, nil
}
