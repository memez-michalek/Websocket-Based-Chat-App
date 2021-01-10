package webtoken

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/dgrijalva/jwt-go"
)

type JWTUser struct {
	username string
	email    string
	jwt.StandardClaims
}

var (
	//secretToken = []byte(os.Getenv("JWTTOKEN"))

	secretToken = []byte("u8x/A?D*G-KaPdSgVkYp3s6v9y$B&E)H")
)

func CreateToken(username string, email string) string {
	fmt.Println(username, email)
	encrypted := jwt.NewWithClaims(jwt.SigningMethodHS256, JWTUser{
		username,
		email,
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Duration(time.Minute * 5)).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
	})

	token, err := encrypted.SignedString(secretToken)
	if err != nil {
		log.Fatal(err)
	}
	return token
}

func Verify(token string) (string, string, error) {
	verifiedToken, err := jwt.ParseWithClaims(token, new(JWTUser), func(token *jwt.Token) (interface{}, error) {
		return secretToken, nil
	})
	fmt.Println(token)
	if err != nil {
		return "", "", errors.New("webtoken is expired")
	}

	if claim, err := verifiedToken.Claims.(*JWTUser); err && verifiedToken.Valid {
		if claim.ExpiresAt < time.Now().Unix() {
			return "", "", errors.New("token has expired")

		}
		return claim.email, claim.username, nil
	}
	return "", "", nil

}
