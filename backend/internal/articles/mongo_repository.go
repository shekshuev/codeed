package articles

import (
	"context"
	"errors"
	"time"

	"github.com/shekshuev/codeed/backend/internal/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ArticleRepositoryImpl is a MongoDB-based implementation of the ArticleRepository interface.
// It provides CRUD operations and filtering for articles stored in the "articles" collection.
type ArticleRepositoryImpl struct {
	collection *mongo.Collection
	log        *logger.Logger
}

// NewMongoArticleRepository returns a new instance of ArticleRepositoryImpl
// initialized with the provided MongoDB database and internal logger.
func NewMongoArticleRepository(db *mongo.Database) *ArticleRepositoryImpl {
	return &ArticleRepositoryImpl{
		collection: db.Collection("articles"),
		log:        logger.NewLogger(),
	}
}

// Create inserts a new article into the database.
// The article is converted from the CreateArticleDTO and defaults to version 1 and draft mode.
func (r *ArticleRepositoryImpl) Create(ctx context.Context, dto CreateArticleDTO) (*ReadArticleDTO, error) {
	article, err := dto.ToArticleFromCreateDTO()
	if err != nil {
		r.log.Sugar.Warnf("invalid course ID in article create: %v", err)
		return nil, err
	}

	_, err = r.collection.InsertOne(ctx, article)
	if err != nil {
		r.log.Sugar.Errorf("failed to insert article: %v", err)
		return nil, err
	}

	r.log.Sugar.Infof("created article: %s", article.ID.Hex())
	return article.ToReadDTO(), nil
}

// GetByID retrieves a single article by its Mongo ObjectID string.
// Returns ErrArticleNotFound if the document doesn't exist or is soft-deleted.
func (r *ArticleRepositoryImpl) GetByID(ctx context.Context, id string) (*ReadArticleDTO, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		r.log.Sugar.Warnf("invalid article id format: %s", id)
		return nil, errors.New("invalid article id format")
	}

	filter := bson.M{"_id": objectID, "deleted_at": bson.M{"$exists": false}}

	var a Article
	err = r.collection.FindOne(ctx, filter).Decode(&a)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			r.log.Sugar.Infof("article not found: %s", id)
			return nil, ErrArticleNotFound
		}
		r.log.Sugar.Errorf("failed to fetch article %s: %v", id, err)
		return nil, err
	}

	r.log.Sugar.Debugf("fetched article: %s", id)
	return a.ToReadDTO(), nil
}

// UpdateByID applies partial updates to an existing article identified by its ID.
// Only fields set in the UpdateArticleDTO are modified. Returns ErrArticleNotFound if no match is found.
func (r *ArticleRepositoryImpl) UpdateByID(ctx context.Context, id string, dto UpdateArticleDTO) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		r.log.Sugar.Warnf("invalid article id format: %s", id)
		return errors.New("invalid article id format")
	}

	update := dto.ToBsonUpdateFromUpdateDTO()
	if len(update) == 0 {
		r.log.Sugar.Infof("no updates provided for article: %s", id)
		return nil
	}

	res, err := r.collection.UpdateByID(ctx, objectID, update)
	if err != nil {
		r.log.Sugar.Errorf("failed to update article %s: %v", id, err)
		return err
	}
	if res.MatchedCount == 0 {
		r.log.Sugar.Infof("article not found for update: %s", id)
		return ErrArticleNotFound
	}

	r.log.Sugar.Infof("updated article: %s", id)
	return nil
}

// DeleteByID soft-deletes an article by setting the deleted_at timestamp.
// Returns ErrArticleNotFound if the document does not exist.
func (r *ArticleRepositoryImpl) DeleteByID(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		r.log.Sugar.Warnf("invalid article id format: %s", id)
		return errors.New("invalid article id format")
	}

	update := bson.M{"$set": bson.M{"deleted_at": time.Now().UTC()}}
	res, err := r.collection.UpdateByID(ctx, objectID, update)
	if err != nil {
		r.log.Sugar.Errorf("failed to delete article %s: %v", id, err)
		return err
	}
	if res.ModifiedCount == 0 {
		r.log.Sugar.Infof("article not found for deletion: %s", id)
		return ErrArticleNotFound
	}

	r.log.Sugar.Infof("soft-deleted article: %s", id)
	return nil
}

