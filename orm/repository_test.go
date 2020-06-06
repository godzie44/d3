package orm

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFindOneReturnErrIfNoSession(t *testing.T) {
	repo := Repository{}
	_, err := repo.FindOne(context.Background(), nil)
	assert.Equal(t, err, ErrSessionNotSet)
}

func TestFindAllReturnErrIfNoSession(t *testing.T) {
	repo := Repository{}
	_, err := repo.FindAll(context.Background(), nil)
	assert.Equal(t, err, ErrSessionNotSet)
}

func TestPersistsReturnErrIfNoSession(t *testing.T) {
	repo := Repository{}
	err := repo.Persists(context.Background(), nil)
	assert.Equal(t, err, ErrSessionNotSet)
}

func TestDeleteReturnErrIfNoSession(t *testing.T) {
	repo := Repository{}
	err := repo.Delete(context.Background(), nil)
	assert.Equal(t, err, ErrSessionNotSet)
}
