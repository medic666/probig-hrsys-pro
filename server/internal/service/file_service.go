package service

import (
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"time"

	"probig/server/internal/dao"
	"probig/server/internal/model"
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
	return dao.DB.Unscoped().Model(&model.File{}).Where("id = ?", id).Update("deleted_at", nil).Error
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
		switch r.TargetType {
		case "person":
			dao.DB.Table("persons").Select("name").Where("id = ?", r.TargetID).Scan(&assoc.TargetName)
		case "company":
			dao.DB.Table("companies").Select("name").Where("id = ?", r.TargetID).Scan(&assoc.TargetName)
		default:
			assoc.TargetName = fmt.Sprintf("%s#%d", r.TargetType, r.TargetID)
		}
		result = append(result, assoc)
	}
	return result, nil
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

