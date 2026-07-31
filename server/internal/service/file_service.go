package service

import (
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"time"

	"probig/server/internal/dao"
	"probig/server/internal/model"
	"probig/server/internal/utils"

	"gorm.io/gorm"
)

func GetFileList(pageNum, pageSize int, name string) ([]model.File, int64, error) {
	tx := dao.DB.Model(&model.File{})
	if name != "" {
		tx = tx.Where("original_name LIKE ?", "%"+name+"%")
	}
	var total int64
	tx.Count(&total)
	var list []model.File
	offset := (pageNum - 1) * pageSize
	err := tx.Order("id DESC").Offset(offset).Limit(pageSize).Find(&list).Error
	return list, total, err
}

func DeleteFileByID(id uint) error {
	return dao.DB.Delete(&model.File{}, id).Error
}

func RestoreFileByID(id uint) error {
	return utils.WithTransaction(dao.DB, func(tx *gorm.DB) error {
		if err := tx.Unscoped().Model(&model.File{}).Where("id = ?", id).Update("deleted_at", nil).Error; err != nil {
			return err
		}
		tx.Unscoped().Model(&model.FileRelation{}).Where("file_id = ?", id).Update("deleted_at", nil)
		return nil
	})
}

func GetDeletedFileList(pageNum, pageSize int) ([]model.File, int64, error) {
	var list []model.File
	var total int64
	tx := dao.DB.Unscoped().Model(&model.File{}).Where("deleted_at IS NOT NULL")
	tx.Count(&total)
	offset := (pageNum - 1) * pageSize
	tx.Offset(offset).Limit(pageSize).Order("deleted_at DESC").Find(&list)
	return list, total, nil
}

