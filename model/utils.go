package model

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func convertToBSON(m Model) (*bson.M, error) {
	b, err := bson.Marshal(m)
	if err != nil {
		return nil, err
	}

	doc := &bson.M{}
	err = bson.Unmarshal(b, doc)

	return doc, err
}

func (r *GenericRepo[T]) collection() (*mongo.Collection, error) {
	c, err := getDBClient()
	if err != nil {
		return nil, err
	}

	collection := getCollection(c, r.Collection)
	return collection, nil
}
