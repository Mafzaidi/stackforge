package memory

import (
	"context"
	"testing"
	"time"

	"github.com/mafzaidi/stackforge/internal/domain/entity"
)

func TestCredentialRepository_CreateAndList(t *testing.T) {
	repo := NewCredentialRepository()
	ctx := context.Background()

	// Test List on empty repository
	credentials, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(credentials) != 0 {
		t.Errorf("Expected empty list, got %d credentials", len(credentials))
	}

	// Test Create
	cred := &entity.Credential{
		ID:        "cred-1",
		Name:      "Test Credential",
		Username:  "testuser",
		Secret:    "secret123",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = repo.Create(ctx, cred)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Test List after Create
	credentials, err = repo.List(ctx)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(credentials) != 1 {
		t.Errorf("Expected 1 credential, got %d", len(credentials))
	}

	// Test GetByID
	retrieved, err := repo.GetByID(ctx, "cred-1")
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if retrieved.ID != cred.ID || retrieved.Name != cred.Name {
		t.Errorf("Retrieved credential doesn't match: got %+v, want %+v", retrieved, cred)
	}
}

func TestCredentialRepository_Update(t *testing.T) {
	repo := NewCredentialRepository()
	ctx := context.Background()

	cred := &entity.Credential{
		ID:        "cred-1",
		Name:      "Original Name",
		Username:  "testuser",
		Secret:    "secret123",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.Create(ctx, cred)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Update the credential
	cred.Name = "Updated Name"
	err = repo.Update(ctx, cred)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	// Verify update
	retrieved, err := repo.GetByID(ctx, "cred-1")
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if retrieved.Name != "Updated Name" {
		t.Errorf("Expected updated name, got %s", retrieved.Name)
	}
}

func TestCredentialRepository_Delete(t *testing.T) {
	repo := NewCredentialRepository()
	ctx := context.Background()

	cred := &entity.Credential{
		ID:        "cred-1",
		Name:      "Test Credential",
		Username:  "testuser",
		Secret:    "secret123",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.Create(ctx, cred)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Delete the credential
	err = repo.Delete(ctx, "cred-1")
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify deletion
	_, err = repo.GetByID(ctx, "cred-1")
	if err == nil {
		t.Error("Expected error when getting deleted credential, got nil")
	}
}