func ComputeFileMD5(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := md5.New()
	io.Copy(h, f)
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

func FindFileByMD5(md5hash string) (*model.File, error) {
	var f model.File
	err := dao.DB.Where("md5 = ? AND deleted_at IS NULL", md5hash).First(&f).Error
	if err != nil {
		return nil, err
	}
	return &f, nil
}

func PermanentDeleteFile(id uint) (int64, error) {
	var count int64
	dao.DB.Model(&model.FileRelation{}).Where("file_id = ?", id).Count(&count)
	if count > 0 {
		return count, fmt.Errorf("该文件仍被 %d 个实体使用", count)
	}
	var f model.File
	if err := dao.DB.First(&f, id).Error; err != nil {
		return 0, err
	}
	os.Remove(f.Path)
	return 0, dao.DB.Unscoped().Delete(&f).Error
}

type FileAssociation struct {
	TargetType string `json:"target_type"`
	TargetID   uint   `json:"target_id"`
	TargetName string `json:"target_name"`
}

func GetFileAssociations(fileID uint) ([]FileAssociation, error) {
	var relations []model.FileRelation
	dao.DB.Where("file_id = ?", fileID).Find(&relations)
	var result []FileAssociation
	for _, r := range relations {
		assoc := FileAssociation{TargetType: r.TargetType, TargetID: r.TargetID}
		assoc.TargetName = resolveTargetName(r.TargetType, r.TargetID)
		result = append(result, assoc)
	}
	return result, nil
}

func resolveTargetName(targetType string, targetID uint) string {
	switch targetType {
	case "person":
		var name string
		dao.DB.Table("persons").Select("name").Where("id = ? AND deleted_at IS NULL", targetID).Scan(&name)
		if name != "" { return "人员: " + name }
	case "company":
		var name string
		dao.DB.Table("companies").Select("name").Where("id = ? AND deleted_at IS NULL", targetID).Scan(&name)
		if name != "" { return "公司: " + name }
	case "position_event":
		var row struct{ PersonID uint; EventType string; EffectiveDate time.Time }
		dao.DB.Table("position_events").Select("person_id, event_type, effective_date").Where("id = ? AND deleted_at IS NULL", targetID).Scan(&row)
		if row.PersonID > 0 {
			var pn string; dao.DB.Table("persons").Select("name").Where("id = ? AND deleted_at IS NULL", row.PersonID).Scan(&pn)
			return fmt.Sprintf("职务事件: %s 的 %s (%s)", pn, row.EventType, row.EffectiveDate.AddDate(0,0,0).Format("2006-01-02"))
		}
	case "attendance_event":
		var row struct{ DailyID uint; EventType string; SubType string }
		dao.DB.Table("attendance_event_details").Select("daily_id, event_type, sub_type").Where("id = ? AND deleted_at IS NULL", targetID).Scan(&row)
		if row.DailyID > 0 {
			var dailyRow struct{ PersonID uint; EventDate time.Time }
			dao.DB.Table("attendance_daily").Select("person_id, event_date").Where("id = ? AND deleted_at IS NULL", row.DailyID).Scan(&dailyRow)
			if dailyRow.PersonID > 0 {
				var pn string; dao.DB.Table("persons").Select("name").Where("id = ? AND deleted_at IS NULL", dailyRow.PersonID).Scan(&pn)
				return fmt.Sprintf("考勤事件: %s 的 %s-%s (%s)", pn, row.EventType, row.SubType, dailyRow.EventDate.AddDate(0,0,0).Format("2006-01-02"))
			}
		}
	case "annual_leave_event":
		var row struct{ PersonID uint; EventType string; EffectiveDate time.Time }
		dao.DB.Table("annual_leave_account_events").Select("person_id, event_type, effective_date").Where("id = ? AND deleted_at IS NULL", targetID).Scan(&row)
		if row.PersonID > 0 {
			var pn string; dao.DB.Table("persons").Select("name").Where("id = ? AND deleted_at IS NULL", row.PersonID).Scan(&pn)
			return fmt.Sprintf("年假事件: %s 的 %s (%s)", pn, row.EventType, row.EffectiveDate.AddDate(0,0,0).Format("2006-01-02"))
		}
	case "salary_event":
		var row struct{ PersonID uint; EventType string; BelongMonth string }
		dao.DB.Table("salary_events").Select("person_id, event_type, belong_month").Where("id = ? AND deleted_at IS NULL", targetID).Scan(&row)
		if row.PersonID > 0 {
			var pn string; dao.DB.Table("persons").Select("name").Where("id = ? AND deleted_at IS NULL", row.PersonID).Scan(&pn)
			return fmt.Sprintf("工资事件: %s 的 %s (%s)", pn, row.EventType, row.BelongMonth)
		}
	}
	return fmt.Sprintf("%s#%d", targetType, targetID)
}

type FileTargetResult struct {
	Relation model.FileRelation `json:"relation"`
	File     model.File         `json:"file"`
}

func GetFilesForTarget(targetType string, targetID uint) ([]FileTargetResult, error) {
	if !targetExists(targetType, targetID) {
		return []FileTargetResult{}, nil
	}
	var relations []model.FileRelation
	dao.DB.Where("target_type = ? AND target_id = ?", targetType, targetID).Find(&relations)
	var fileIDs []uint
	for _, r := range relations { fileIDs = append(fileIDs, r.FileID) }
	var files []model.File
	if len(fileIDs) > 0 { dao.DB.Where("id IN ?", fileIDs).Find(&files) }
	var list []FileTargetResult
	for _, rel := range relations { for _, f := range files { if f.ID == rel.FileID { list = append(list, FileTargetResult{rel, f}) } } }
	return list, nil
}

var targetTableNames = map[string]string{
	"person":              "persons",
	"company":             "companies",
	"position_event":      "position_events",
	"attendance_event":    "attendance_events",
	"annual_leave_event":  "annual_leave_account_events",
	"salary_event":        "salary_events",
}

func targetExists(targetType string, targetID uint) bool {
	table := targetTableNames[targetType]
	if table == "" { return true }
	var count int64
	dao.DB.Table(table).Where("id = ?", targetID).Count(&count)
	return count > 0
}

func CountFileAssociations(fileID uint) int64 {
	var count int64
	dao.DB.Model(&model.FileRelation{}).Where("file_id = ?", fileID).Count(&count)
	return count
}

func CleanOrphanFiles() (int, error) {
	deadline := time.Now().AddDate(0, 0, -30)
	var files []model.File
	dao.DB.Unscoped().Where("deleted_at IS NOT NULL AND deleted_at < ?", deadline).Find(&files)
	count := 0
	for _, f := range files {
		var relCount int64
		dao.DB.Model(&model.FileRelation{}).Where("file_id = ?", f.ID).Count(&relCount)
		if relCount == 0 {
			os.Remove(f.Path)
			dao.DB.Unscoped().Delete(&f)
			count++
		}
	}
	return count, nil
}

