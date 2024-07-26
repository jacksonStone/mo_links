package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

type JaxAuth[User any] struct {
	GetUser                  func(userId string) (User, error)
	GetUserHashPasswordField func(user User) string
	GetUserPasswordSaltField func(user User) string
	CreateRawCookieContents  func(user User) string
	CreateWrongPasswordError func() error
	GetEncryptionSecret      func() string
	GetHMACKey               func() string
	Experation               time.Duration
	CookieName               string
	UseDevCookie             bool
}

func NewJaxAuth[User any]() *JaxAuth[User] {
	return &JaxAuth[User]{
		Experation:   180 * 24 * time.Hour, // 180 days
		CookieName:   "auth",
		UseDevCookie: false,
	}
}

func (ja *JaxAuth[User]) AttemptLoginAndGetCookie(userId string, password string) (string, error) {
	user, err := ja.GetUser(userId)
	if err != nil {
		return "", err
	}

	hashPassword := ja.GetUserHashPasswordField(user)
	saltForPassword := ja.GetUserPasswordSaltField(user)
	if ja.hash(password+saltForPassword) != hashPassword {
		return "", ja.CreateWrongPasswordError()
	}

	cookieContents := ja.CreateRawCookieContents(user)
	encryptionSecret := ja.GetEncryptionSecret()
	nonce := getNonce()

	block, err := aes.NewCipher([]byte(encryptionSecret))
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	crypted := aesGCM.Seal(nil, nonce, []byte(cookieContents), nil)
	expiration := time.Now().Add(ja.Experation).Unix()
	hmacValue := ja.hash(string(crypted) + string(nonce) + fmt.Sprint(expiration) + ja.GetHMACKey())

	cookieContent := fmt.Sprintf("%x:%d:%x:%s", nonce, expiration, crypted, hmacValue)
	if ja.UseDevCookie {
		return fmt.Sprintf("%s=%s; Max-Age=%d; SameSite=Strict; Path=/", ja.CookieName, cookieContent, int(ja.Experation.Seconds())), nil
	}
	return fmt.Sprintf("%s=%s; HttpOnly; Max-Age=%d; SameSite=Strict; Secure; Path=/", ja.CookieName, cookieContent, int(ja.Experation.Seconds())), nil
}

func (ja *JaxAuth[User]) AttemptCookieDecryption(rawCookieHeader string) (string, error) {
	if rawCookieHeader == "" {
		return "", errors.New("no cookie value")
	}

	headers := parseCookieHeader(rawCookieHeader)
	rawAuthCookie, ok := headers[ja.CookieName]
	if !ok {
		return "", errors.New("no auth header")
	}

	if !ja.verifyBodyWithHMAC(rawAuthCookie) {
		return "", errors.New("tampered contents")
	}

	parts := strings.Split(rawAuthCookie, ":")
	nonce, _ := hex.DecodeString(parts[0])
	expiration, _ := strconv.ParseInt(parts[1], 10, 64)
	encryptedText, _ := hex.DecodeString(parts[2])

	if expiration < time.Now().Unix() {
		return "", errors.New("expired cookie")
	}

	block, err := aes.NewCipher([]byte(ja.GetEncryptionSecret()))
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	dec, err := aesGCM.Open(nil, nonce, encryptedText, nil)
	if err != nil {
		return "", err
	}

	return string(dec), nil
}

func (ja *JaxAuth[User]) verifyBodyWithHMAC(encryptionBody string) bool {
	parts := strings.Split(encryptionBody, ":")
	if len(parts) < 4 {
		return false
	}

	nonce, _ := hex.DecodeString(parts[0])
	expiration := parts[1]
	encryptedText, _ := hex.DecodeString(parts[2])
	hmacValue := parts[3]

	newHMAC := ja.hash(string(encryptedText) + string(nonce) + expiration + ja.GetHMACKey())
	return hmacValue == newHMAC
}

func (ja *JaxAuth[User]) hash(val string) string {
	h := sha256.New()
	h.Write([]byte(val))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func getNonce() []byte {
	nonce := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		panic(err)
	}
	return nonce
}
func (ja *JaxAuth[User]) GenerateSalt() string {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		panic(err)
	}
	return base64.StdEncoding.EncodeToString(salt)
}

func parseCookieHeader(cookieHeader string) map[string]string {
	result := make(map[string]string)
	pairs := strings.Split(cookieHeader, "; ")

	for _, pair := range pairs {
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) == 2 {
			result[parts[0]] = parts[1]
		}
	}

	return result
}
