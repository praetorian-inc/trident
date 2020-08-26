package db

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
	"github.com/lib/pq"
)

type Metadata map[string]string

func (a Metadata) Value() (driver.Value, error) {
	return json.Marshal(a)
}

func (a *Metadata) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(b, &a)
}

type Campaign struct {
	gorm.Model
	CreatedBy        string //take it or leave it really meh
	NotBefore        time.Time
	NotAfter         time.Time
	ScheduleInterval time.Duration
	Users            pq.StringArray `gorm:"type:varchar(255)[]"`
	Passwords        pq.StringArray `gorm:"type:varchar(255)[]"`
	Provider         string
	ProviderMetadata postgres.Jsonb
}

type Task struct {
	gorm.Model
	CampaignID     uint
	TargetUser     string
	TargetPassword string
	NotBefore      time.Time
	NotAfter       time.Time
	Result         bool
}
