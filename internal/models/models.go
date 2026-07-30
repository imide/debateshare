package models

import (
	"time"

	"github.com/splode/fname"
	"gorm.io/gorm"
)

type Room struct {
	gorm.Model
	Code      string    `gorm:"uniqueIndex;not null" json:"code"`
	Name      string    `gorm:"size:120" json:"name"`
	ExpiresAt time.Time `gorm:"index;not null" json:"expiresAt"`

	Files       []File       `gorm:"foreignKey:RoomID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL" json:"-"`
	Subscribers []Subscriber `json:"-"`
}

type File struct {
	gorm.Model
	RoomID      uint   `gorm:"index;not null" json:"-"`
	UUID        string `gorm:"size:36;not null" json:"-"`
	Name        string `gorm:"size:255;not null" json:"name"`
	Size        int64  `gorm:"not null" json:"size"`
	ContentType string `gorm:"size:100" json:"contentType"`
}

// Subscriber
// Until a domain is made and a proper (and cheap!) way of sending emails is found and created, Subscriber will not be used. just here for reference
type Subscriber struct {
	gorm.Model
	RoomID uint   `gorm:"uniqueIndex:idx_room_email;not null"`
	Email  string `gorm:"uniqueIndex:idx_room_email;size:254;not null"`
}

func GenerateCode() (string, error) {
	return fname.NewGenerator(fname.WithDelimiter("_")).Generate()
}

func LiveRoom(db *gorm.DB, code string) (*Room, error) {
	var room Room
	err := db.Where("code = ? AND expires_at > ?", code, time.Now()).First(&room).Error
	if err != nil {
		return nil, err
	}
	return &room, nil
}
