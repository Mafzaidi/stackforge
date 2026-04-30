package mongodb

import (
	"context"
	"math"
	"time"

	"github.com/mafzaidi/stackforge/internal/domain/entity"
	"github.com/mafzaidi/stackforge/internal/domain/repository"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type vaultDocument struct {
	ID          string    `bson:"id,omitempty"`
	UserID      string    `bson:"user_id"`
	Name        string    `bson:"name"`
	Description string    `bson:"description"`
	Icon        string    `bson:"icon"`
	IsDefault   bool      `bson:"is_default"`
	CreatedAt   time.Time `bson:"created_at"`
	UpdatedAt   time.Time `bson:"updated_at"`
}

type masterDataDocument struct {
	ID          string    `bson:"id,omitempty"`
	Module      string    `bson:"module"`
	Type        string    `bson:"type"`
	Name        string    `bson:"name"`
	Description string    `bson:"description"`
	Icon        string    `bson:"icon"`
	Color       string    `bson:"color"`
	SortOrder   string    `bson:"sort_order"`
	IsActive    bool      `bson:"is_active"`
	Metadata    bson.M    `bson:"metadata,omitempty"`
	CreatedAt   time.Time `bson:"created_at"`
	UpdatedAt   time.Time `bson:"updated_at"`
}

// credentialDocument represents the MongoDB document structure for a credential.
type credentialDocument struct {
	ID                primitive.ObjectID  `bson:"_id,omitempty"`
	CredemtialID      string              `bson:"credential_id,omitempty"`
	UserID            string              `bson:"user_id"`
	VaultID           string              `bson:"vault_id"`
	CategoryID        string              `bson:"category_id"`
	Title             string              `bson:"title"`
	SiteUrl           string              `bson:"site_url"`
	FaviconUrl        *string             `bson:"favicon_url,omitempty"`
	UsernameEncrypted string              `bson:"username_encrypted"`
	PasswordEncrypted string              `bson:"password_encrypted"`
	NotesEncrypted    *string             `bson:"notes_encrypted,omitempty"`
	IsFavorite        bool                `bson:"is_favorite"`
	PasswordStrength  int64               `bson:"password_strength"`
	LastUsedAt        *time.Time          `bson:"last_used_at,omitempty"`
	PasswordChangedAt *time.Time          `bson:"password_changed_at,omitempty"`
	ExpiresAt         *time.Time          `bson:"expires_at,omitempty"`
	AutoLogin         bool                `bson:"auto_login"`
	CreatedAt         time.Time           `bson:"created_at"`
	UpdatedAt         time.Time           `bson:"updated_at"`
	Vault             *vaultDocument      `bson:"vault,omitempty"`
	Category          *masterDataDocument `bson:"category,omitempty"`
	Tags              []*entity.Tag       `bson:"tags,omitempty"`
}

// credentialRepository implements repository.CredentialRepository using MongoDB.
type credentialRepository struct {
	collection *mongo.Collection
}

// NewCredentialRepository creates a new MongoDB credential repository.
// Accepts a MongoDB client from ConnectionManager and database name.
// Requirements: 2.5
func NewCredentialRepository(client *mongo.Client, dbName string) repository.CredentialRepository {
	return &credentialRepository{
		collection: client.Database(dbName).Collection("credentials"),
	}
}

// EnsureIndexes creates required indexes for the credentials collection.
func (r *credentialRepository) EnsureIndexes(ctx context.Context) error {
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

func (r *credentialRepository) toEntity(doc *credentialDocument) *entity.Credential {
	cred := &entity.Credential{
		ID:                doc.ID.Hex(),
		UserID:            doc.UserID,
		VaultID:           doc.VaultID,
		CategoryID:        doc.CategoryID,
		Title:             doc.Title,
		SiteUrl:           doc.SiteUrl,
		FaviconUrl:        doc.FaviconUrl,
		UsernameEncrypted: doc.UsernameEncrypted,
		PasswordEncrypted: doc.PasswordEncrypted,
		NotesEncrypted:    doc.NotesEncrypted,
		IsFavorite:        doc.IsFavorite,
		PasswordStrength:  doc.PasswordStrength,
		LastUsedAt:        doc.LastUsedAt,
		PasswordChangedAt: doc.PasswordChangedAt,
		ExpiresAt:         doc.ExpiresAt,
		AutoLogin:         doc.AutoLogin,
		CreatedAt:         doc.CreatedAt,
		UpdatedAt:         doc.UpdatedAt,
		Vault:             (*entity.Vault)(doc.Vault),
		Tags:              doc.Tags,
	}

	if doc.Category != nil {
		cred.Category = &entity.MasterData{
			ID:          doc.Category.ID,
			Module:      doc.Category.Module,
			Type:        doc.Category.Type,
			Name:        doc.Category.Name,
			Description: doc.Category.Description,
			Icon:        doc.Category.Icon,
			Color:       doc.Category.Color,
			SortOrder:   doc.Category.SortOrder,
			IsActive:    doc.Category.IsActive,
			Metadata:    map[string]interface{}(doc.Category.Metadata),
			CreatedAt:   doc.Category.CreatedAt,
			UpdatedAt:   doc.Category.UpdatedAt,
		}
	}

	return cred
}

func (r *credentialRepository) GetByID(ctx context.Context, id string) (*entity.Credential, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var doc credentialDocument
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&doc)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return r.toEntity(&doc), nil
}

func (r *credentialRepository) List(ctx context.Context, limit, offset int) ([]*entity.Credential, error) {
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}).SetSkip(int64(offset)).SetLimit(int64(limit))
	cursor, err := r.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var credentials []*entity.Credential
	for cursor.Next(ctx) {
		var doc credentialDocument
		if err := cursor.Decode(&doc); err != nil {
			return nil, err
		}
		credentials = append(credentials, r.toEntity(&doc))
	}
	return credentials, cursor.Err()
}

