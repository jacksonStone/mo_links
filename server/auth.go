package main

import (
	"encoding/json"
	"os"
	"strconv"
)

type User struct {
	salt           string
	hashedPassword string
	id             int32
	email          string
}

type TrimmedUser struct {
	Id    int32  `json:"id"`
	Email string `json:"email"`
}

var Auth *JaxAuth[User]

func InitAuth() {
	Auth = NewJaxAuth[User]()
	Auth.GetUserPasswordSaltField = func(user User) string {
		return user.salt
	}
	Auth.GetUserHashPasswordField = func(user User) string {
		return user.hashedPassword
	}
	// HMACKey and EncryptionKey are stored in hex in the environment variables
	// EncryptionKey is 16 bytes long, HMACKey is 32 bytes long
	Auth.GetHMACKey = func() string {
		return os.Getenv("HMAC_KEY")[:32]
	}
	Auth.GetEncryptionSecret = func() string {
		return os.Getenv("ENCRYPTION_KEY")[:16]
	}
	Auth.CreateRawCookieContents = func(user User) string {
		cookieContents, _ := json.Marshal(TrimmedUser{
			Id:    user.id,
			Email: user.email,
		})
		return string(cookieContents)
	}
	Auth.GetUser = func(userId string) (User, error) {
		id, err := strconv.Atoi(userId)
		if err != nil {
			return User{}, err
		}
		return dbGetUser(int32(id))
	}
	// TODO: remove this
	Auth.UseDevCookie = os.Getenv("NODE_ENV") == "development"
}
