package event

import "time"

type Event struct {
	ID         int64     `json:"id" db:"id"`
	EntityType string    `json:"entityType" db:"entity_type"`
	EntityID   int64     `json:"entityId" db:"entity_id"`
	EventType  string    `json:"eventType" db:"event_type"`
	Payload    string    `json:"payload" db:"payload"`
	OperatorID int64     `json:"operatorId" db:"operator_id"`
	CreatedAt  time.Time `json:"createdAt" db:"created_at"`
	Remark     string    `json:"remark" db:"remark"`
	IsDeleted  int       `json:"isDeleted" db:"is_deleted"`
}

type EventFilter struct {
	EntityType string `form:"entityType"`
	EntityID   int64  `form:"entityId"`
	EventType  string `form:"eventType"`
	OperatorID int64  `form:"operatorId"`
}
