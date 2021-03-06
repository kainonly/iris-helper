package tokenx

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"time"
)

var (
	def = tokenOption{
		key:    nil,
		method: jwt.SigningMethodHS256,
	}
)

type tokenOption struct {
	key    []byte
	method jwt.SigningMethod
}

// Load signing key
//	@param `value` []byte
func LoadKey(value []byte) {
	def.key = value
}

// Set signature method
//	@param `value` jwt.SigningMethod
func SigningMethod(value jwt.SigningMethod) {
	def.method = value
}

type Token struct {
	Value  string
	Claims jwt.MapClaims
}

// create a token
// 	@param `claims` jwt.MapClaims http://self-issued.info/docs/draft-ietf-oauth-json-web-token.html#Claims
// 	@param `expires` time.Duration
// 	@return `token` *Token
func Make(claims jwt.MapClaims, expires time.Duration) (token *Token, err error) {
	token = new(Token)
	claims["jti"] = uuid.New()
	claims["iat"] = time.Now().Unix()
	claims["exp"] = time.Now().Add(expires).Unix()
	ref := jwt.NewWithClaims(def.method, claims)
	token.Claims = ref.Claims.(jwt.MapClaims)
	if token.Value, err = ref.SignedString(def.key); err != nil {
		return
	}
	return
}

// token refresh logic
// 	@param `claims` jwt.MapClaims
// 	@return jwt.MapClaims
type RefreshHandle func(claims jwt.MapClaims) (jwt.MapClaims, error)

// verify that the token is valid
// 	@param `value` string
// 	@param `refresh` RefreshHandle
// 	@return `claims` jwt.MapClaims
func Verify(value string, refresh RefreshHandle) (claims jwt.MapClaims, err error) {
	var token *jwt.Token
	if token, err = jwt.Parse(value, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return def.key, nil
	}); err != nil {
		if ve, ok := err.(*jwt.ValidationError); ok {
			if ve.Errors == jwt.ValidationErrorExpired && refresh != nil {
				if token != nil {
					return refresh(token.Claims.(jwt.MapClaims))
				}
			}
		}
		return
	}
	return token.Claims.(jwt.MapClaims), nil
}
