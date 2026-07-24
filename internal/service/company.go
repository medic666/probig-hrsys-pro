package service

import (
	"errors"

	"probig/internal/dao"
	"probig/internal/middleware"

	"github.com/gin-gonic/gin"
	"probig/internal/models"
)

func GetCompanyList(page, pageSize int, keyword string) ([]models.Company, int64, error) {
	return dao.GetCompanyList(page, pageSize, keyword)
}

func GetCompany(id uint) (*models.Company, error) {
	return dao.GetCompanyByID(id)
}

func CreateCompany(c *gin.Context, company *models.Company) error {
	if company.CreditCode != "" {
		existing, _ := dao.GetCompanyByCreditCode(company.CreditCode)
		if existing != nil {
			return errors.New("统一社会信用代码已存在")
		}
	}
	if err := dao.CreateCompany(company); err != nil {
		return err
	}
	middleware.RecordAudit(c, "新增", "company", company.ID, nil, company, "")
	return nil
}

func UpdateCompany(c *gin.Context, company *models.Company) error {
	old, err := dao.GetCompanyByID(company.ID)
	if err != nil {
		return errors.New("公司不存在")
	}
	if err := dao.UpdateCompany(company); err != nil {
		return err
	}
	middleware.RecordAudit(c, "修改", "company", company.ID, old, company, "")
	return nil
}

func DeleteCompany(c *gin.Context, id uint) error {
	company, err := dao.GetCompanyByID(id)
	if err != nil {
		return errors.New("公司不存在")
	}
	if err := dao.SoftDeleteCompany(id); err != nil {
		return err
	}
	middleware.RecordAudit(c, "删除", "company", id, company, nil, "")
	return nil
}

func RestoreCompany(c *gin.Context, id uint) error {
	if err := dao.RestoreCompany(id); err != nil {
		return err
	}
	middleware.RecordAudit(c, "恢复", "company", id, nil, nil, "")
	return nil
}
