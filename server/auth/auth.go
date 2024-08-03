package auth

import (
	"encoding/json"
	"mo_links/common"
	"mo_links/db"
	"mo_links/jaxauth"
	"os"
	"strconv"
)

var Auth *jaxauth.JaxAuth[common.User]

func InitAuth() {
	Auth = jaxauth.NewJaxAuth[common.User]()
	Auth.GetUserPasswordSaltField = func(user common.User) string {
		return user.Salt
	}
	Auth.GetUserHashPasswordField = func(user common.User) string {
		return user.HashedPassword
	}
	// HMACKey and EncryptionKey are stored in hex in the environment variables
	// EncryptionKey is 16 bytes long, HMACKey is 32 bytes long
	Auth.GetHMACKey = func() string {
		return os.Getenv("HMAC_KEY")[:32]
	}
	Auth.GetEncryptionSecret = func() string {
		return os.Getenv("ENCRYPTION_KEY")[:16]
	}
	Auth.CreateRawCookieContents = func(user common.User) string {
		cookieContents, _ := json.Marshal(common.TrimmedUser{
			Id:                   user.Id,
			Email:                user.Email,
			ActiveOrganizationId: user.ActiveOrganizationId,
		})
		return string(cookieContents)
	}
	Auth.GetUser = func(userId string) (common.User, error) {
		id, err := strconv.Atoi(userId)
		if err != nil {
			return common.User{}, err
		}
		return db.DbGetUser(int64(id))
	}
	// TODO: remove this
	Auth.UseDevCookie = os.Getenv("NODE_ENV") == "development"
}
