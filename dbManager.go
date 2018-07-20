package main

import (
	"encoding/json"
	"net/http"

	"github.com/cheikhshift/gos/core"

	"gopkg.in/mgo.v2/bson"
)

type Configuration struct {
	Id            bson.ObjectId `bson:"_id,omitempty"`
	Config        map[string]interface{}
	Name          string
	Nodes         int
	ApiKey, Owner string
	Redeploy      bool
}

func AddConfiguration(r *http.Request, o string) error {
	var c Configuration
	err := ParseBody(r, &c)
	if err != nil {
		return err
	}

	dbs.New(&c)
	c.Name = "New Configuration"
	c.Config = make(map[string]interface{})
	c.ApiKey = core.NewLen(80)
	c.Owner = o
	err = dbs.Save(&c)
	if err != nil {
		return err
	}

	return nil
}

func UpdateConfiguration(r *http.Request, o string) error {

	var c Configuration
	err := ParseBody(r, &c)
	if err != nil {
		return err
	}
	c.Owner = o

	err = dbs.Update(&c)
	if err != nil {
		return err
	}

	return nil
}

func DeleteConfiguration(r *http.Request, o string) error {

	c := Configuration{}
	id := r.FormValue("id")

	c.Id = bson.ObjectIdHex(id)
	c.Owner = o

	err := dbs.Remove(&c)
	if err != nil {
		return err
	}

	return nil
}

func ListConfigurations(r *http.Request, o string, c interface{}) error {

	qry := dbs.Query(c, bson.M{"owner": o})

	err := qry.All(c)

	if err != nil {
		return err
	}

	return nil
}

func ParseBody(r *http.Request, i interface{}) error {
	decoder := json.NewDecoder(r.Body)

	err := decoder.Decode(i)
	if err != nil {
		return err
	}
	return nil
}

func HandleError(w http.ResponseWriter, err error, statusCode int) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	// The following variable message
	// will be encoded into a json string
	message := bson.M{"error": err.Error()}
	enc := json.NewEncoder(w)
	err = enc.Encode(message)
	return err
}
