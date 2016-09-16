package utils

import (
	"bufio"
	crand "crypto/rand"
	"crypto/sha512"
	"crypto/subtle"
	"encoding/hex"
	"io"
	"math/rand"
	"time"

	"golang.org/x/crypto/hkdf"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func FormatDate(t time.Time) string {
	return t.UTC().Format("2006-01-02T15:04:05+00:00")
}

func Sponge(context []byte, key []byte) io.Reader {
	// First 32 hexadecimal digits of e.
	const HKDF_SALT = "2b7e151628aed2a6abf7158809cf4f3c"
	return hkdf.New(sha512.New, key, FromHex(HKDF_SALT), context)
}

func DeriveKey(context string, size int, masterKey []byte) (key []byte) {
	key = make([]byte, size)
	io.ReadFull(Sponge([]byte(context), masterKey), key)
	return
}

// excludes 0O1lI (base 57) for legibility
var Base57 = "23456789abcdefghijkmnopqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ"

func DeriveBase57Key(context string, length int, masterKey []byte) (key []byte) {
	prng := bufio.NewReader(Sponge([]byte(context), masterKey))
	key = make([]byte, length)
	for i := range key {
		for {
			digit, _ := prng.ReadByte()
			digit &= 63
			if digit < 57 {
				key[i] = byte(Base57[digit])
				break
			}
		}
	}
	return
}

func RandBase57String(length int) string {
	ret := make([]byte, length)
	for i := 0; i < length; i++ {
		ret[i] = Base57[rand.Intn(len(Base57))]
	}
	return string(ret)
}

func CryptoRand(n int) (key []byte) {
	key = make([]byte, n)
	crand.Read(key)
	return
}

func ToHex(bytes []byte) string {
	return hex.EncodeToString(bytes)
}

func FromHex(digits string) []byte {
	bytes, _ := hex.DecodeString(digits)
	return bytes
}

func DeriveExpiryCode(context string, offset time.Duration, key []byte) []byte {
	context += time.Now().UTC().Add(offset).Truncate(time.Hour).String()
	return DeriveBase57Key(context, 22, key)
}

func CheckExpiryCode(code string, context string, key string) bool {
	for o := time.Duration(0); o >= -time.Hour; o -= time.Hour {
		if subtle.ConstantTimeCompare([]byte(code), DeriveExpiryCode(context, o, FromHex(key))) == 1 {
			return true
		}
	}
	return false
}
