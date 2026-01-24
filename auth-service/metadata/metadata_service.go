package metadata

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	metadatadb "github.com/bpurdy1/auth-service/metadata/metadata-db"
)

var (
	ErrMetadataNotFound = errors.New("metadata not found")
	ErrInvalidInput     = errors.New("invalid input")
)

// UserMetadata is the public type for user metadata
type UserMetadata = metadatadb.UserMetadatum

// toPublic converts internal DB type to public type
func toPublic(m metadatadb.UserMetadatum) UserMetadata {
	return UserMetadata(m)
}

// toPublicSlice converts a slice of internal DB types to public types
func toPublicSlice(ms []metadatadb.UserMetadatum) []UserMetadata {
	result := make([]UserMetadata, len(ms))
	for i, m := range ms {
		result[i] = toPublic(m)
	}
	return result
}

// SetMetadataInput is the input for setting metadata
type SetMetadataInput = metadatadb.UpsertUserMetadataParams

type MetadataService interface {
	Set(ctx context.Context, input SetMetadataInput) (UserMetadata, error)
	Get(ctx context.Context, userID int64, key string) (UserMetadata, error)
	GetValue(ctx context.Context, userID int64, key string) (string, error)
	List(ctx context.Context, userID int64) ([]UserMetadata, error)
	AsMap(ctx context.Context, userID int64) (map[string]string, error)
	Delete(ctx context.Context, userID int64, key string) error
	DeleteAll(ctx context.Context, userID int64) error
	Count(ctx context.Context, userID int64) (int64, error)
}

type metadataServiceImpl struct {
	querier metadatadb.Querier
}

func NewMetadataService(db metadatadb.DBTX) MetadataService {
	return &metadataServiceImpl{
		querier: metadatadb.New(db),
	}
}

func (s *metadataServiceImpl) Set(ctx context.Context, input SetMetadataInput) (UserMetadata, error) {
	if input.Key == "" {
		return UserMetadata{}, ErrInvalidInput
	}

	meta, err := s.querier.UpsertUserMetadata(ctx, input)
	if err != nil {
		return UserMetadata{}, fmt.Errorf("failed to set metadata: %w", err)
	}

	return meta, nil
}

func (s *metadataServiceImpl) Get(ctx context.Context, userID int64, key string) (UserMetadata, error) {
	meta, err := s.querier.GetUserMetadata(ctx, metadatadb.GetUserMetadataParams{
		UserID: userID,
		Key:    key,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return UserMetadata{}, ErrMetadataNotFound
		}
		return UserMetadata{}, fmt.Errorf("failed to get metadata: %w", err)
	}

	return meta, nil
}

func (s *metadataServiceImpl) GetValue(ctx context.Context, userID int64, key string) (string, error) {
	meta, err := s.Get(ctx, userID, key)
	if err != nil {
		return "", err
	}
	return meta.Value, nil
}

func (s *metadataServiceImpl) List(ctx context.Context, userID int64) ([]UserMetadata, error) {
	metas, err := s.querier.ListUserMetadata(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list metadata: %w", err)
	}
	return metas, nil
}

func (s *metadataServiceImpl) AsMap(ctx context.Context, userID int64) (map[string]string, error) {
	metas, err := s.List(ctx, userID)
	if err != nil {
		return nil, err
	}

	result := make(map[string]string, len(metas))
	for _, m := range metas {
		result[m.Key] = m.Value
	}
	return result, nil
}

func (s *metadataServiceImpl) Delete(ctx context.Context, userID int64, key string) error {
	err := s.querier.DeleteUserMetadata(
		ctx,
		metadatadb.DeleteUserMetadataParams{
			UserID: userID,
			Key:    key,
		})
	if err != nil {
		return fmt.Errorf("failed to delete metadata: %w", err)
	}
	return nil
}

func (s *metadataServiceImpl) DeleteAll(ctx context.Context, userID int64) error {
	err := s.querier.DeleteAllUserMetadata(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to delete all metadata: %w", err)
	}
	return nil
}

func (s *metadataServiceImpl) Count(ctx context.Context, userID int64) (int64, error) {
	count, err := s.querier.CountUserMetadata(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("failed to count metadata: %w", err)
	}
	return count, nil
}
