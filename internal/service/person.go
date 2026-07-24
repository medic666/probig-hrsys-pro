package service

import (
	"errors"

	"probig/internal/dao"
	"probig/internal/models"
	"probig/internal/middleware"

	"github.com/gin-gonic/gin"
)

func GetPersonList(page, pageSize int, keyword string) ([]models.Person, int64, error) {
	return dao.GetPersonList(page, pageSize, keyword)
}

func GetPerson(id uint) (*models.Person, error) {
	return dao.GetPersonByID(id)
}

func CreatePerson(c *gin.Context, p *models.Person) error {
	if p.IDCardPlain != "" {
		existing, _ := dao.GetPersonByIDCard(p.IDCardPlain)
		if existing != nil {
			return errors.New("身份证号已存在")
		}
	}
	if err := dao.CreatePerson(p); err != nil {
		return err
	}
	middleware.RecordAudit(c, "新增", "person", p.ID, nil, p, "")
	return nil
}

func UpdatePerson(c *gin.Context, p *models.Person) error {
	old, err := dao.GetPersonByID(p.ID)
	if err != nil {
		return errors.New("人员不存在")
	}
	if err := dao.UpdatePerson(p); err != nil {
		return err
	}
	middleware.RecordAudit(c, "修改", "person", p.ID, old, p, "")
	return nil
}

func DeletePerson(c *gin.Context, id uint) error {
	p, err := dao.GetPersonByID(id)
	if err != nil {
		return errors.New("人员不存在")
	}
	if err := dao.SoftDeletePerson(id); err != nil {
		return err
	}
	middleware.RecordAudit(c, "删除", "person", id, p, nil, "")
	return nil
}

func RestorePerson(c *gin.Context, id uint) error {
	if err := dao.RestorePerson(id); err != nil {
		return err
	}
	middleware.RecordAudit(c, "恢复", "person", id, nil, nil, "")
	return nil
}
