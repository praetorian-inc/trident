package db

import (
	"time"

	"github.com/jinzhu/gorm"
)

type Campaign struct {
	gorm.Model
	CreatedBy string //take it or leave it really meh
	NotBefore time.Time
	NotAfter  time.Time
	TaskList  []Task
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
