package commonhttp

import (
	"context"

	"github.com/gin-gonic/gin"
)

type User struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

type userCtxKey int

const UserCtxKey userCtxKey = iota

func UserFromContext(ctx context.Context) (*User, bool) {
	c, ok := ctx.(*gin.Context)
	if !ok {
		return nil, false
	}
	u, ok := c.Get(UserCtxKey)
	if !ok {
		return nil, false
	}
	user, ok := u.(*User)
	return user, ok
}

func GatewayUserAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		user := &User{
			ID:    c.GetHeader("X-Forwarded-ID"),
			Email: c.GetHeader("X-Forwarded-Email"),
		}
		c.Set(UserCtxKey, user)
		c.Next()
	}
}
