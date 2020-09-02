package db

import (
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/lib/pq"
)

type Datastore interface {
	InsertCampaign(*Campaign) error

	SelectResults(Query) ([]Result, error)
	InsertResult(*Result) error
}

type TridentDB struct {
	db *gorm.DB
}

type Query struct {
	ReturnedFields []string
	Filter         map[string]interface{}
}

type ConnectionError struct {
	Msg string
}

func (ce *ConnectionError) Error() string {
	return fmt.Sprintf("connection error: %s", ce.Msg)
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
	s.db.AutoMigrate(&Result{})

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

func (s *TridentDB) InsertResult(res *Result) error {
	return s.db.Create(res).Error
}

const (
	StreamingInsertTimeout = 3 * time.Second
	StreamingInsertMax     = 5000
)

func (s *TridentDB) StreamingInsertResults() chan *Result {
	results := make(chan *Result, StreamingInsertMax)
	go func() {
		for {
			txn, err := s.db.DB().Begin()
			if err != nil {
				log.Fatal(err)
			}

			stmt, err := txn.Prepare(pq.CopyIn("results",
				"campaign_id", "ip", "timestamp", "username", "password",
				"valid", "locked", "mfa", "rate_limited", "metadata",
			))
			if err != nil {
				log.Fatal(err)
			}

			execres := func(r *Result) {
				_, err = stmt.Exec(
					r.CampaignID, r.IP, r.Timestamp, r.Username, r.Password,
					r.Valid, r.Locked, r.MFA, r.RateLimited, r.Metadata,
				)
				if err != nil {
					log.Printf("error in streaming exec: %s", err)
					results <- r
				}
			}

			// 1st iter: block until we read a single result
			// Nth iter: attempt to Exec a result, but allow timeout
			//  At most, we will write StreamingInsertMax records at a time
			//  within StreamingInsertTimeout seconds
			execres(<-results)

			count := 1
			timer := time.NewTimer(StreamingInsertTimeout)
			for {
				select {
				case r := <-results:
					execres(r)

					count += 1
					if count > StreamingInsertMax {
						goto commit
					}
				case <-timer.C:
					goto commit
				}

				if !timer.Stop() {
					<-timer.C
				}
				timer.Reset(StreamingInsertTimeout)
			}

		commit:
			log.Printf("streaming %d records to db", count)

			_, err = stmt.Exec()
			if err != nil {
				log.Fatal(err)
			}

			err = stmt.Close()
			if err != nil {
				log.Fatal(err)
			}

			err = txn.Commit()
			if err != nil {
				log.Fatal(err)
			}
		}
	}()
	return results
}
