package mongo

import (
	"context"
	"fmt"
	"github.com/Paincake/first-admin-lab/internal/config"
	"go.mongodb.org/mongo-driver/bson"
	"os"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	code, err := run(m)
	if err != nil {
		fmt.Println(err)
	}
	os.Exit(code)
}

func run(m *testing.M) (code int, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err = Connection(ctx)
	if err != nil {
		return -1, fmt.Errorf("error connecting to database:%w", err)
	}
	return m.Run(), nil
}

func TestMongoSchemaError(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := Connection(ctx)
	t.Cleanup(func() {
		err := client.Client.Disconnect(ctx)
		if err != nil {
			t.Fatalf("database connection closing failed:%s", err)
		}
	})
	cfg := config.MustLoad()
	_, err = client.Client.Database(cfg.DbName).Collection("hardware").InsertOne(ctx, bson.D{
		{"name", "test"},
		{"category", "test"},
		{"orders", bson.D{{}}},
	})
	if err == nil {
		t.Fatal("test failed")
	}
}
