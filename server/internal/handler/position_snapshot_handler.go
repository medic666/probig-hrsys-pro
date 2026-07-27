package handler

import (
	"strconv"

	"probig/server/internal/service"
	"probig/server/internal/utils"

	"github.com/gin-gonic/gin"
)

func GetPersonCurrentPosition(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	snapshot, err := service.GetCurrentPositionSnapshot(uint(id))
	if err != nil {
		utils.Error(c, "暂无职务信息")
		return
	}
	utils.Success(c, snapshot)
}

func GetPersonPositionHistory(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	history, err := service.GetPositionSnapshotHistory(uint(id))
	if err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.Success(c, history)
}
