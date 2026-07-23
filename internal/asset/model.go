package asset

import "time"

type Asset struct {
	ID          int64     `json:"id" db:"id"`
	AssetType   string    `json:"assetType" db:"asset_type"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	Content     string    `json:"content" db:"content"`
	Status      int       `json:"status" db:"status"`
	Version     int       `json:"version" db:"version"`
	ParentID    int64     `json:"parentId" db:"parent_id"`
	IsCurrent   int       `json:"isCurrent" db:"is_current"`
	CreatedAt   time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt   time.Time `json:"updatedAt" db:"updated_at"`
}
