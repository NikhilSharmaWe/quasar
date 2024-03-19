package model

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	AlreadyExistsErr = "element already exists"
)

type Model interface {
	Id() string
}

type GenericRepo[T any] struct {
	Collection string
}

type User struct {
	ID        string    `bson:"_id,omitempty"`
	CreatedAt time.Time `bson:"created_at"`
	Username  string    `bson:"username"`
	Firstname string    `bson:"firstname"`
	Lastname  string    `bson:"lastname"`
	Password  []byte    `bson:"password"`
}

func (u *User) Id() string {
	return u.ID
}

type Meeting struct {
	ID         string    `bson:"_id,omitempty"`
	CreatedAt  time.Time `bson:"created_at"`
	Organizer  string    `bson:"organizer"`
	MeetingKey string    `bson:"meeting_key"`
}

func (m *Meeting) Id() string {
	return m.ID
}

type Chat struct {
	ID         string    `bson:"_id,omitempty"`
	CreatedAt  time.Time `bson:"created_at"`
	Username   string    `bson:"username"`
	MeetingKey string    `bson:"meeting_key"`
	Message    string    `bson:"message"`
}

func (c *Chat) Id() string {
	return c.ID
}

type Code struct {
	ID         string `bson:"_id,omitempty"`
	Code       string `bson:"code"`
	MeetingKey string `bson:"meeting_key"`
}

func (c *Code) Id() string {
	return c.ID
}

func (r *GenericRepo[T]) SaveAccount(u *User) error {

	filter := make(map[string]interface{})
	filter["username"] = u.Username

	exists, err := r.IsExists(filter)
	if err != nil {
		return err
	}

	if exists {
		return errors.New(AlreadyExistsErr)
	}

	return r.Save(u)
}

func (r *GenericRepo[T]) Save(model Model) error {
	doc, err := convertToBSON(model)
	if err != nil {
		return err
	}

	collection, err := r.collection()
	if err != nil {
		return err
	}

	_, err = collection.InsertOne(context.Background(), doc)
	return err
}

func (r *GenericRepo[T]) UpdateOne(filter interface{}, update interface{}) error {
	collection, err := r.collection()
	if err != nil {
		return err
	}

	// doc, err := convertToBSON(model)
	// if err != nil {
	// 	return err
	// }

	_, err = collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return err
	}

	return nil
}

func (r *GenericRepo[T]) Delete(filters bson.M) error {
	collection, err := r.collection()
	if err != nil {
		return err
	}

	_, err = collection.DeleteOne(context.Background(), filters)
	return err
}

func (r *GenericRepo[T]) FindOne(filters bson.M) (*T, error) {
	res := new(T)
	collection, err := r.collection()
	if err != nil {
		return res, err
	}

	doc := collection.FindOne(context.Background(), filters)
	err = doc.Decode(res)

	return res, err
}

func (r *GenericRepo[T]) Find(filters bson.M) ([]T, error) {
	res := []T{}
	collection, err := r.collection()
	if err != nil {
		return res, err
	}

	options := options.Find()

	cur, err := collection.Find(context.Background(), filters, options)
	if err != nil {
		return res, err
	}

	err = cur.All(context.Background(), &res)
	return res, err
}

func (r *GenericRepo[T]) IsExists(filter map[string]interface{}) (bool, error) {
	collection, err := r.collection()
	if err != nil {
		return false, err
	}

	count, err := collection.CountDocuments(context.Background(), filter)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}
