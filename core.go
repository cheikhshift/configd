package main

import (
	"fmt"
	"net/http"
	"net/smtp"
	"os"
	"time"

	"gopkg.in/mgo.v2/bson"
)

const trialPeriod int64 = (84600) * 20

var StripeKey string = os.Getenv("STRIPE_SECRET")

type User struct {
	Email           string        `valid:"email,unique,required"`
	Id              bson.ObjectId `bson:"_id,omitempty"`
	Password, Owner string
	Joined          time.Time
	StripeID        string
}

func redirect(w http.ResponseWriter, req *http.Request) {
	// remove/add not default ports from req.Host
	target := fmt.Sprintf("https://configd.gophersauce.com%s", req.URL.Path)
	if len(req.URL.RawQuery) > 0 {
		target += fmt.Sprintf("?%s", req.URL.RawQuery)
	}

	http.Redirect(w, req, target, 301)
}

func Login(r *http.Request, u *User) error {

	err := ParseBody(r, u)
	if err != nil {
		return err
	}

	qry := dbs.Query(u, bson.M{"email": u.Email, "password": u.Password})

	err = qry.One(u)

	if err != nil {
		return err
	}

	return nil
}

func Join(r *http.Request, u *User) error {

	err := ParseBody(r, u)
	if err != nil {
		return err
	}
	u.Joined = time.Now()
	u.Owner = "sys"
	dbs.New(u)

	err = dbs.Save(u)
	if err != nil {
		return err
	}
	return nil
}

func ResetPassword(email string, password string) error {
	var u User
	_, err := dbs.UpdateAll(&u, bson.M{"email": email}, bson.M{"$set": bson.M{"password": password}})
	if err != nil {
		return err
	}

	//send new password via email

	return nil
}

func UpdatePassword(userid string, old_password, password string) error {
	var u User
	id := bson.ObjectIdHex(userid)
	err := dbs.Query(u, bson.M{"_id": id, "password": old_password}).One(&u)
	if err != nil {
		return err
	}
	_, err = dbs.UpdateAll(u, bson.M{"_id": id, "password": old_password}, bson.M{"$set": bson.M{"password": password}})
	if err != nil {
		return err
	}

	//send new password via email

	return nil
}

func HasSubscribed(o string) bool {
	var u User

	err := dbs.Query(&u, bson.M{"_id": bson.ObjectIdHex(o)}).One(&u)
	if err != nil {
		return false
	}

	return u.StripeID != ""
}

func SendEmail(to, subject, body string) error {
	from := os.Getenv("GMAIL_USERNAME")
	pass := os.Getenv("GMAIL_PASSWORD")

	msg := fmt.Sprintf("From: %s\nTo: %s\nSubject:%s\n\n%s\nDelivered with Love from Configd.", from, to, subject, body)

	err := smtp.SendMail("smtp.gmail.com:587",
		smtp.PlainAuth("", from, pass, "smtp.gmail.com"),
		from, []string{to}, []byte(msg))

	if err != nil {
		return err
	}

	return nil
}
