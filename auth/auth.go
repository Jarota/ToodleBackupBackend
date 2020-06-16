package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

var (
	ErrInvalidHash 			= errors.New("The encoded hash is not in the correct format")
	ErrIncompatibleVersion	= errors.New("Incompatible version of argon2")
)

type argonParams struct {
	memory		uint32
	iterations	uint32
	parallelism uint8
	saltLength	uint32
	keyLength	uint32
}

var p *argonParams = &argonParams{
	memory: 		64 * 1024,
	iterations: 	3,
	parallelism: 	2,
	saltLength:		16,
	keyLength:		32,
}

func HashPassword(password string) (encodedHash string, err error) {
	salt, err := generateRandomBytes(p.saltLength)
	if err != nil {
		return "", err
	}

	hash := argon2.IDKey([]byte(password), salt, p.iterations, p.memory, p.parallelism, p.keyLength)

	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	encodedHash = fmt.Sprintf("$argon2id$v=%d$%s$%s", argon2.Version, b64Salt, b64Hash)

	return encodedHash, nil
}

func generateRandomBytes(n uint32) ([]byte, error) {
	b := make([]byte, n)

	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func ComparePassAndHash(password, encodedHash string) (match bool, err error) {
	salt, hash, err := decodeHash(encodedHash)
	if err != nil {
		return false, err
	}

	otherHash := argon2.IDKey([]byte(password), salt, p.iterations, p.memory, p.parallelism, p.keyLength)

	// using subtle.ConstantTimeCompare to help prevent timing attacks
	if subtle.ConstantTimeCompare(hash, otherHash) == 1 {
		return true, nil
	}
	return false, nil
}

func decodeHash(encodedHash string) (salt, hash []byte, err error) {
	vals := strings.Split(encodedHash, "$")
	if len(vals) != 5 {
		return nil, nil, ErrInvalidHash
	}

	var version int
	_, err = fmt.Sscanf(vals[2], "v=%d", &version)
	if err != nil {
		return nil, nil, err
	}
	if version != argon2.Version {
		return nil, nil, ErrIncompatibleVersion
	}

	salt, err = base64.RawStdEncoding.DecodeString(vals[3])
	if err != nil {
		return nil, nil, err
	}

	hash, err = base64.RawStdEncoding.DecodeString(vals[4])
	if err != nil {
		return nil, nil, err
	}

	return salt, hash, nil
	
}