package mongodb

import (
	"context"

	"github.com/mafzaidi/stackforge/internal/domain/entity"
	"github.com/mafzaidi/stackforge/internal/domain/repository"
	"go.mongodb.org/mongo-driver/mongo"
)

// todoRepository implements repository.TodoRepository using MongoDB.
type todoRepository struct {
	client *mongo.Client
	dbName string
}

// NewTodoRepository creates a new MongoDB todo repository.
// Accepts a MongoDB client from ConnectionManager and database name.
// Requirements: 2.5
func NewTodoRepository(client *mongo.Client, dbName string) repository.TodoRepository {
	return &todoRepository{
		client: client,
		dbName: dbName,
	}
}

func (r *todoRepository) GetByID(ctx context.Context, id string) (*entity.Todo, error) {
	// TODO: implement
	return nil, nil
}

func (r *todoRepository) List(ctx context.Context) ([]*entity.Todo, error) {
	// TODO: implement
	return nil, nil
}

func (r *todoRepository) Create(ctx context.Context, t *entity.Todo) error {
	// TODO: implement
	return nil
}

func (r *todoRepository) Update(ctx context.Context, t *entity.Todo) error {
	// TODO: implement
	return nil
}

func (r *todoRepository) Delete(ctx context.Context, id string) error {
	// TODO: implement
	return nil
}
