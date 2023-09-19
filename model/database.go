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

func (r *GenericRepo[T]) SaveAccount(u *User) error {
	un := u.Username

	exists, err := r.IsExistsByField("username", un)
	if err != nil {
		return err
	}

	if exists {
		return errors.New(AlreadyExistsErr)
	}

	doc, err := convertToBSON(u)
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

func (r *GenericRepo[T]) IsExistsByField(field string, val any) (bool, error) {
	collection, err := r.collection()
	if err != nil {
		return false, err
	}

	count, err := collection.CountDocuments(context.Background(), bson.M{field: val})
	if err != nil {
		return false, err
	}

	return count > 0, nil
}
