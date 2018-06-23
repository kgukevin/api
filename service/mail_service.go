package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/HackIllinois/api-commons/database"
	"github.com/HackIllinois/api-mail/config"
	"github.com/HackIllinois/api-mail/models"
	"gopkg.in/mgo.v2/bson"
	"net/http"
)

var db database.MongoDatabase

func init() {
	db_connection, err := database.InitMongoDatabase(config.MAIL_DB_HOST, config.MAIL_DB_NAME)

	if err != nil {
		panic(err)
	}

	db = db_connection
}

/*
	Send mail to the users in the given mailing list, using the provided template
	Substitution will be generated based on user info
*/
func SendMailByList(mail_order_list models.MailOrderList) (*models.MailStatus, error) {
	mail_list, err := GetMailList(mail_order_list.ListID)

	if err != nil {
		return nil, err
	}

	mail_order := models.MailOrder{
		IDs:      mail_list.UserIDs,
		Template: mail_order_list.Template,
	}

	return SendMailByID(mail_order)
}

/*
	Send mail the the users with the given ids, using the provided template
	Substitution will be generated based on user info
*/
func SendMailByID(mail_order models.MailOrder) (*models.MailStatus, error) {
	var mail_info models.MailInfo

	mail_info.Content = models.Content{
		TemplateID: mail_order.Template,
	}

	mail_info.Recipients = make([]models.Recipient, len(mail_order.IDs))
	for i, id := range mail_order.IDs {
		user_info, err := GetUserInfo(id)

		if err != nil {
			return nil, err
		}

		mail_info.Recipients[i].Address = models.Address{
			Email: user_info.Email,
			Name:  user_info.Username,
		}
		mail_info.Recipients[i].Substitutions = models.Substitutions{
			"name": user_info.Username,
		}
	}

	return SendMail(mail_info)
}

/*
	Send mail based on the given mailing info
	Returns the results of sending the mail
*/
func SendMail(mail_info models.MailInfo) (*models.MailStatus, error) {
	body := bytes.Buffer{}
	json.NewEncoder(&body).Encode(&mail_info)

	client := http.Client{}
	req, err := http.NewRequest("POST", config.SPARKPOST_API+"/transmissions/", &body)

	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", config.SPARKPOST_APIKEY)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("Failed to send mail")
	}

	var mail_status models.MailStatus
	json.NewDecoder(resp.Body).Decode(&mail_status)

	return &mail_status, nil
}

/*
	Create a mailing list with the given id and initial set of user, if provided
*/
func CreateMailList(mail_list models.MailList) error {
	if mail_list.UserIDs == nil {
		mail_list.UserIDs = []string{}
	}

	return db.Insert("lists", &mail_list)
}

/*
	Adds the given users to the specified mailing list
*/
func AddToMailList(mail_list models.MailList) error {
	selector := bson.M{
		"id": mail_list.ID,
	}

	modifier := bson.M{
		"$addToSet": bson.M{
			"userids": bson.M{
				"$each": mail_list.UserIDs,
			},
		},
	}

	return db.Update("lists", selector, &modifier)
}

/*
	Removes the given users from the specified mailing list
*/
func RemoveFromMailList(mail_list models.MailList) error {
	selector := bson.M{
		"id": mail_list.ID,
	}

	modifier := bson.M{
		"$pull": bson.M{
			"userids": bson.M{
				"$in": mail_list.UserIDs,
			},
		},
	}

	return db.Update("lists", selector, &modifier)
}

/*
	Gets the mail list with the given id
*/
func GetMailList(id string) (*models.MailList, error) {
	query := bson.M{
		"id": id,
	}

	var mail_list models.MailList
	err := db.FindOne("lists", query, &mail_list)

	if err != nil {
		return nil, err
	}

	return &mail_list, nil
}
