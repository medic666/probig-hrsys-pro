package file

import (
	"mime/multipart"
	"time"

	"probig/internal/pkg/database"
)

type File = database.File
type FileRelation = database.FileRelation

type FileFilter struct {
	FileName string `form:"file_name"`
	FileType string `form:"file_type"`
	PageNum  int    `form:"page_num"`
	PageSize int    `form:"page_size"`
}

type FileVO struct {
	ID         uint       `json:"id"`
	FileName   string     `json:"file_name"`
	FileType   string     `json:"file_type"`
	FileSize   int64      `json:"file_size"`
	FilePath   string     `json:"file_path"`
	UploaderID uint       `json:"uploader_id"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

type UploadRequest struct {
	File       *multipart.FileHeader `form:"file" binding:"required"`
	TargetType string                `form:"target_type"`
	TargetID   uint                  `form:"target_id"`
}

type RelationUpdateRequest struct {
	Add    []RelationTarget `json:"add"`
	Remove []RelationTarget `json:"remove"`
}

type RelationTarget struct {
	TargetType string `json:"target_type"`
	TargetID   uint   `json:"target_id"`
}

type FileRelationVO struct {
	ID         uint   `json:"id"`
	TargetType string `json:"target_type"`
	TargetID   uint   `json:"target_id"`
	TargetName string `json:"target_name"`
}
