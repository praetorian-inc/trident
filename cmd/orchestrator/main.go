package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/golang/gddo/httputil/header"
	"github.com/kelseyhightower/envconfig"
	"github.com/mattn/go-colorable"
	log "github.com/sirupsen/logrus"
	"github.com/snowzach/rotatefilehook"
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
