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
		ApplyURI(fmt.Sprintf("mongodb://%s:%s", cfg.Host, cfg.Port)).
		SetAuth(credentials)
	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		return nil, err
	}

	mongoClient = MongoClient{Client: client, DbName: cfg.DbName}
	log.Printf("connection: %s %s %s %s %s", cfg.Host, cfg.Port, cfg.DbUser, cfg.DbPassword, cfg.DbName)
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
	cursor, err := coll.Aggregate(ctx, ConstructPipeline(filter))
	if err != nil {
		return nil, err
	}

	var products []repository.Product
	err = cursor.All(ctx, &products)
	if err != nil {
		return nil, err
	}
	return products, nil
}

func (c *MongoClient) AddProductToCart(clientName string, clientPhone string, product repository.Product) ([]repository.Product, error) {
	coll := c.Client.Database(c.DbName).Collection("hardware")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	session, err := c.Client.StartSession()
	if err != nil {
		return nil, err
	}
	defer session.EndSession(ctx)
	transactionPipeline := mongo.Pipeline{}
	transactionPipeline = WithMatch(WithMatch(transactionPipeline, "manufacturer", product.Id.Manufacturer), "name", product.Id.Name)

	_, err = session.WithTransaction(ctx, func(ctx mongo.SessionContext) (interface{}, error) {
		_, err := coll.UpdateOne(ctx, bson.D{}, bson.D{{
			"$inc", bson.D{{
				"quantity", -1,
			}},
		}})

		if err != nil {
			return nil, err
		}
		order := bson.D{
			{"date", time.Now().UTC().Format(time.RFC3339)},
			{"status", "ORDERED"},
			{"client", bson.D{{"name", clientName}, {"phone", clientPhone}}},
		}
		_, err = coll.UpdateOne(ctx, bson.D{}, bson.D{{
			"$push", bson.D{{
				"orders", order},
			}},
		})
		return nil, err

	})

	if err != nil {
		return nil, err
	}
	err = session.CommitTransaction(ctx)
	if err != nil {
		return nil, err
	}
	log.Println("transaction completed")
	pipeline := mongo.Pipeline{}
	pipeline = WithMatch(WithMatch(WithUnwind(pipeline, "$orders"), "orders.status", "ORDERED"), "orders.client.name", clientName)
	pipeline = append(pipeline, bson.D{{
		"$group", bson.D{
			{"_id", "_$id"},
		},
	}})
	cursor, err := coll.Find(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	log.Println("client cart fetch completed")

	var products []repository.Product
	err = cursor.All(ctx, &products)
	if err != nil {
		return nil, err
	}
	return products, nil
}