func (r *credentialRepository) ListByUserID(ctx context.Context, userID string, page, limit int) (*repository.PaginatedCredentials, error) {
	offset := int64((page - 1) * limit)

	filter := bson.M{"user_id": userID}
	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetSkip(offset).
		SetLimit(int64(limit))

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var items []*entity.Credential
	for cursor.Next(ctx) {
		var doc credentialDocument
		if err := cursor.Decode(&doc); err != nil {
			return nil, err
		}
		items = append(items, r.toEntity(&doc))
	}
	if err := cursor.Err(); err != nil {
		return nil, err
	}

	totalItems, err := r.CountByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	totalPages := int(math.Ceil(float64(totalItems) / float64(limit)))

	return &repository.PaginatedCredentials{
		Items:        items,
		TotalItems:   totalItems,
		TotalPages:   totalPages,
		CurrentPage:  page,
		ItemsPerPage: limit,
	}, nil
}

func (r *credentialRepository) CountByUserID(ctx context.Context, userID string) (int64, error) {
	count, err := r.collection.CountDocuments(ctx, bson.M{"user_id": userID})
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (r *credentialRepository) Create(ctx context.Context, c *entity.Credential) error {
	doc := credentialDocument{
		ID:                primitive.NewObjectID(),
		UserID:            c.UserID,
		Title:             c.Title,
		UsernameEncrypted: c.UsernameEncrypted,
		PasswordEncrypted: c.PasswordEncrypted,
		CreatedAt:         c.CreatedAt,
		UpdatedAt:         c.UpdatedAt,
	}

	result, err := r.collection.InsertOne(ctx, doc)
	if err != nil {
		return err
	}

	// Set the generated ID back to the entity
	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		c.ID = oid.Hex()
	}
	return nil
}

func (r *credentialRepository) Update(ctx context.Context, c *entity.Credential) error {
	objectID, err := primitive.ObjectIDFromHex(c.ID)
	if err != nil {
		return err
	}

	update := bson.M{
		"$set": bson.M{
			"name":       c.Title,
			"username":   c.UsernameEncrypted,
			"secret":     c.PasswordEncrypted,
			"updated_at": c.UpdatedAt,
		},
	}

	_, err = r.collection.UpdateOne(ctx, bson.M{"_id": objectID}, update)
	return err
}

func (r *credentialRepository) Delete(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	_, err = r.collection.DeleteOne(ctx, bson.M{"_id": objectID})
	return err
}
