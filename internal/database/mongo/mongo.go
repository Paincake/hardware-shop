package mongo

import (
	"context"
	"fmt"
	"github.com/Paincake/first-admin-lab/internal/config"
	"github.com/Paincake/first-admin-lab/internal/database/repository"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"time"
)

type MongoClient struct {
	Client *mongo.Client
	DbName string
}

var mongoClient MongoClient

func Connection(ctx context.Context) (*MongoClient, error) {
	if mongoClient.Client != nil {
		return &mongoClient, nil
	}

	cfg := config.MustLoad()
	credentials := options.Credential{
		AuthMechanism: "SCRAM-SHA-256",
		Username:      cfg.DbUser,
		Password:      cfg.DbPassword,
	}

	clientOpts := options.Client().
		ApplyURI(fmt.Sprintf("mongodb://%s:%s/", cfg.Host, cfg.Port)).
		SetAuth(credentials)
	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		return nil, err
	}
	mongoClient = MongoClient{Client: client, DbName: cfg.DbName}
	return &mongoClient, nil
}

func CloseConnection(ctx context.Context) {
	if err := mongoClient.Client.Disconnect(ctx); err != nil {
		log.Fatalf("error disconnecting from database:%s", err)
	}
}

func (c *MongoClient) GetProducts(filter repository.ProductFilter) ([]repository.Product, error) {
	coll := c.Client.Database(c.DbName).Collection("hardware")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	pipeline := mongo.Pipeline{}
	pipeline = WithSearch(pipeline, filter.Keyword)
	pipeline = WithMatch(pipeline, "category", filter.Category)
	pipeline = WithMatch(pipeline, "cost", bson.D{{"$gte", filter.FloorCost}})
	pipeline = WithMatch(pipeline, "cost", bson.D{{"$lte", filter.CeilingCost}})
	cursor, err := coll.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}

	var products []repository.Product
	err = cursor.All(ctx, &products)
	if err != nil {

	}
	return products, nil
}

func (c *MongoClient) AddProductToCart(clientName string, clientPhone string, product repository.Product) ([]repository.CartElement, error) {
	coll := c.Client.Database(c.DbName).Collection("hardware")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	order := bson.D{
		{"date", time.Now()},
		{"status", "ORDERED"},
		{"client", bson.D{{"name", clientName}, {"phone", clientPhone}}},
	}
	_, err := coll.UpdateOne(ctx, bson.D{{"name", product.Id.Name}}, bson.D{{
		"$push", bson.D{{
			"orders", order},
		}},
		{
			"$inc", bson.D{{
				"quantity", -1,
			}},
		}})
	if err != nil {
		return nil, fmt.Errorf("error on update cart query: %s", err)
	}
	pipeline := mongo.Pipeline{}
	pipeline = WithMatch(WithMatch(WithUnwind(pipeline, "$orders"), "orders.status", "ORDERED"), "orders.client.name", clientName)

	pipeline = append(pipeline, bson.D{{
		"$group", bson.D{
			{"_id", "$_id"},
			{"category", bson.D{{"$first", "$category"}}},
			{"cost", bson.D{{"$first", "$cost"}}},
		},
	}})

	cursor, err := coll.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("error on find client cart query: %s", err)
	}

	var products []repository.CartElement
	err = cursor.All(ctx, &products)
	if err != nil {
		return nil, err
	}
	log.Println("client cart fetch completed")
	return products, nil
}
