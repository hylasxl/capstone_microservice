package models

import (
	"context"
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// Chat represents a conversation between users
type Chat struct {
	ID            primitive.ObjectID `bson:"_id,omitempty"`
	Participants  []uint32           `bson:"participants"` // List of user IDs in the chat
	LastMessage   string             `bson:"last_message"` // Last message content (optional)
	LastMessageAt time.Time          `bson:"last_message_at"`
	CreatedAt     time.Time          `bson:"created_at"`
	UpdatedAt     time.Time          `bson:"updated_at"`
}

// CollectionName returns the MongoDB collection name for chats
func (Chat) CollectionName() string {
	return "chats"
}

// CreateChat inserts a new chat into MongoDB
func CreateChat(ctx context.Context, db *mongo.Database, chat *Chat) (*mongo.InsertOneResult, error) {
	collection := db.Collection(chat.CollectionName())

	existingChat := Chat{}
	err := collection.FindOne(ctx, bson.M{"participants": bson.M{"$all": chat.Participants}}).Decode(&existingChat)
	if err == nil {
		return nil, errors.New("chat already exists")
	}

	chat.ID = primitive.NewObjectID()
	chat.LastMessage = ""
	chat.LastMessageAt = time.Now()
	chat.CreatedAt = time.Now()
	chat.UpdatedAt = time.Now()

	return collection.InsertOne(ctx, chat)
}

func FindChatByID(ctx context.Context, db *mongo.Database, id primitive.ObjectID) (*Chat, error) {
	var chat Chat
	collection := db.Collection(chat.CollectionName())

	err := collection.FindOne(ctx, bson.M{"_id": id}).Decode(&chat)
	if err != nil {
		return nil, err
	}
	return &chat, nil
}

// FindChatsByUser retrieves all chats for a given user
func FindChatsByUser(ctx context.Context, db *mongo.Database, userID uint) ([]Chat, error) {
	var chats []Chat
	collection := db.Collection(Chat{}.CollectionName())

	findOptions := options.Find()
	findOptions.SetSort(bson.M{"last_message_at": -1}) // Sort newest first

	cursor, err := collection.Find(ctx, bson.M{"participants": userID}, findOptions)
	if err != nil {
		return nil, err
	}
	defer func(cursor *mongo.Cursor, ctx context.Context) {
		err := cursor.Close(ctx)
		if err != nil {

		}
	}(cursor, ctx)

	for cursor.Next(ctx) {
		var chat Chat
		if err := cursor.Decode(&chat); err != nil {
			return nil, err
		}
		chats = append(chats, chat)
	}

	return chats, nil
}

func DeleteChatByID(ctx context.Context, db *mongo.Database, id string) error {
	chatId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	collection := db.Collection(Chat{}.CollectionName())
	_, err = collection.DeleteOne(ctx, bson.M{"_id": chatId})
	if err != nil {
		return err
	}
	return nil
}

// UpdateChat updates a chat (e.g., updating the last message)
func UpdateChat(ctx context.Context, db *mongo.Database, id primitive.ObjectID, update bson.M) (*mongo.UpdateResult, error) {
	collection := db.Collection(Chat{}.CollectionName())

	update["updated_at"] = time.Now() // Always update timestamp
	return collection.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": update})
}

func FindChatByParticipants(ctx context.Context, db *mongo.Database, userIDs []uint32) (*Chat, error) {
	collection := db.Collection(Chat{}.CollectionName())

	fmt.Printf("userIDs: %v\n", userIDs)
	filter := bson.M{
		"participants": bson.M{"$all": userIDs},
	}

	var chat Chat
	err := collection.FindOne(ctx, filter).Decode(&chat)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil // No matching chat found
		}
		return nil, err // Other DB error
	}

	return &chat, nil
}

func GetChatList(ctx context.Context, db *mongo.Database, req GetChatListRequest) ([]ChatList, error) {
	collection := db.Collection(Chat{}.CollectionName())

	// Pagination options
	skip := int64((req.Page - 1) * req.PageSize)
	limit := int64(req.PageSize)

	// Query to find chats where the user is a participant
	findOptions := options.Find().
		SetSort(bson.M{"last_message_at": -1}). // Sort by last message timestamp (newest first)
		SetSkip(skip).
		SetLimit(limit)

	cursor, err := collection.Find(ctx, bson.M{"participants": req.AccountID}, findOptions)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var chatLists []ChatList
	for cursor.Next(ctx) {
		var chat Chat
		if err := cursor.Decode(&chat); err != nil {
			return nil, err
		}

		// Determine the target account (the other participant)
		var targetAccountID uint64
		for _, participant := range chat.Participants {
			if uint64(participant) != req.AccountID {
				targetAccountID = uint64(participant)
				break
			}
		}

		// Fetch last message details
		var lastMessage Message
		msgCollection := db.Collection(Message{}.CollectionName())
		err := msgCollection.FindOne(ctx, bson.M{"chat_id": chat.ID}, options.FindOne().SetSort(bson.M{"timestamp": -1})).Decode(&lastMessage)
		if err != nil && err != mongo.ErrNoDocuments {
			return nil, err
		}

		// Count unread messages for this chat
		unreadCount, err := msgCollection.CountDocuments(ctx, bson.M{
			"chat_id":     chat.ID,
			"receiver_id": req.AccountID,
			"is_read":     false,
		})
		if err != nil {
			return nil, err
		}

		// Construct chat list entry
		chatLists = append(chatLists, ChatList{
			ChatID:                chat.ID.Hex(),
			AccountID:             req.AccountID,
			TargetAccountID:       targetAccountID,
			DisplayName:           "", // Fetch from user service if needed
			AvatarURL:             "", // Fetch from user service if needed
			LastMessageTimestamp:  lastMessage.Timestamp,
			LastMessageContent:    lastMessage.Content,
			UnreadMessageQuantity: uint64(unreadCount),
			Page:                  req.Page,
			PageSize:              req.PageSize,
			Participants:          chat.Participants, // Add participants list
		})
	}

	return chatLists, nil
}
