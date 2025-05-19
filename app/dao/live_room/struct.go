package live_room

import "time"

// 直播间表
type LiveRoom struct {
	ID          int       `gorm:"column:id;primary_key" json:"id"`
	UniqueID    int64     `gorm:"column:unique_id" json:"unique_id"`
	LiveType    int       `gorm:"column:live_type" json:"live_type"`
	Name        string    `gorm:"column:name" json:"name"`
	Description string    `gorm:"column:description" json:"description"`
	Icon        string    `gorm:"column:icon" json:"icon"`
	Cover       string    `gorm:"column:cover" json:"cover"`
	Sort        int       `gorm:"column:sort" json:"sort"`
	IsDisabled  int       `gorm:"column:is_disabled" json:"is_disabled"`
	IsDefault   int       `gorm:"column:is_default" json:"is_default"`
	CreatedAt   time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (m *LiveRoom) TableName() string {
	return "live_room"
}
