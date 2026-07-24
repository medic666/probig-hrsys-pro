package handlers

import (
	"fmt"
	"net/http"

	"probig/middleware"
	"probig/services"
	"probig/utils"

	"github.com/gin-gonic/gin"
)

func Login(c *gin.Context) {
	var input services.LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ErrBadRequest(c, "参数错误")
		return
	}
	result, err := services.Login(input)
	if err != nil {
		utils.ErrBadRequest(c, err.Error())
		return
	}
	utils.Success(c, result)
}

func GetUserInfo(c *gin.Context) {
	userID := middleware.GetUserID(c)
	user, err := services.GetUser(userID)
	if err != nil {
		utils.ErrBadRequest(c, "用户不存在")
		return
	}
	utils.Success(c, gin.H{
		"user_id":        user.ID,
		"username":       user.Username,
		"is_first_login": user.IsFirstLogin,
		"status":         user.Status,
		"person_id":      user.PersonID,
		"person":         user.Person,
		"roles":          user.Roles,
		"permissions":    middleware.GetUserPermissions(userID),
	})
}

func ChangePassword(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var input struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ErrBadRequest(c, "参数错误")
		return
	}
	if err := services.ChangePassword(userID, input.OldPassword, input.NewPassword); err != nil {
		utils.ErrBadRequest(c, err.Error())
		return
	}
	utils.Success(c, nil)
}

func ResetPassword(c *gin.Context) {
	type req struct {
		UserID      uint   `json:"user_id"`
		NewPassword string `json:"new_password"`
	}
	var input req
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ErrBadRequest(c, "参数错误")
		return
	}
	if err := services.ResetPassword(input.UserID, input.NewPassword); err != nil {
		utils.ErrBadRequest(c, err.Error())
		return
	}
	utils.Success(c, nil)
}

func DownloadFile(c *gin.Context) {
	id := c.Param("id")
	var fid uint
	if _, err := fmt.Sscanf(id, "%d", &fid); err != nil {
		utils.ErrBadRequest(c, "参数错误")
		return
	}
	file, err := services.GetFile(fid)
	if err != nil {
		utils.ErrBadRequest(c, "文件不存在")
		return
	}
	c.Header("Content-Disposition", "attachment; filename="+file.Name)
	c.Data(http.StatusOK, file.MimeType, file.Content)
}

func PublicDownloadFile(c *gin.Context) {
	id := c.Param("id")
	var fid uint
	if _, err := fmt.Sscanf(id, "%d", &fid); err != nil {
		utils.ErrBadRequest(c, "参数错误")
		return
	}
	file, err := services.GetFile(fid)
	if err != nil {
		utils.ErrBadRequest(c, "文件不存在")
		return
	}
	c.Header("Content-Disposition", "attachment; filename="+file.Name)
	c.Data(http.StatusOK, file.MimeType, file.Content)
}
