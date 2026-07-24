package dao

import (
	"time"

	"probig/internal/models"
)

func GetPositionSnapshotsByPersonID(personID uint) ([]models.PositionSnapshot, error) {
	var list []models.PositionSnapshot
	if err := DB().Where("person_id = ?", personID).Order("snapshot_date ASC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func GetPositionSnapshot(personID uint, date time.Time) (*models.PositionSnapshot, error) {
	var s models.PositionSnapshot
	d := date.Format("2006-01-02")
	if err := DB().Where("person_id = ? AND snapshot_date = ?", personID, d).First(&s).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

func CreatePositionSnapshots(snapshots []models.PositionSnapshot) error {
	if len(snapshots) == 0 {
		return nil
	}
	return DB().Create(&snapshots).Error
}

func DeletePositionSnapshotsByPersonID(personID uint) error {
	return DB().Where("person_id = ?", personID).Delete(&models.PositionSnapshot{}).Error
}
