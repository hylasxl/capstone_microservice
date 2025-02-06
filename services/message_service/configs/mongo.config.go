package configs

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var Client *mongo.Client

func ConnectMongoDB() {
	mongoURI := fmt.Sprintf("mongodb://%s:%s@%s:%s",
		os.Getenv("MONGO_USER"),
		os.Getenv("MONGO_PASSWORD"),
		os.Getenv("MONGO_HOST"),
		os.Getenv("MONGO_PORT"),
	)

	clientOptions := options.Client().ApplyURI(mongoURI)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	log.Println("Connected to MongoDB!")
	Client = client

	MigrateCollections(Client.Database(os.Getenv("MONGO_DB")))
}

func GetMongoCollection(collectionName string) *mongo.Collection {
	return Client.Database(os.Getenv("MONGO_DB")).Collection(collectionName)
}

func MigrateCollections(db *mongo.Database) {
	createChatIndexes(db)
	createMessageIndexes(db)
}

func createChatIndexes(db *mongo.Database) {
	collection := db.Collection("chats")
	_, err := collection.Indexes().CreateMany(context.Background(), []mongo.IndexModel{
		{Keys: bson.M{"participants": 1}},
		{Keys: bson.M{"last_message_at": -1}},
	})
	if err != nil {
		log.Fatalf("Failed to create index on participants: %v", err)
	}

	log.Println("Indexes for chats collection have been ensured")
}

func createMessageIndexes(db *mongo.Database) {
	collection := db.Collection("messages")
	_, err := collection.Indexes().CreateMany(context.Background(), []mongo.IndexModel{
		{Keys: bson.M{"chat_id": 1}},     // Index for faster retrieval in a chat
		{Keys: bson.M{"sender_id": 1}},   // Index for sender queries
		{Keys: bson.M{"receiver_id": 1}}, // Index for receiver queries
		{Keys: bson.M{"timestamp": -1}},  // Index for faster retrieval in a chat

	})
	if err != nil {
		log.Fatalf("Failed to create index on participants: %v", err)
	}
	log.Println("Indexes for messages collection have been ensured")
}
