package file

import "gorm.io/gorm"

type File struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	Name       string         `gorm:"type:varchar(128);not null" json:"name"`
	Size       int64          `json:"size"`
	MimeType   string         `gorm:"type:varchar(64)" json:"mimeType"`
	Content    []byte         `gorm:"type:blob" json:"-"`
	UploaderID uint           `json:"uploaderId"`
	CreatedAt  string         `gorm:"type:datetime" json:"createdAt"`
	UpdatedAt  string         `gorm:"type:datetime" json:"updatedAt"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

func (File) TableName() string { return "files" }

type FileRelation struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	FileID     uint           `gorm:"not null;index" json:"fileId"`
	TargetType string         `gorm:"type:varchar(32);not null" json:"targetType"`
	TargetID   uint           `gorm:"not null" json:"targetId"`
	CreatedAt  string         `gorm:"type:datetime" json:"createdAt"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

func (FileRelation) TableName() string { return "file_relations" }
