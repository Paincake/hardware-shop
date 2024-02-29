package mongo

import (
	"github.com/Paincake/first-admin-lab/internal/database/repository"
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

func ConstructPipeline(filter repository.ProductFilter) mongo.Pipeline {
	var pipeline mongo.Pipeline
	if filter.Keyword != "" {
		pipeline = append(pipeline, bson.D{
			{"$match", bson.D{
				{"$text", bson.D{{"$search", filter.Keyword}}},
			}},
		})
	}
	if filter.Category != "" {
		pipeline = append(pipeline, bson.D{
			{"$match", bson.D{
				{"category", filter.Category},
			}},
		})
	}
	if filter.FloorCost != 0.0 {
		pipeline = append(pipeline, bson.D{
			{"$match", bson.D{
				{"cost", bson.D{{"$gte", filter.FloorCost}}},
			}},
		})
	}
	if filter.CeilingCost != 0.0 {
		pipeline = append(pipeline, bson.D{
			{"$match", bson.D{
				{"cost", bson.D{{"$lte", filter.CeilingCost}}},
			}},
		})
	}
	return pipeline

}
