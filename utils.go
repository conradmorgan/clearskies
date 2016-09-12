package main

import (
	"bufio"
	crand "crypto/rand"
	"crypto/sha512"
	"encoding/hex"
	"io"
	"math/rand"
	"time"

	"golang.org/x/crypto/hkdf"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func sponge(context []byte, key []byte) io.Reader {
	// First 32 hexadecimal digits of e.
	const HKDF_SALT = "2b7e151628aed2a6abf7158809cf4f3c"
	return hkdf.New(sha512.New, key, fromHex(HKDF_SALT), context)
}

func deriveKey(context string, size int, masterKey []byte) (key []byte) {
	key = make([]byte, size)
	io.ReadFull(sponge([]byte(context), masterKey), key)
	return
}

func deriveBase57Key(context string, length int, masterKey []byte) (key []byte) {
	const BASE_57 = "23456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"
	prng := bufio.NewReader(sponge([]byte(context), masterKey))
	key = make([]byte, length)
	for i := range key {
		for {
			digit, _ := prng.ReadByte()
			digit &= 63
			if digit < 57 {
				key[i] = byte(BASE_57[digit])
				break
			}
		}
	}
	return
}

// excludes 0O1lI (base 57) for legibility
var symbols = "23456789abcdefghijkmnopqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ"

func randBase57String(length int) string {
	ret := make([]byte, length)
	for i := 0; i < length; i++ {
		ret[i] = symbols[rand.Intn(len(symbols))]
	}
	return string(ret)
}

func cryptoRand(n int) (key []byte) {
	key = make([]byte, n)
	crand.Read(key)
	return
}

func toHex(bytes []byte) string {
	return hex.EncodeToString(bytes)
}

func fromHex(digits string) []byte {
	bytes, _ := hex.DecodeString(digits)
	return bytes
}
