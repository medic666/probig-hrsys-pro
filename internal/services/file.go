package services

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/medic666/probig/internal/models"
)

type FileService struct {
	db        *sqlx.DB
	audit     *AuditService
	uploadDir string
	maxSize   int64
}

func NewFileService(db *sqlx.DB, audit *AuditService, uploadDir string, maxSize int64) *FileService {
	return &FileService{db: db, audit: audit, uploadDir: uploadDir, maxSize: maxSize}
}

func (s *FileService) Upload(uploadedBy uint, originalName string, reader io.Reader, size int64, mimeType string, ip string) (*models.File, error) {
	if size > s.maxSize {
		return nil, fmt.Errorf("文件大小超过限制 %d bytes", s.maxSize)
	}

	ext := filepath.Ext(originalName)
	filename := uuid.New().String() + ext
	savePath := filepath.Join(s.uploadDir, filename)
	fullPath := savePath

	dst, err := os.Create(fullPath)
	if err != nil {
		return nil, fmt.Errorf("create file: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, reader); err != nil {
		os.Remove(fullPath)
		return nil, fmt.Errorf("write file: %w", err)
	}

	tx, err := s.db.Beginx()
	if err != nil {
		os.Remove(fullPath)
		return nil, err
	}
	defer tx.Rollback()

	result, err := tx.Exec(
		`INSERT INTO files (filename, original_name, path, size, mime_type, uploaded_by, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		filename, originalName, savePath, size, mimeType, uploadedBy, time.Now(),
	)
	if err != nil {
		os.Remove(fullPath)
		return nil, err
	}
	fileID, _ := result.LastInsertId()

	s.audit.Log(tx, uploadedBy, "create", "file", ptrUint(uint(fileID)),
		map[string]interface{}{"original_name": originalName, "size": size}, ip)

	if err := tx.Commit(); err != nil {
		os.Remove(fullPath)
		return nil, err
	}

	return &models.File{
		ID:           uint(fileID),
		Filename:     filename,
		OriginalName: originalName,
		Path:         savePath,
		Size:         size,
		MimeType:     mimeType,
		UploadedBy:   uploadedBy,
		CreatedAt:    time.Now(),
	}, nil
}

func (s *FileService) List(page, pageSize int, keyword string) ([]models.File, int64, error) {
	var total int64
	countSQL := "SELECT COUNT(*) FROM files WHERE deleted_at IS NULL"
	args := []interface{}{}
	if keyword != "" {
		countSQL += " AND original_name LIKE ?"
		args = append(args, "%"+keyword+"%")
	}
	if err := s.db.Get(&total, countSQL, args...); err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	querySQL := "SELECT * FROM files WHERE deleted_at IS NULL"
	if keyword != "" {
		querySQL += " AND original_name LIKE ?"
	}
	querySQL += " ORDER BY created_at DESC LIMIT ? OFFSET ?"
	queryArgs := append(args, pageSize, offset)

	var files []models.File
	if err := s.db.Select(&files, querySQL, queryArgs...); err != nil {
		return nil, 0, err
	}
	if files == nil {
		files = []models.File{}
	}
	return files, total, nil
}

func (s *FileService) Delete(fileID, userID uint, ip string) error {
	tx, err := s.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec("DELETE FROM file_associations WHERE file_id = ?", fileID)
	if err != nil {
		return err
	}

	var f models.File
	if err := tx.Get(&f, "SELECT * FROM files WHERE id = ?", fileID); err != nil {
		return err
	}

	os.Remove(f.Path)

	_, err = tx.Exec("DELETE FROM files WHERE id = ?", fileID)
	if err != nil {
		return err
	}

	s.audit.Log(tx, userID, "delete", "file", ptrUint(fileID), f, ip)
	return tx.Commit()
}

func (s *FileService) CreateAssociation(req models.FileAssociationRequest, userID uint, ip string) (*models.FileAssociation, error) {
	tx, err := s.db.Beginx()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	result, err := tx.Exec(
		`INSERT INTO file_associations (file_id, target_type, target_id, created_at)
		 VALUES (?, ?, ?, ?)`,
		req.FileID, req.TargetType, req.TargetID, time.Now(),
	)
	if err != nil {
		return nil, err
	}
	assocID, _ := result.LastInsertId()

	s.audit.Log(tx, userID, "create", "file_association", ptrUint(uint(assocID)), req, ip)

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &models.FileAssociation{
		ID:         uint(assocID),
		FileID:     req.FileID,
		TargetType: req.TargetType,
		TargetID:   req.TargetID,
		CreatedAt:  time.Now(),
	}, nil
}

func (s *FileService) DeleteAssociation(assocID, userID uint, ip string) error {
	tx, err := s.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var assoc models.FileAssociation
	if err := tx.Get(&assoc, "SELECT * FROM file_associations WHERE id = ?", assocID); err != nil {
		return err
	}

	_, err = tx.Exec("DELETE FROM file_associations WHERE id = ?", assocID)
	if err != nil {
		return err
	}

	s.audit.Log(tx, userID, "delete", "file_association", ptrUint(assocID), assoc, ip)
	return tx.Commit()
}

func (s *FileService) GetAssociations(targetType string, targetID uint) ([]map[string]interface{}, error) {
	rows, err := s.db.Queryx(
		`SELECT fa.id, fa.file_id, fa.target_type, fa.target_id, fa.created_at,
		 f.original_name, f.filename, f.size, f.mime_type
		 FROM file_associations fa
		 JOIN files f ON f.id = fa.file_id
		 WHERE fa.target_type = ? AND fa.target_id = ? AND f.deleted_at IS NULL
		 ORDER BY fa.created_at DESC`,
		targetType, targetID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []map[string]interface{}
	for rows.Next() {
		row := make(map[string]interface{})
		if err := rows.MapScan(row); err != nil {
			return nil, err
		}
		result = append(result, row)
	}
	if result == nil {
		result = []map[string]interface{}{}
	}
	return result, nil
}
