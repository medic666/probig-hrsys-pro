package position

import (
	"time"

	"gorm.io/gorm"
	"probig/internal/pkg/database"
)

type PositionEvent = database.PositionEvent
type PositionSnapshot = database.PositionSnapshot
type Person = database.Person

var DB *gorm.DB

func SetDB(db *gorm.DB) {
	DB = db
}

type PositionSnapshotWithName struct {
	PositionSnapshot
	PersonName string `json:"person_name"`
}

type PositionEventWithName struct {
	PositionEvent
	PersonName string `json:"person_name"`
}

var FarFutureDate = time.Date(9999, 12, 31, 0, 0, 0, 0, time.UTC)
