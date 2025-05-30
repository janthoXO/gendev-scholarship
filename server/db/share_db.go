package db

import (
	"context"
	"server/domain"
	"slices"

	"server/utils"

	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type QueryEntity struct {
	ShareId string       `bson:"_id,omitempty" json:"shareId"`
	Query   domain.Query `json:"query"`
}

var (
	sharedDb *mongo.Database
)

// InitShareDb connects to MongoDB and sets up the global client
func InitShareDb() {
	clientOpts := options.Client().ApplyURI(utils.Cfg.Database.Url)
	clientOpts.SetAuth(options.Credential{
		Username: utils.Cfg.Database.User,
		Password: utils.Cfg.Database.Password,
	})

	mongoClient, err := mongo.Connect(context.TODO(), clientOpts)
	if err != nil {
		log.WithError(err).Fatal("Failed to connect to MongoDB")
	}

	sharedDb = mongoClient.Database(utils.Cfg.Database.Name)
	log.Infof("Connected to MongoDB at %s", utils.Cfg.Database.Url)

	names, err := sharedDb.ListCollectionNames(context.TODO(), bson.D{{}})
	if err != nil {
		log.WithError(err).Fatal("Failed to list collections in MongoDB")
	}

	if !slices.Contains(names, "queries") {
		err = sharedDb.CreateCollection(context.TODO(), "queries")
		if err != nil {
			log.WithError(err).Fatal("Failed to create 'queries' collection in MongoDB")
		} else {
			log.Info("Created 'queries' collection in MongoDB")
		}
	}
}

func SaveQuery(context context.Context, queryEntity QueryEntity) (shareId string, err error) {
	collection := sharedDb.Collection("queries")

	// Insert the query entity into the collection
	result, err := collection.InsertOne(context, queryEntity)
	if err != nil {
		log.WithError(err).Error("Failed to save query")
		return shareId, err
	}

	shareId = result.InsertedID.(string)
	return shareId, nil
}

func GetQueryById(context context.Context, shareId string) (*domain.Query, error) {
	collection := sharedDb.Collection("queries")
	var queryEntity QueryEntity

	err := collection.FindOne(context, bson.M{"_id": shareId}).
		Decode(&queryEntity)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil // No document found
		}
		log.WithError(err).Error("Failed to find query by ID")
		return nil, err // Other errors
	}

	return &queryEntity.Query, nil
}

func QueryExists(context context.Context, shareId string) (bool, error) {
	collection := sharedDb.Collection("queries")
	count, err := collection.CountDocuments(context, bson.M{"_id": shareId})
	if err != nil {
		log.WithError(err).Error("Failed to check if query exists")
		return false, err
	}

	return count > 0, nil
}
