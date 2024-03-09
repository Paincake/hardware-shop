package mongo

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func WithMatch(pipeline mongo.Pipeline, key string, value any) mongo.Pipeline {
	if key == "" || value == nil {
		return pipeline
	}
	return append(pipeline, bson.D{
		{"$match", bson.D{
			{key, value},
		}},
	})
}

func WithSearch(pipeline mongo.Pipeline, keyword string) mongo.Pipeline {
	if keyword == "" {
		return pipeline
	}
	return append(pipeline, bson.D{
		{"$match", bson.D{
			{"$text", bson.D{{"$search", keyword}}},
		}},
	})
}

func WithUnwind(pipeline mongo.Pipeline, field string) mongo.Pipeline {
	if field == "" {
		return pipeline
	}
	return append(pipeline, bson.D{{
		"$unwind", field,
	}})
}
