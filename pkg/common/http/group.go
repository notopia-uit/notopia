package commonhttp

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func RequireGroup(group string) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := UserFromContext(c)
		if !ok || !user.HasGroup(group) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"code":    "forbidden",
				"message": "requires group " + group,
			})
			return
		}
		c.Next()
	}
}
