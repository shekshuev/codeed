package courses

import (
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Course represents a learning course stored in MongoDB.
type Course struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`        // MongoDB ObjectID
	Title       string             `bson:"title"`                // Course title
	Description string             `bson:"description"`          // Brief description of the course
	AuthorID    primitive.ObjectID `bson:"author_id"`            // ID of the user who created the course
	Tags        []string           `bson:"tags,omitempty"`       // Optional tags for filtering/search
	IsPublished bool               `bson:"is_published"`         // Indicates if course is visible to users
	CreatedAt   time.Time          `bson:"created_at"`           // Creation timestamp (UTC)
	UpdatedAt   time.Time          `bson:"updated_at"`           // Last update timestamp (UTC)
	DeletedAt   *time.Time         `bson:"deleted_at,omitempty"` // Soft-delete timestamp (if deleted)
}

// CreateCourseDTO is used when creating a new course via API or service layer.
type CreateCourseDTO struct {
	Title       string   `json:"title"`       // Course title
	Description string   `json:"description"` // Course description
	AuthorID    string   `json:"author_id"`   // Author's ID (hex string)
	Tags        []string `json:"tags"`        // Optional tags
}

// UpdateCourseDTO defines updatable fields for a course.
// Only non-nil fields will be updated.
type UpdateCourseDTO struct {
	Title       *string   `json:"title,omitempty"`        // New title (optional)
	Description *string   `json:"description,omitempty"`  // New description (optional)
	Tags        *[]string `json:"tags,omitempty"`         // Updated tags (optional)
	IsPublished *bool     `json:"is_published,omitempty"` // Toggle publication status
}

// ReadCourseDTO is the output format returned from the service or API layer.
type ReadCourseDTO struct {
	ID          string   `json:"id"`           // ObjectID as hex string
	Title       string   `json:"title"`        // Title
	Description string   `json:"description"`  // Description
	AuthorID    string   `json:"author_id"`    // Author's ObjectID as hex
	Tags        []string `json:"tags"`         // Tags
	IsPublished bool     `json:"is_published"` // Publication status
	CreatedAt   string   `json:"created_at"`   // Creation time (RFC3339)
	UpdatedAt   string   `json:"updated_at"`   // Last updated (RFC3339)
}

// FilterCoursesDTO defines optional filters for listing courses.
type FilterCoursesDTO struct {
	Title       *string  // Optional full or partial title to match (case-insensitive)
	Tags        []string // Optional list of tags; course must have at least one
	IsPublished *bool    // Optional publication status filter
	AuthorID    *string  // Optional author ID (hex) to restrict search
}

// ToReadDTO converts a Course into a ReadCourseDTO for output.
func (c Course) ToReadDTO() *ReadCourseDTO {
	return &ReadCourseDTO{
		ID:          c.ID.Hex(),
		Title:       c.Title,
		Description: c.Description,
		AuthorID:    c.AuthorID.Hex(),
		Tags:        c.Tags,
		IsPublished: c.IsPublished,
		CreatedAt:   c.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   c.UpdatedAt.Format(time.RFC3339),
	}
}

// ToCourseFromCreateDTO converts a CreateCourseDTO into a Course struct with generated ID and timestamps.
func (dto CreateCourseDTO) ToCourseFromCreateDTO() (*Course, error) {
	authorID, err := primitive.ObjectIDFromHex(dto.AuthorID)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	return &Course{
		ID:          primitive.NewObjectID(),
		Title:       dto.Title,
		Description: dto.Description,
		AuthorID:    authorID,
		Tags:        dto.Tags,
		IsPublished: false,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

// ToBsonUpdateFromUpdateDTO maps UpdateCourseDTO to a bson.M for use in MongoDB update operations.
// Includes only non-nil fields and always updates `updated_at`.
func (dto UpdateCourseDTO) ToBsonUpdateFromUpdateDTO() bson.M {
	set := bson.M{}
	if dto.Title != nil {
		set["title"] = *dto.Title
	}
	if dto.Description != nil {
		set["description"] = *dto.Description
	}
	if dto.Tags != nil {
		set["tags"] = *dto.Tags
	}
	if dto.IsPublished != nil {
		set["is_published"] = *dto.IsPublished
	}
	if len(set) == 0 {
		return bson.M{}
	}
	set["updated_at"] = time.Now().UTC()
	return bson.M{"$set": set}
}
