package file

import (
	"encoding/json"

	"gorm.io/gorm"

	"probig/internal/pkg/audit"
)

type Service struct {
	dao *DAO
}

var globalService *Service

func NewService(db *gorm.DB) *Service {
	svc := &Service{dao: NewDAO(db)}
	globalService = svc
	return svc
}

func GetService() *Service {
	return globalService
}

func (s *Service) Upload(name string, size int64, mimeType string, content []byte, uploaderID uint) (*File, error) {
	file := &File{
		Name:       name,
		Size:       size,
		MimeType:   mimeType,
		Content:    content,
		UploaderID: uploaderID,
	}
	if err := s.dao.Create(file); err != nil {
		return nil, err
	}

	afterJSON, _ := json.Marshal(File{
		ID:         file.ID,
		Name:       file.Name,
		Size:       file.Size,
		MimeType:   file.MimeType,
		UploaderID: file.UploaderID,
		CreatedAt:  file.CreatedAt,
	})
	audit.GlobalAuditService.Log(uploaderID, "", "file", file.ID, "create", "", string(afterJSON), "", "")
	return file, nil
}

func (s *Service) GetFile(id uint) (*File, error) {
	return s.dao.GetByID(id)
}

func (s *Service) GetFileInfo(id uint) (*File, error) {
	return s.dao.GetByIDWithoutContent(id)
}

func (s *Service) ListFiles(page, pageSize int, name, mimeType string) ([]File, int64, error) {
	return s.dao.List(page, pageSize, name, mimeType)
}

func (s *Service) DeleteFile(id uint) error {
	before, _ := s.dao.GetByIDWithoutContent(id)
	beforeJSON, _ := json.Marshal(before)

	if err := s.dao.DeleteFile(id); err != nil {
		return err
	}

	audit.GlobalAuditService.Log(0, "", "file", id, "delete", string(beforeJSON), "", "", "")
	return nil
}

func (s *Service) RestoreFile(id uint) error {
	if err := s.dao.RestoreFile(id); err != nil {
		return err
	}

	audit.GlobalAuditService.Log(0, "", "file", id, "restore", "", "", "", "")
	return nil
}

func (s *Service) AddRelation(fileID, targetID uint, targetType string) error {
	relation := &FileRelation{
		FileID:     fileID,
		TargetType: targetType,
		TargetID:   targetID,
	}
	if err := s.dao.CreateRelation(relation); err != nil {
		return err
	}

	afterJSON, _ := json.Marshal(relation)
	audit.GlobalAuditService.Log(0, "", "file_relation", relation.ID, "create", "", string(afterJSON), "", "")
	return nil
}

func (s *Service) RemoveRelation(relationID uint) error {
	if err := s.dao.DeleteRelation(relationID); err != nil {
		return err
	}

	audit.GlobalAuditService.Log(0, "", "file_relation", relationID, "delete", "", "", "", "")
	return nil
}

func (s *Service) GetFilesByTarget(targetType string, targetID uint) ([]File, error) {
	return s.dao.GetFilesByTarget(targetType, targetID)
}

func (s *Service) AssociateFilesWithTarget(fileIDs []uint, targetType string, targetID uint) error {
	var relations []FileRelation
	for _, fileID := range fileIDs {
		relations = append(relations, FileRelation{
			FileID:     fileID,
			TargetType: targetType,
			TargetID:   targetID,
		})
	}
	return s.dao.BatchCreateRelations(relations)
}

func (s *Service) DeleteRelationsByTarget(targetType string, targetID uint) error {
	return s.dao.DeleteRelationsByTarget(targetType, targetID)
}
