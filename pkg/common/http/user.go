package commonhttp

import (
	"context"
	"encoding/json"
	"slices"

	"github.com/gin-gonic/gin"
)

type User struct {
	ID     string   `json:"id"`
	Email  string   `json:"email"`
	Name   string   `json:"name"`
	Groups []string `json:"groups"`
	Roles  []string `json:"roles"`
}

func (u *User) HasGroup(group string) bool {
	return slices.Contains(u.Groups, group)
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

func unmarshalHeaderList(v string) []string {
	var list []string
	if v == "" {
		return []string{}
	}
	if err := json.Unmarshal([]byte(v), &list); err != nil {
		return []string{}
	}
	return list
}

func GatewayUserAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		user := &User{
			ID:     c.GetHeader("X-Forwarded-ID"),
			Email:  c.GetHeader("X-Forwarded-Email"),
			Name:   c.GetHeader("X-Forwarded-Name"),
			Groups: unmarshalHeaderList(c.GetHeader("X-Forwarded-Groups")),
			Roles:  unmarshalHeaderList(c.GetHeader("X-Forwarded-Roles")),
		}
		c.Set(UserCtxKey, user)
		c.Next()
	}
}
