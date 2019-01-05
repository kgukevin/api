package service

import (
	"errors"
	"github.com/HackIllinois/api/common/database"
	"github.com/HackIllinois/api/services/rsvp/config"
	"github.com/HackIllinois/api/services/rsvp/models"
)

var db database.Database

func init() {
	db_connection, err := database.InitDatabase(config.RSVP_DB_HOST, config.RSVP_DB_NAME)

	if err != nil {
		panic(err)
	}

	db = db_connection
}

/*
	Returns the rsvp associated with the given user id
*/
func GetUserRsvp(id string) (*models.UserRsvp, error) {
	query := database.QuerySelector{
		"id": id,
	}

	var rsvp models.UserRsvp
	err := db.FindOne("rsvps", query, &rsvp)

	if err != nil {
		return nil, err
	}

	return &rsvp, nil
}

/*
	Creates the rsvp associated with the given user id
*/
func CreateUserRsvp(id string, rsvp models.UserRsvp) error {
	_, err := GetUserRsvp(id)

	if err != database.ErrNotFound {
		if err != nil {
			return err
		}
		return errors.New("Rsvp already exists")
	}

	err = db.Insert("rsvps", &rsvp)

	return err
}

/*
	Updates the rsvp associated with the given user id
*/
func UpdateUserRsvp(id string, rsvp models.UserRsvp) error {
	selector := database.QuerySelector{
		"id": id,
	}

	err := db.Update("rsvps", selector, &rsvp)

	return err
}

/*
	Returns all rsvp stats
*/
func GetStats() (map[string]interface{}, error) {
	return db.GetStats("rsvps", []string{"isattending"})
}
