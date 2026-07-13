package middleware

import (
	"common/config"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

func TokenHandler(userId string) (string, error) {
	data := config.DataConfig.JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userId": userId,
		"exp":    time.Now().Add(time.Hour * time.Duration(1) * 2).Unix(),
		"iat":    time.Now().Unix(),
	})
	tokenString, err := token.SignedString([]byte(data.AppKey))
	return tokenString, err
}
func TokenGet(tokenString string) (jwt.MapClaims, error) {
	data := config.DataConfig.JWT
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("异常:Unexpected signing method: %v", token.Header["alg"])
		}

		// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
		return []byte(data.AppKey), nil
	})
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		fmt.Println(claims["foo"], claims["nbf"])
		return claims, nil
	} else {
		fmt.Println(err)
		return nil, err
	}

}
func GetTokenUserId(ctx context.Context) (int, error) {
	userId, ok := ctx.Value("userId").(int)
	if !ok {
		return 0, errors.New("用户id获取失败")
	}
	return userId, nil
}
