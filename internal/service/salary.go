package service

import (
	"probig/internal/dao"
	"probig/internal/middleware"
	"probig/internal/models"

	"github.com/gin-gonic/gin"
)

func GetSalaryEventList(page, pageSize int, personID uint, belongMonth, eventType string) ([]models.SalaryEvent, int64, error) {
	return dao.GetSalaryEventList(page, pageSize, personID, belongMonth, eventType)
}

func GetSalaryEvent(id uint) (*models.SalaryEvent, error) {
	return dao.GetSalaryEventByID(id)
}

func CreateSalaryEvent(c *gin.Context, e *models.SalaryEvent) error {
	if err := dao.CreateSalaryEvent(e); err != nil {
		return err
	}
	middleware.RecordAudit(c, "新增", "salary_event", e.ID, nil, e, "")
	return nil
}

func UpdateSalaryEvent(c *gin.Context, e *models.SalaryEvent) error {
	old, err := dao.GetSalaryEventByID(e.ID)
	if err != nil {
		return err
	}
	if err := dao.UpdateSalaryEvent(e); err != nil {
		return err
	}
	middleware.RecordAudit(c, "修改", "salary_event", e.ID, old, e, "")
	return nil
}

func DeleteSalaryEvent(c *gin.Context, id uint) error {
	e, err := dao.GetSalaryEventByID(id)
	if err != nil {
		return err
	}
	if err := dao.DeleteSalaryEvent(id); err != nil {
		return err
	}
	middleware.RecordAudit(c, "删除", "salary_event", id, e, nil, "")
	return nil
}
