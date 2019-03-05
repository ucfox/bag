package mongokit

import (
	"gopkg.in/mgo.v2/bson"
)

type BaseEntity struct {
	Id bson.ObjectId `json:"id"        bson:"_id,omitempty"`
}

type CommonList struct {
	Items interface{} `json:"items"`
	Total int         `json:"total"`
}
