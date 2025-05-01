package courses

import (
	"context"
	"errors"
)

// ErrCourseNotFound is returned when a course with the specified ID does not exist.
var ErrCourseNotFound = errors.New("course not found")

// Repository defines the behavior of a course repository.
//
// This interface abstracts storage operations related to course entities.
type CourseRepository interface {
	// Create inserts a new course using data from the CreateCourseDTO.
	Create(ctx context.Context, dto CreateCourseDTO) (*ReadCourseDTO, error)

	// GetByID fetches a course by its MongoDB ObjectID string.
	// Returns ErrCourseNotFound if no matching document exists.
	GetByID(ctx context.Context, id string) (*ReadCourseDTO, error)

	// UpdateByID applies partial updates to a course with the given ID.
	// Only non-nil fields from the UpdateCourseDTO are updated.
	UpdateByID(ctx context.Context, id string, dto UpdateCourseDTO) error

	// DeleteByID performs a soft delete by setting the deleted_at timestamp.
	DeleteByID(ctx context.Context, id string) error

	// Find searches for courses matching the given filter criteria.
	// Results are limited to non-deleted courses.
	Find(ctx context.Context, filter FilterCoursesDTO) ([]*ReadCourseDTO, error)
}
