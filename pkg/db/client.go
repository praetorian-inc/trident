package db

import (
	"fmt"
	"log"
	"net/url"
	"strings"

	// we are importing this so that we can use the cloudsql driver with gorm
	_ "github.com/GoogleCloudPlatform/cloudsql-proxy/proxy/dialers/postgres"
	"github.com/jinzhu/gorm"

	// also importing this one so we can use the postgres dialer
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

type ConnectionError struct {
	Msg string
}

func (ce *ConnectionError) Error() string {
	return fmt.Sprintf("connection error: %s", ce.Msg)
}

type TridentDB struct {
	db *gorm.DB
}

type Query struct {
	ReturnedFields []string
	Filter         map[string]interface{}
}

func New(connectionString string) (*TridentDB, error) {
	u, err := url.Parse(connectionString)
	if err != nil {
		log.Fatal(err)
	}

	driver := u.Scheme
	instance := u.Host
	user := u.User.Username()
	password, set := u.User.Password()
	if !set {
		return nil, &ConnectionError{Msg: "no password was provided to authenticate to the database."}
	}

	database := strings.Trim(u.Path, "/")

	parsedConnectionString := fmt.Sprintf("host=%s user=%s dbname=%s password=%s", instance, user, database, password)
	for k, v := range u.Query() {
		parsedConnectionString += fmt.Sprintf(" %s=%s", k, v[0])
	}

	var s TridentDB

	s.db, err = gorm.Open(driver, parsedConnectionString)
	if err != nil {
		msg := fmt.Sprintf("gorm encountered an error %s", err)
		return nil, &ConnectionError{Msg: msg}
	}

	s.db.AutoMigrate(&Campaign{})

	return &s, nil
}

func (t *TridentDB) InsertCampaign(campaign *Campaign) error {
	return t.db.Create(campaign).Error
}

func (t *TridentDB) SelectResults(query Query) ([]Result, error) {
	var results []Result

	err := t.db.Select(query.ReturnedFields).
		Where(query.Filter).
		Order("timestamp DESC").
		Find(&results).
		Error
	if err != nil {
		return nil, err
	}

	return results, nil
}
