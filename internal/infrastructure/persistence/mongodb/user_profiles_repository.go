package mongodb

import (
	"context"
	"fmt"
	"time"

	"github.com/mafzaidi/stackforge/internal/domain/entity"
	"github.com/mafzaidi/stackforge/internal/domain/repository"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type userProfilesDocument struct {
	ID                 primitive.ObjectID `bson:"_id,omitempty"`
	UserID             string             `bson:"user_id"`
	EncryptionKeyHash  *string            `bson:"encryption_key_hash"`
	MasterPasswordHash *string            `bson:"master_password_hash"`
	PasswordSalt       *string            `bson:"password_salt"`
	CreatedAt          time.Time          `bson:"created_at"`
	UpdatedAt          time.Time          `bson:"updated_at"`
}

type userProfilesRepository struct {
	collection *mongo.Collection
}

func NewUserProfilesRepository(client *mongo.Client, dbName string) repository.UserProfilesRepository {
	return &userProfilesRepository{
		collection: client.Database(dbName).Collection("userProfiles"),
	}
}

func (r *userProfilesRepository) EnsureIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "user_id", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "created_at", Value: -1}},
		},
	}
	_, err := r.collection.Indexes().CreateMany(ctx, indexes)
	return err
}

func (r *userProfilesRepository) toEntity(doc *userProfilesDocument) *entity.UserProfiles {
	return &entity.UserProfiles{
		ID:                 doc.ID.Hex(),
		UserID:             doc.UserID,
		EncryptionKeyHash:  doc.EncryptionKeyHash,
		MasterPasswordHash: doc.MasterPasswordHash,
		PasswordSalt:       doc.PasswordSalt,
		CreatedAt:          doc.CreatedAt,
		UpdatedAt:          doc.UpdatedAt,
	}
}

func (r *userProfilesRepository) GetByID(ctx context.Context, id string) (*entity.UserProfiles, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var doc userProfilesDocument
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&doc)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return r.toEntity(&doc), nil
}

func (r *userProfilesRepository) GetByUserID(ctx context.Context, userID string) (*entity.UserProfiles, error) {
	var doc userProfilesDocument
	err := r.collection.FindOne(ctx, bson.M{"user_id": userID}).Decode(&doc)
	if err == mongo.ErrNoDocuments {
		return nil, fmt.Errorf("user profile not found for user: %s", userID)
	}
	if err != nil {
		return nil, err
	}
	return r.toEntity(&doc), nil
}

func (r *userProfilesRepository) List(ctx context.Context) ([]*entity.UserProfiles, error) {
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var profiles []*entity.UserProfiles
	for cursor.Next(ctx) {
		var doc userProfilesDocument
		if err := cursor.Decode(&doc); err != nil {
			return nil, err
		}
		profiles = append(profiles, r.toEntity(&doc))
	}
	return profiles, cursor.Err()
}

func (r *userProfilesRepository) Create(ctx context.Context, t *entity.UserProfiles) error {
	doc := userProfilesDocument{
		ID:                 primitive.NewObjectID(),
		UserID:             t.UserID,
		EncryptionKeyHash:  t.EncryptionKeyHash,
		MasterPasswordHash: t.MasterPasswordHash,
		PasswordSalt:       t.PasswordSalt,
		CreatedAt:          t.CreatedAt,
		UpdatedAt:          t.UpdatedAt,
	}
	result, err := r.collection.InsertOne(ctx, doc)
	if err != nil {
		return err
	}

	// Set the generated ID back to the entity
	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		t.ID = oid.Hex()
	}
	return nil
}

func (r *userProfilesRepository) Update(ctx context.Context, t *entity.UserProfiles) error {
	objectID, err := primitive.ObjectIDFromHex(t.ID)
	if err != nil {
		return err
	}

	update := bson.M{
		"$set": bson.M{
			"user_id":              t.UserID,
			"encryption_key_hash":  t.EncryptionKeyHash,
			"master_password_hash": t.MasterPasswordHash,
			"password_salt":        t.PasswordSalt,
			"updated_at":           t.UpdatedAt,
		},
	}

	result, err := r.collection.UpdateOne(ctx, bson.M{"_id": objectID}, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return fmt.Errorf("user profile not found: %s", t.ID)
	}
	return nil
}

func (r *userProfilesRepository) Delete(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": objectID})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return fmt.Errorf("user profile not found: %s", id)
	}
	return nil
}
