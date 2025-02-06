package models

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Message struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	ChatID      primitive.ObjectID `bson:"chat_id"`
	SenderID    uint               `bson:"sender_id"`
	ReceiverID  uint               `bson:"receiver_id"`
	Content     string             `bson:"content"`
	MessageType string             `bson:"message_type"`
	Timestamp   int64              `bson:"timestamp"`
	CreatedAt   time.Time          `bson:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at"`
	IsDeleted   bool               `bson:"is_deleted"`
	IsRecalled  bool               `bson:"is_recalled"`
	IsRead      bool               `bson:"is_read"`
}

func (Message) CollectionName() string {
	return "messages"
}

func InsertMessage(ctx context.Context, db *mongo.Database, msg *Message) (*mongo.InsertOneResult, error) {
	msg.ID = primitive.NewObjectID()
	msg.Timestamp = time.Now().Unix()
	msg.CreatedAt = time.Now()
	msg.UpdatedAt = time.Now()

	collection := db.Collection(msg.CollectionName())
	return collection.InsertOne(ctx, msg)
}

func FindMessageByID(ctx context.Context, db *mongo.Database, id primitive.ObjectID) (*Message, error) {
	var msg Message
	collection := db.Collection(msg.CollectionName())

	err := collection.FindOne(ctx, bson.M{"_id": id}).Decode(&msg)
	if err != nil {
		return nil, err
	}
	return &msg, nil
}

func UpdateMessage(ctx context.Context, db *mongo.Database, id primitive.ObjectID, update bson.M) (*mongo.UpdateResult, error) {
	collection := db.Collection(Message{}.CollectionName())

	// Ensure we have something to update
	if len(update) == 0 {
		return nil, errors.New("no valid update fields provided")
	}

	update["updated_at"] = time.Now() // Always update timestamp
	return collection.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": update})
}

func DeleteMessage(ctx context.Context, db *mongo.Database, id primitive.ObjectID) (*mongo.UpdateResult, error) {
	return UpdateMessage(ctx, db, id, bson.M{"is_deleted": true})
}

func MarkMessageAsRead(ctx context.Context, db *mongo.Database, id primitive.ObjectID) (*mongo.UpdateResult, error) {
	return UpdateMessage(ctx, db, id, bson.M{"is_read": true})
}

func FindMessagesByUserID(ctx context.Context, db *mongo.Database, userID uint, page, limit int) ([]Message, error) {
	var messages []Message
	collection := db.Collection(Message{}.CollectionName())

	filter := bson.M{
		"$or": []bson.M{
			{"sender_id": userID},
			{"receiver_id": userID},
		},
	}

	findOptions := options.Find()
	findOptions.SetSort(bson.M{"timestamp": -1}) // Sort newest first
	findOptions.SetSkip(int64((page - 1) * limit))
	findOptions.SetLimit(int64(limit))

	cursor, err := collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var msg Message
		if err := cursor.Decode(&msg); err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}

	return messages, nil
}

func RecallMessage(ctx context.Context, db *mongo.Database, id primitive.ObjectID, userID uint) (*mongo.UpdateResult, error) {
	collection := db.Collection(Message{}.CollectionName())
	
	filter := bson.M{"_id": id, "sender_id": userID}

	update := bson.M{
		"$set": bson.M{
			"is_recalled": true,
			"updated_at":  time.Now(),
			"content":     "[Message Recalled]", // Optional: Replace content with placeholder
		},
	}

	return collection.UpdateOne(ctx, filter, update)
}
