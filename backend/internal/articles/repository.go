package articles

import (
	"context"
	"errors"
)

// ErrArticleNotFound is returned when an article is not found by ID or filters.
var ErrArticleNotFound = errors.New("article not found")

// Repository defines storage operations for managing articles.
//
// All operations apply only to non-deleted articles (i.e., where `deleted_at` is nil),
// unless explicitly noted (e.g., FindAllVersions).
type ArticleRepository interface {
	// Create inserts a new article (typically version 1 and marked as draft).
	Create(ctx context.Context, dto CreateArticleDTO) (*ReadArticleDTO, error)

	// GetByID retrieves an article by its ObjectID string.
	// Returns ErrArticleNotFound if the article does not exist or is soft-deleted.
	GetByID(ctx context.Context, id string) (*ReadArticleDTO, error)

	// UpdateByID performs a partial update on the article.
	// Only fields set in the UpdateArticleDTO are modified.
	// Returns ErrArticleNotFound if the article does not exist.
	UpdateByID(ctx context.Context, id string, dto UpdateArticleDTO) error

	// DeleteByID soft-deletes the article by setting `deleted_at`.
	DeleteByID(ctx context.Context, id string) error

	// CloneWithIncrementedVersion duplicates the article by ID,
	// increments its version, and sets draft = true.
	// Useful for creating a new version of existing content.
	CloneWithIncrementedVersion(ctx context.Context, id string) (*ReadArticleDTO, error)

	// Find retrieves articles matching filter criteria.
	// Typically used to show published content in a course.
	Find(ctx context.Context, filter FilterArticlesDTO) ([]*ReadArticleDTO, error)

	// FindAllVersions returns all versions (including drafts) of the same article chain.
	// Useful for editors and audit trails.
	FindAllVersions(ctx context.Context, courseID, title string) ([]*ReadArticleDTO, error)
}
