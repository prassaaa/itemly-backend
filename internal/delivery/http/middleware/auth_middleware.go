package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/prassaaa/itemly-backend/internal/delivery/http/dto"
	"github.com/prassaaa/itemly-backend/internal/model"
	"github.com/prassaaa/itemly-backend/internal/usecase"
	jwtutil "github.com/prassaaa/itemly-backend/pkg/jwt"
)

func JWTAuth(jwtService *jwtutil.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "authorization header is required"})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "invalid authorization header format"})
			return
		}

		claims, err := jwtService.ValidateToken(parts[1])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "invalid or expired token"})
			return
		}

		c.Set("userID", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)
		c.Next()
	}
}

func RoleAuth(roles ...model.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("role")
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, dto.ErrorResponse{Error: "role not found in context"})
			return
		}

		roleStr, ok := userRole.(string)
		if !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, dto.ErrorResponse{Error: "invalid role type"})
			return
		}

		for _, r := range roles {
			if model.Role(roleStr) == r {
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(http.StatusForbidden, dto.ErrorResponse{Error: "insufficient permissions"})
	}
}

func PermissionAuth(permUsecase usecase.PermissionUsecase, requiredPermissions ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("role")
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, dto.ErrorResponse{Error: "role not found in context"})
			return
		}

		roleStr, ok := userRole.(string)
		if !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, dto.ErrorResponse{Error: "invalid role type"})
			return
		}

		for _, perm := range requiredPermissions {
			if !permUsecase.HasPermission(roleStr, perm) {
				c.AbortWithStatusJSON(http.StatusForbidden, dto.ErrorResponse{Error: "insufficient permissions"})
				return
			}
		}

		c.Next()
	}
}
