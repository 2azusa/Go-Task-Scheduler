package middlerware

import (
	"errors"
	"pulse/admin/internal/model/resp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

var JwtKey = []byte("45df45rds4")

var (
	ErrTokenExpired     = errors.New("token is expired")
	ErrTokenNotValidYet = errors.New("token not active yet")
	ErrTokenMalformed   = errors.New("that's not even a token")
	ErrTokenInvalid     = errors.New("couldn't handle this token")
)

type MyClaims struct {
	ID       int
	UserName string
	jwt.RegisteredClaims
}

// SetToken
func SetToken(id int, username string) (string, error) {
	expireTime := time.Now().Add(24 * time.Hour)
	claims := MyClaims{
		id,
		username,
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expireTime),
			Issuer:    "pulse",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(JwtKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// CheckToken
func CheckToken(tokenString string) (*MyClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &MyClaims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("非预期的签名算法")
		}
		return JwtKey, nil
	})
	if err != nil {
		var ve *jwt.ValidationError
		if errors.As(err, &ve) {
			if ve.Errors&jwt.ValidationErrorMalformed != 0 {
				return nil, ErrTokenMalformed // Token格式错误
			} else if ve.Errors&jwt.ValidationErrorExpired != 0 {
				return nil, ErrTokenExpired // Token已过期
			} else if jwt.ValidationErrorNotValidYet != 0 {
				return nil, ErrTokenNotValidYet // Token尚未激活
			}
		}
		return nil, ErrTokenInvalid // 无法处理这个Token
	}

	if claims, ok := token.Claims.(*MyClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrTokenInvalid
}

// JwtToken
func JwtToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenHeader := c.Request.Header.Get("Authorization")
		if tokenHeader == "" {
			// c.JSON(http.StatusUnauthorized, gin.H{"error": "请求未携带token"})
			resp.FailWithDetailed(resp.ERROR, gin.H{"reload": true}, "请求未携带token", c)
			c.Abort()
			return
		}

		parts := strings.SplitN(tokenHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			resp.FailWithDetailed(resp.ERROR, gin.H{"reload": true}, "token格式不正确", c)
			c.Abort()
			return
		}

		claims, err := CheckToken(parts[1])
		if err != nil {
			resp.FailWithDetailed(resp.ERROR, gin.H{"reload": true}, "无效的token", c)
			c.Abort()
			return
		}

		c.Set("userID", claims.ID)
		c.Set("username", claims.UserName)
		c.Next()
	}
}