// Find queries articles using flexible filters: by course, draft status, and tags.
// Only non-deleted articles are returned. Tags use `$in` logic (any match).
func (r *ArticleRepositoryImpl) Find(ctx context.Context, filterDTO FilterArticlesDTO) ([]*ReadArticleDTO, error) {
	filter := bson.M{"deleted_at": bson.M{"$exists": false}}

	if filterDTO.CourseID != nil {
		cid, err := primitive.ObjectIDFromHex(*filterDTO.CourseID)
		if err != nil {
			r.log.Sugar.Warnf("invalid course id in filter: %v", *filterDTO.CourseID)
			return nil, errors.New("invalid course id format")
		}
		filter["course_id"] = cid
	}

	if filterDTO.IsDraft != nil {
		filter["is_draft"] = *filterDTO.IsDraft
	}

	if len(filterDTO.Tags) > 0 {
		filter["tags"] = bson.M{"$in": filterDTO.Tags}
	}

	r.log.Sugar.Infof("finding articles with filter: %+v", filter)

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		r.log.Sugar.Errorf("failed to execute find query: %v", err)
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []*ReadArticleDTO
	for cursor.Next(ctx) {
		var article Article
		if err := cursor.Decode(&article); err != nil {
			r.log.Sugar.Errorf("failed to decode article: %v", err)
			return nil, err
		}
		results = append(results, article.ToReadDTO())
	}

	if err := cursor.Err(); err != nil {
		r.log.Sugar.Errorf("cursor iteration error: %v", err)
		return nil, err
	}

	r.log.Sugar.Infof("found %d article(s)", len(results))
	return results, nil
}

func (r *ArticleRepositoryImpl) CloneWithIncrementedVersion(ctx context.Context, id string) (*ReadArticleDTO, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		r.log.Sugar.Warnf("invalid article ID format for clone: %s", id)
		return nil, errors.New("invalid article id format")
	}

	var original Article
	err = r.collection.FindOne(ctx, bson.M{
		"_id":        objectID,
		"deleted_at": bson.M{"$exists": false},
	}).Decode(&original)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrArticleNotFound
		}
		return nil, err
	}

	filter := bson.M{
		"course_id":  original.CourseID,
		"title":      original.Title,
		"deleted_at": bson.M{"$exists": false},
	}
	opts := options.FindOne().SetSort(bson.D{
		{Key: "version", Value: -1},
	})

	var latest Article
	err = r.collection.FindOne(ctx, filter, opts).Decode(&latest)
	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		return nil, err
	}

	now := time.Now().UTC()
	clone := original
	clone.ID = primitive.NewObjectID()
	clone.Version = max(original.Version, latest.Version) + 1
	clone.CreatedAt = now
	clone.UpdatedAt = now
	clone.DeletedAt = nil

	if _, err := r.collection.InsertOne(ctx, &clone); err != nil {
		r.log.Sugar.Errorf("failed to clone article: %v", err)
		return nil, err
	}

	r.log.Sugar.Infof("cloned article %s -> %s (v%d)", id, clone.ID.Hex(), clone.Version)

	return clone.ToReadDTO(), nil
}

func (r *ArticleRepositoryImpl) FindAllVersions(ctx context.Context, courseID, title string) ([]*ReadArticleDTO, error) {
	cid, err := primitive.ObjectIDFromHex(courseID)
	if err != nil {
		r.log.Sugar.Warnf("invalid course ID in FindAllVersions: %s", courseID)
		return nil, errors.New("invalid course id format")
	}

	filter := bson.M{
		"course_id":  cid,
		"title":      title,
		"deleted_at": bson.M{"$exists": false},
	}

	opts := options.Find().SetSort(bson.D{{Key: "version", Value: 1}})
	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		r.log.Sugar.Errorf("failed to fetch versions for article %q in course %s: %v", title, courseID, err)
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []*ReadArticleDTO
	for cursor.Next(ctx) {
		var article Article
		if err := cursor.Decode(&article); err != nil {
			r.log.Sugar.Errorf("failed to decode article version: %v", err)
			return nil, err
		}
		results = append(results, article.ToReadDTO())
	}

	if err := cursor.Err(); err != nil {
		r.log.Sugar.Errorf("cursor error while reading versions: %v", err)
		return nil, err
	}

	r.log.Sugar.Infof("found %d versions for article %q in course %s", len(results), title, courseID)
	return results, nil
}
