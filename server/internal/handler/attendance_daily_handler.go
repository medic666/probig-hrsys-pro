package handler

import (
	"strconv"

	"probig/server/internal/service"
	"probig/server/internal/utils"

	"github.com/gin-gonic/gin"
)

func GetDailyProjections(c *gin.Context) {
	personID, _ := strconv.ParseUint(c.Query("person_id"), 10, 64)
	list, err := service.GetDailyProjections(uint(personID), c.Query("date_start"), c.Query("date_end"))
	if err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.Success(c, list)
}

func GetEventsByPersonDate(c *gin.Context) {
	personID, _ := strconv.ParseUint(c.Param("personId"), 10, 64)
	date := c.Param("date")
	events, err := service.GetAttendanceEventsByPersonDate(uint(personID), date)
	if err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.Success(c, events)
}
