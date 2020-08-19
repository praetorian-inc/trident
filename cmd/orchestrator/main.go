package main

import (
	"http"
	"time"
)

type SprayingTask struct {
	TargetURL         string
	CandidatePassword string
	UserList          []string
}

type SprayingCampaign struct {
	CreatedAt time.Time
	CreatedBy string
	TaskList  []SprayingTask
}

func main() {

}
