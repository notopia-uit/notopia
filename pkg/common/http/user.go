package commonhttp

import (
	"context"
	"strings"

	"github.com/gin-gonic/gin"
)

type UserList []string

type User struct {
	ID     string   `json:"id"`
	Email  string   `json:"email"`
	Groups UserList `json:"groups"`
	Roles  UserList `json:"roles"`
}

func (u *UserList) UnmarshalHeader(headerValue string) {
	s := strings.TrimPrefix(headerValue, "[")
	s = strings.TrimSuffix(s, "]")
	s = strings.TrimSpace(s)

	if s == "" {
		*u = []string{}
		return
	}

	*u = strings.Fields(s)
}

type userCtxKey int

const UserCtxKey userCtxKey = iota

func userToContext(ctx context.Context, user *User) context.Context {
	return context.WithValue(ctx, UserCtxKey, user)
}

func UserFromContext(ctx context.Context) (*User, bool) {
	u, ok := ctx.Value(UserCtxKey).(*User)
	return u, ok
}

func GatewayUserAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader("X-Forwarded-ID")
		email := c.GetHeader("X-Forwarded-Email")
		rawGroups := c.GetHeader("X-Forwarded-Groups")
		rawRoles := c.GetHeader("X-Forwarded-Roles")

		user := &User{
			ID:     id,
			Email:  email,
			Groups: nil,
			Roles:  nil,
		}

		user.Groups.UnmarshalHeader(rawGroups)
		user.Roles.UnmarshalHeader(rawRoles)

		c.Set(UserCtxKey, user)

		c.Request = c.Request.WithContext(
			userToContext(c.Request.Context(), user),
		)

		c.Next()
	}
}
