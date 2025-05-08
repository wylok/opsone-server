package middleware

import (
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
	"inner/conf/platform_conf"
	"inner/modules/databases"
	"inner/modules/kits"
	"net/http"
	"time"
)

var (
	log = kits.Log{}
)

func GenerateToken(userId, roles, userName, nickName, departmentId, secret string, expiresIn time.Duration) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := make(jwt.MapClaims)
	claims["exp"] = time.Now().Add(expiresIn).Unix() // 设置超时时间
	claims["iat"] = time.Now().Unix()
	claims["user_id"] = userId
	claims["user_name"] = userName
	claims["nick_name"] = nickName
	claims["department_id"] = departmentId
	claims["roles"] = roles
	token.Claims = claims
	tokenString, err := token.SignedString([]byte(secret))
	return tokenString, err
}

func ParseToken(token, secret string) (map[string]string, error) {
	Token, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		var err *jwt.ValidationError
		if errors.As(err, &err) {
			if err.Errors&jwt.ValidationErrorMalformed != 0 {
				return nil, err
			}
			if err.Errors&(jwt.ValidationErrorExpired|jwt.ValidationErrorNotValidYet) != 0 {
				return nil, err
			}
		}
	}
	if Token != nil {
		claims := Token.Claims.(jwt.MapClaims)
		tm := time.Unix(int64(claims["exp"].(float64)), 0)
		return map[string]string{"user_id": claims["user_id"].(string), "user_name": claims["user_name"].(string),
			"nick_name": claims["nick_name"].(string), "department_id": claims["department_id"].(string), "token": token,
			"roles":       claims["roles"].(string),
			"expire_time": tm.Format("2006-01-02 15:04:05")}, nil
	}
	return nil, err
}

func VerifyToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		var (
			token string
			err   error
			Token []databases.Token
			Data  = map[string]string{}
			d     = map[string]interface{}{}
		)
		defer func() {
			if r := recover(); r != nil {
				err = errors.New(fmt.Sprint(r))
			}
			if err != nil {
				log.Error(err)
			}
		}()
		if c != nil {
			//从url取token
			token = c.Request.FormValue("token")
			//从body取token
			if token == "" {
				token = c.Request.PostFormValue("token")
			}
			//从header取token
			if token == "" {
				token = c.GetHeader("token")
			}
			//从cookie取token
			if token == "" {
				token, _ = c.Cookie("token")
			}
			if token != "" {
				TokenKey := "auth_token_verify_" + kits.MD5(token)
				if token == platform_conf.PublicToken {
					d = map[string]interface{}{"user_id": "platform", "roles": "platform",
						"token": platform_conf.PublicToken, "expire_time": "2999-12-31 23:59:00"}
				} else {
					if rc.Exists(ctx, TokenKey).Val() == 1 {
						for k, v := range rc.HGetAll(ctx, TokenKey).Val() {
							d[k] = v
						}
					} else {
						Data, err = ParseToken(token, platform_conf.CryptKey)
						if len(Data) > 0 {
							sql := "join users on users.user_id=token.user_id and users.user_id=? and users.status=?"
							db.Joins(sql, c.GetString("user_id"), "active")
							db.Where("token.token=? and token.expire_at>=?", token, time.Now()).Find(&Token)
							if len(Token) > 0 {
								for k, v := range Data {
									d[k] = v
									rc.HSet(ctx, TokenKey, k, v)
								}
								rc.Expire(ctx, TokenKey, 5*time.Minute)
							}
						}
					}
				}
				if len(d) > 0 {
					c.Set("user_id", cast.ToString(d["user_id"]))
					c.Set("user_name", cast.ToString(d["user_name"]))
					c.Set("roles", cast.ToString(d["roles"]))
					c.Set("token", token)
					c.Next()
					return
				}
			} else {
				log.Error(c.Request.RequestURI + "无法获取token")
			}
		}
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false, "message": "Unauthorized", "data": map[string]string{}})
		c.Abort()
		return
	}
}
