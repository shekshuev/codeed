package courses

import (
	"context"
	"errors"
	"time"

	"github.com/shekshuev/codeed/backend/internal/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// MongoRepository is a MongoDB-backed implementation of the Repository interface for managing courses.
type CourseRepositoryImpl struct {
	collection *mongo.Collection
	log        *logger.Logger
}

// NewMongoRepository creates a new instance of MongoRepository that operates on the "courses" collection.
func NewCourseRepository(db *mongo.Database) *CourseRepositoryImpl {
	return &CourseRepositoryImpl{
		collection: db.Collection("courses"),
		log:        logger.NewLogger(),
	}
}

// Create inserts a new course into the MongoDB collection.
// It converts the DTO to a full course entity with generated ID and timestamps.
//
// Returns the created course as a ReadCourseDTO, or an error if insertion fails.
func (r *CourseRepositoryImpl) Create(ctx context.Context, dto CreateCourseDTO) (*ReadCourseDTO, error) {
	course, err := dto.ToCourseFromCreateDTO()
	if err != nil {
		r.log.Sugar.Warnf("failed to parse author ID: %v", err)
		return nil, err
	}

	_, err = r.collection.InsertOne(ctx, course)
	if err != nil {
		r.log.Sugar.Errorf("failed to insert course: %v", err)
		return nil, err
	}

	r.log.Sugar.Infof("created course: %s", course.ID.Hex())
	return course.ToReadDTO(), nil
}

// GetByID fetches a course by its MongoDB ObjectID string.
// The course must not be soft-deleted (i.e., deleted_at must not exist).
//
// Returns ErrCourseNotFound if no course is found or the ID is invalid.
func (r *CourseRepositoryImpl) GetByID(ctx context.Context, id string) (*ReadCourseDTO, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		r.log.Sugar.Warnf("invalid course ID format: %s", id)
		return nil, errors.New("invalid course id format")
	}

	filter := bson.M{"_id": objectID, "deleted_at": bson.M{"$exists": false}}
	var course Course
	err = r.collection.FindOne(ctx, filter).Decode(&course)
	if errors.Is(err, mongo.ErrNoDocuments) {
		r.log.Sugar.Infof("course not found: %s", id)
		return nil, ErrCourseNotFound
	} else if err != nil {
		r.log.Sugar.Errorf("failed to fetch course %s: %v", id, err)
		return nil, err
	}

	r.log.Sugar.Debugf("fetched course: %s", id)
	return course.ToReadDTO(), nil
}

// UpdateByID updates an existing course by its ID using fields provided in UpdateCourseDTO.
// Only non-nil fields are updated. Automatically updates the `updated_at` timestamp.
//
// Returns ErrCourseNotFound if the course does not exist or is soft-deleted.
func (r *CourseRepositoryImpl) UpdateByID(ctx context.Context, id string, dto UpdateCourseDTO) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		r.log.Sugar.Warnf("invalid course ID format: %s", id)
		return errors.New("invalid course id format")
	}

	update := dto.ToBsonUpdateFromUpdateDTO()
	if len(update) == 0 {
		r.log.Sugar.Infof("no updates provided for course: %s", id)
		return nil
	}

	res, err := r.collection.UpdateByID(ctx, objectID, update)
	if err != nil {
		r.log.Sugar.Errorf("failed to update course %s: %v", id, err)
		return err
	}
	if res.MatchedCount == 0 {
		r.log.Sugar.Infof("course not found for update: %s", id)
		return ErrCourseNotFound
	}

	r.log.Sugar.Infof("updated course: %s", id)
	return nil
}

// DeleteByID performs a soft-delete on the course with the specified ID
// by setting the `deleted_at` field to the current UTC time.
//
// Returns ErrCourseNotFound if the course does not exist.
func (r *CourseRepositoryImpl) DeleteByID(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		r.log.Sugar.Warnf("invalid course ID format: %s", id)
		return errors.New("invalid course id format")
	}

	update := bson.M{"$set": bson.M{"deleted_at": time.Now().UTC()}}
	res, err := r.collection.UpdateByID(ctx, objectID, update)
	if err != nil {
		r.log.Sugar.Errorf("failed to soft-delete course %s: %v", id, err)
		return err
	}
	if res.ModifiedCount == 0 {
		r.log.Sugar.Infof("course not found for deletion: %s", id)
		return ErrCourseNotFound
	}

	r.log.Sugar.Infof("soft-deleted course: %s", id)
	return nil
}

// Find returns a list of courses matching the given filter criteria.
// Only non-deleted courses are returned. Supports filtering by:
//   - partial title match (case-insensitive regex),
//   - tags (at least one matching),
//   - author ID,
//   - publication status.
//
// Returns an error if the filter is invalid or a database issue occurs.
func (r *CourseRepositoryImpl) Find(ctx context.Context, filterDTO FilterCoursesDTO) ([]*ReadCourseDTO, error) {
	filter := bson.M{"deleted_at": bson.M{"$exists": false}}

	if filterDTO.Title != nil && *filterDTO.Title != "" {
		filter["title"] = bson.M{"$regex": *filterDTO.Title, "$options": "i"}
	}

	if len(filterDTO.Tags) > 0 {
		filter["tags"] = bson.M{"$in": filterDTO.Tags}
	}

	if filterDTO.IsPublished != nil {
		filter["is_published"] = *filterDTO.IsPublished
	}

	if filterDTO.AuthorID != nil {
		authorID, err := primitive.ObjectIDFromHex(*filterDTO.AuthorID)
		if err != nil {
			r.log.Sugar.Warnf("invalid author ID in filter: %v", *filterDTO.AuthorID)
			return nil, errors.New("invalid author id format")
		}
		filter["author_id"] = authorID
	}

	r.log.Sugar.Infof("searching courses with filter: %+v", filter)

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		r.log.Sugar.Errorf("failed to execute find query: %v", err)
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []*ReadCourseDTO
	for cursor.Next(ctx) {
		var course Course
		if err := cursor.Decode(&course); err != nil {
			r.log.Sugar.Errorf("failed to decode course document: %v", err)
			return nil, err
		}
		results = append(results, course.ToReadDTO())
	}

	if err := cursor.Err(); err != nil {
		r.log.Sugar.Errorf("cursor error while iterating courses: %v", err)
		return nil, err
	}

	r.log.Sugar.Infof("found %d course(s)", len(results))
	return results, nil
}
