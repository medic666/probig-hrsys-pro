package middleware

import (
	"probig/database"
	"probig/models"
	"probig/utils"

	"github.com/gin-gonic/gin"
)

func PermissionMiddleware(permKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := GetUserID(c)
		if userID == 0 {
			c.Next()
			return
		}

		var count int64
		database.DB.Table("user_roles ur").
			Joins("JOIN role_permissions rp ON ur.role_id = rp.role_id").
			Joins("JOIN permissions p ON rp.permission_id = p.id").
			Where("ur.user_id = ? AND p.permission_key = ? AND ur.deleted_at IS NULL AND rp.deleted_at IS NULL", userID, permKey).
			Count(&count)

		if count == 0 {
			utils.ErrForbidden(c, "权限不足")
			c.Abort()
			return
		}
		c.Next()
	}
}

func GetUserPermissions(userID uint) []string {
	var perms []models.Permission
	database.DB.Table("permissions p").
		Joins("JOIN role_permissions rp ON p.id = rp.permission_id").
		Joins("JOIN user_roles ur ON rp.role_id = ur.role_id").
		Where("ur.user_id = ? AND ur.deleted_at IS NULL AND rp.deleted_at IS NULL", userID).
		Group("p.id").
		Find(&perms)

	var keys []string
	for _, p := range perms {
		keys = append(keys, p.PermissionKey)
	}
	return keys
}
