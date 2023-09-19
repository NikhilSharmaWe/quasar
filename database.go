package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Model interface {
	Id() string
}

type GenericRepo[T any] struct {
	Collection string
}

func (r *GenericRepo[T]) collection() (*mongo.Collection, error) {
	c, err := GetDBClient()
	if err != nil {
		return nil, err
	}

	collection := GetCollection(c, r.Collection)
	return collection, nil
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

func GetDBClient() (*mongo.Client, error) {

	err := godotenv.Load("vars.env")
	if err != nil {
		return nil, fmt.Errorf("failed to load .env file")
	}

	clientOptions := options.Client().ApplyURI(MONGO_URL)

	ctx, _ := context.WithTimeout(context.Background(), time.Second*10)

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func GetCollection(client *mongo.Client, name string) *mongo.Collection {
	collection := client.Database(DATABASE_NAME).Collection(name)
	return collection
}

func (r *GenericRepo[T]) SaveAccount(u *User) error {
	un := u.Username

	exists, err := r.IsExistsByField("username", un)
	if err != nil {
		return err
	}

	if exists {
		return errors.New(alreadyExistsErr)
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
