package articles

import (
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Article represents a single piece of course content (e.g. a lecture, chapter, or lesson).
// Stored in MongoDB, each article belongs to a course and may be versioned.
// Drafts can be hidden from the learner until published.
type Article struct {
	ID         primitive.ObjectID `bson:"_id,omitempty"`        // MongoDB ObjectID
	CourseID   primitive.ObjectID `bson:"course_id"`            // Associated course ID
	Title      string             `bson:"title"`                // Article title
	ContentMD  string             `bson:"content_md"`           // Markdown content
	ContentTxt string             `bson:"content_txt"`          // Plaintext for search indexing
	Order      int                `bson:"order"`                // Order within course structure
	Version    int                `bson:"version"`              // Article version (for history tracking)
	Tags       []string           `bson:"tags"`                 // Optional tags for filtering
	IsDraft    bool               `bson:"is_draft"`             // Whether the article is a draft
	CreatedAt  time.Time          `bson:"created_at"`           // UTC timestamp when created
	UpdatedAt  time.Time          `bson:"updated_at"`           // UTC timestamp when last updated
	DeletedAt  *time.Time         `bson:"deleted_at,omitempty"` // Timestamp of soft-deletion (nil if active)
}

// CreateArticleDTO is used to create a new article from the client/API side.
type CreateArticleDTO struct {
	CourseID   string   `json:"course_id"`   // Course ID (hex string)
	Title      string   `json:"title"`       // Article title
	ContentMD  string   `json:"content_md"`  // Markdown content
	ContentTxt string   `json:"content_txt"` // Plaintext content (for search)
	Order      int      `json:"order"`       // Display order within course
	Tags       []string `json:"tags"`        // Optional tags
}

// UpdateArticleDTO contains fields that can be updated in an article.
// Only non-nil fields will be applied during update.
type UpdateArticleDTO struct {
	Title      *string   `json:"title,omitempty"`       // New title (optional)
	ContentMD  *string   `json:"content_md,omitempty"`  // New Markdown content
	ContentTxt *string   `json:"content_txt,omitempty"` // New plaintext content
	Order      *int      `json:"order,omitempty"`       // New order
	Version    *int      `json:"version,omitempty"`     // New version number
	Tags       *[]string `json:"tags,omitempty"`        // Updated tags
	IsDraft    *bool     `json:"is_draft,omitempty"`    // Draft flag toggle
}

// ReadArticleDTO is the public-facing structure returned by services or APIs.
type ReadArticleDTO struct {
	ID         string   `json:"id"`          // Article ObjectID as hex
	CourseID   string   `json:"course_id"`   // Course ObjectID as hex
	Title      string   `json:"title"`       // Article title
	ContentMD  string   `json:"content_md"`  // Markdown content
	ContentTxt string   `json:"content_txt"` // Plaintext content
	Order      int      `json:"order"`       // Order within course
	Version    int      `json:"version"`     // Article version
	Tags       []string `json:"tags"`        // Tags
	IsDraft    bool     `json:"is_draft"`    // Whether article is a draft
	CreatedAt  string   `json:"created_at"`  // RFC3339 timestamp
	UpdatedAt  string   `json:"updated_at"`  // RFC3339 timestamp
}

// FilterArticlesDTO defines filter criteria for querying articles.
type FilterArticlesDTO struct {
	CourseID *string  // Filter by course ID (hex)
	Title    *string  // Optional partial title match (case-insensitive)
	Tags     []string // At least one of the specified tags
	IsDraft  *bool    // Filter by draft status
	Version  *int     // Exact version (optional)
}

// ToReadDTO maps an internal Article model to the public-facing ReadArticleDTO.
func (a Article) ToReadDTO() *ReadArticleDTO {
	return &ReadArticleDTO{
		ID:         a.ID.Hex(),
		CourseID:   a.CourseID.Hex(),
		Title:      a.Title,
		ContentMD:  a.ContentMD,
		ContentTxt: a.ContentTxt,
		Order:      a.Order,
		Version:    a.Version,
		Tags:       a.Tags,
		IsDraft:    a.IsDraft,
		CreatedAt:  a.CreatedAt.Format(time.RFC3339),
		UpdatedAt:  a.UpdatedAt.Format(time.RFC3339),
	}
}

// ToArticleFromCreateDTO converts incoming DTO into a new Article model.
// Automatically assigns a new ObjectID, timestamps, version=1, and marks as draft.
func (dto CreateArticleDTO) ToArticleFromCreateDTO() (*Article, error) {
	courseID, err := primitive.ObjectIDFromHex(dto.CourseID)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	return &Article{
		ID:         primitive.NewObjectID(),
		CourseID:   courseID,
		Title:      dto.Title,
		ContentMD:  dto.ContentMD,
		ContentTxt: dto.ContentTxt,
		Order:      dto.Order,
		Version:    1,
		IsDraft:    true,
		Tags:       dto.Tags,
		CreatedAt:  now,
		UpdatedAt:  now,
	}, nil
}

// ToBsonUpdateFromUpdateDTO builds a MongoDB-compatible update map from an UpdateArticleDTO.
// Only fields that are explicitly set (non-nil) will be included.
func (dto UpdateArticleDTO) ToBsonUpdateFromUpdateDTO() bson.M {
	set := bson.M{}
	if dto.Title != nil {
		set["title"] = *dto.Title
	}
	if dto.ContentMD != nil {
		set["content_md"] = *dto.ContentMD
	}
	if dto.ContentTxt != nil {
		set["content_txt"] = *dto.ContentTxt
	}
	if dto.Order != nil {
		set["order"] = *dto.Order
	}
	if dto.Version != nil {
		set["version"] = *dto.Version
	}
	if dto.Tags != nil {
		set["tags"] = *dto.Tags
	}
	if dto.IsDraft != nil {
		set["is_draft"] = *dto.IsDraft
	}
	if len(set) == 0 {
		return bson.M{}
	}
	set["updated_at"] = time.Now().UTC()
	return bson.M{"$set": set}
}
