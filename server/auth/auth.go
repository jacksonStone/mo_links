package auth

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"mo_links/auth/jaxauth"
	"mo_links/common"
	"mo_links/db"
	"os"
	"strconv"
)

var jaxAuthInstance *jaxauth.JaxAuth[common.User]

func GetCookieName() string {
	return jaxAuthInstance.CookieName
}
func AttemptLoginAndGetCookie(userId int64, plainTextPassword string) (string, error) {
	return jaxAuthInstance.AttemptLoginAndGetCookie(strconv.FormatInt(userId, 10), plainTextPassword)
}
func AttemptCookieDecryption(rawCookieHeader string) (common.TrimmedUser, error) {
	decryptedCookie, err := jaxAuthInstance.AttemptCookieDecryption(rawCookieHeader)
	if err != nil {
		return common.TrimmedUser{}, err
	}
	// parse cookie as json
	var user common.TrimmedUser
	err = json.Unmarshal([]byte(decryptedCookie), &user)
	if err != nil {
		return common.TrimmedUser{}, err
	}
	return user, nil
}
func GenerateSalt() string {
	return jaxAuthInstance.GenerateSalt()
}
func GenerateUrlSafeToken() string {
	randBytes := make([]byte, 32)
	if _, err := rand.Read(randBytes); err != nil {
		panic(err)
	}
	return base64.RawURLEncoding.EncodeToString(randBytes)
}
func GetHashedPasswordFromRawTextPassword(plainTextPassword string, salt string) string {
	return jaxAuthInstance.GetHashedPasswordFromRawTextPassword(plainTextPassword, salt)
}
func CreateNewCookie(user common.User) (string, error) {
	return jaxAuthInstance.CreateNewCookie(user)
}
func InitAuth() {
	jaxAuthInstance = jaxauth.NewJaxAuth[common.User]()
	jaxAuthInstance.GetUserPasswordSaltField = func(user common.User) string {
		return user.Salt
	}
	jaxAuthInstance.GetUserHashPasswordField = func(user common.User) string {
		return user.HashedPassword
	}
	// HMACKey and EncryptionKey are stored in hex in the environment variables
	// EncryptionKey is 16 bytes long, HMACKey is 32 bytes long
	jaxAuthInstance.GetHMACKey = func() string {
		return os.Getenv("HMAC_KEY")[:32]
	}
	jaxAuthInstance.GetEncryptionSecret = func() string {
		return os.Getenv("ENCRYPTION_KEY")[:16]
	}
	jaxAuthInstance.CreateRawCookieContents = func(user common.User) string {
		cookieContents, _ := json.Marshal(common.TrimmedUser{
			Id:                   user.Id,
			Email:                user.Email,
			ActiveOrganizationId: user.ActiveOrganizationId,
			VerifiedEmail:        user.VerifiedEmail,
		})
		return string(cookieContents)
	}
	jaxAuthInstance.GetUser = func(userId string) (common.User, error) {
		id, err := strconv.Atoi(userId)
		if err != nil {
			return common.User{}, err
		}
		return db.DbGetUser(int64(id))
	}
	// TODO: Maybe a better signal is needed, but since I have node all over the place, this is the easiest way to do it
	jaxAuthInstance.UseDevCookie = os.Getenv("NODE_ENV") == "development"
}
