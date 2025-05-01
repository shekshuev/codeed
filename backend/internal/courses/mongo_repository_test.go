package courses

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func setupMongo(t *testing.T) (*CourseRepositoryImpl, func()) {
	ctx := context.Background()

	container, err := mongodb.Run(ctx, "mongo:6")
	assert.NoError(t, err)

	t.Cleanup(func() {
		_ = container.Terminate(ctx)
	})

	uri, err := container.ConnectionString(ctx)
	assert.NoError(t, err)

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	assert.NoError(t, err)

	db := client.Database("testdb")
	return NewMongoRepository(db), func() {
		_ = client.Disconnect(ctx)
	}
}

func TestMongoRepository_Create(t *testing.T) {
	repo, disconnect := setupMongo(t)
	defer disconnect()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	t.Run("successfully creates course", func(t *testing.T) {
		dto := CreateCourseDTO{
			Title:       "Advanced Go",
			Description: "Deep dive into concurrency and internals.",
			AuthorID:    "64e21f6a2f1c8e1c0a3d4f5b",
			Tags:        []string{"go", "concurrency"},
		}

		result, err := repo.Create(ctx, dto)
		assert.NoError(t, err)
		assert.NotNil(t, result)

		assert.Equal(t, dto.Title, result.Title)
		assert.Equal(t, dto.Description, result.Description)
		assert.Equal(t, dto.AuthorID, result.AuthorID)
		assert.Equal(t, dto.Tags, result.Tags)
		assert.False(t, result.IsPublished)
		assert.NotEmpty(t, result.ID)
		assert.NotEmpty(t, result.CreatedAt)
		assert.NotEmpty(t, result.UpdatedAt)
	})

	t.Run("returns error on invalid author ID", func(t *testing.T) {
		dto := CreateCourseDTO{
			Title:       "Broken",
			Description: "This should fail",
			AuthorID:    "not-a-valid-id",
			Tags:        []string{},
		}

		result, err := repo.Create(ctx, dto)
		assert.Nil(t, result)
		assert.Error(t, err)
	})
}

func TestMongoRepository_GetByID(t *testing.T) {
	repo, disconnect := setupMongo(t)
	defer disconnect()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	dto := CreateCourseDTO{
		Title:       "System Design",
		Description: "High-level architectures",
		AuthorID:    "64e21f6a2f1c8e1c0a3d4f5b",
		Tags:        []string{"architecture", "scalability"},
	}
	created, err := repo.Create(ctx, dto)
	assert.NoError(t, err)

	t.Run("returns course by ID", func(t *testing.T) {
		course, err := repo.GetByID(ctx, created.ID)
		assert.NoError(t, err)
		assert.NotNil(t, course)
		assert.Equal(t, created.ID, course.ID)
		assert.Equal(t, created.Title, course.Title)
	})

	t.Run("returns error for invalid ID format", func(t *testing.T) {
		course, err := repo.GetByID(ctx, "invalid-id")
		assert.Nil(t, course)
		assert.Error(t, err)
		assert.EqualError(t, err, "invalid course id format")
	})

	t.Run("returns error for non-existent course", func(t *testing.T) {
		fakeID := primitive.NewObjectID().Hex()
		course, err := repo.GetByID(ctx, fakeID)
		assert.Nil(t, course)
		assert.ErrorIs(t, err, ErrCourseNotFound)
	})

	t.Run("returns error for soft-deleted course", func(t *testing.T) {
		err := repo.DeleteByID(ctx, created.ID)
		assert.NoError(t, err)

		course, err := repo.GetByID(ctx, created.ID)
		assert.Nil(t, course)
		assert.ErrorIs(t, err, ErrCourseNotFound)
	})
}

func TestMongoRepository_UpdateByID(t *testing.T) {
	repo, disconnect := setupMongo(t)
	defer disconnect()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	dto := CreateCourseDTO{
		Title:       "Original Title",
		Description: "Initial desc",
		AuthorID:    "64e21f6a2f1c8e1c0a3d4f5b",
		Tags:        []string{"initial"},
	}

	created, err := repo.Create(ctx, dto)
	assert.NoError(t, err)

	t.Run("successfully updates title and tags", func(t *testing.T) {
		newTitle := "Updated Title"
		newTags := []string{"go", "backend"}
		update := UpdateCourseDTO{
			Title: &newTitle,
			Tags:  &newTags,
		}

		err := repo.UpdateByID(ctx, created.ID, update)
		assert.NoError(t, err)

		updated, err := repo.GetByID(ctx, created.ID)
		assert.NoError(t, err)
		assert.Equal(t, newTitle, updated.Title)
		assert.ElementsMatch(t, newTags, updated.Tags)
	})

	t.Run("does nothing when update DTO is empty", func(t *testing.T) {
		err := repo.UpdateByID(ctx, created.ID, UpdateCourseDTO{})
		assert.NoError(t, err)
	})

	t.Run("returns error if course not found", func(t *testing.T) {
		fakeID := primitive.NewObjectID().Hex()
		newTitle := "ShouldNotUpdate"
		err := repo.UpdateByID(ctx, fakeID, UpdateCourseDTO{Title: &newTitle})
		assert.ErrorIs(t, err, ErrCourseNotFound)
	})

	t.Run("returns error for invalid ID format", func(t *testing.T) {
		err := repo.UpdateByID(ctx, "invalid-id", UpdateCourseDTO{})
		assert.Error(t, err)
		assert.EqualError(t, err, "invalid course id format")
	})
}

func TestMongoRepository_DeleteByID(t *testing.T) {
	repo, disconnect := setupMongo(t)
	defer disconnect()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	dto := CreateCourseDTO{
		Title:       "To Be Deleted",
		Description: "This will be soft deleted",
		AuthorID:    "64e21f6a2f1c8e1c0a3d4f5b",
		Tags:        []string{"delete"},
	}

	created, err := repo.Create(ctx, dto)
	assert.NoError(t, err)

	t.Run("successfully soft deletes course", func(t *testing.T) {
		err := repo.DeleteByID(ctx, created.ID)
		assert.NoError(t, err)

		_, err = repo.GetByID(ctx, created.ID)
		assert.ErrorIs(t, err, ErrCourseNotFound)
	})

	t.Run("returns error if ID format is invalid", func(t *testing.T) {
		err := repo.DeleteByID(ctx, "invalid-object-id")
		assert.EqualError(t, err, "invalid course id format")
	})

	t.Run("returns error if course does not exist", func(t *testing.T) {
		fakeID := primitive.NewObjectID().Hex()
		err := repo.DeleteByID(ctx, fakeID)
		assert.ErrorIs(t, err, ErrCourseNotFound)
	})
}

func TestMongoRepository_Find(t *testing.T) {
	repo, disconnect := setupMongo(t)
	defer disconnect()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	authorID := "64e21f6a2f1c8e1c0a3d4f5b"

	courses := []CreateCourseDTO{
		{
			Title:       "Golang Basics",
			Description: "Intro to Go",
			AuthorID:    authorID,
			Tags:        []string{"go", "backend"},
		},
		{
			Title:       "Advanced Go",
			Description: "Deep dive into Go",
			AuthorID:    authorID,
			Tags:        []string{"go"},
		},
		{
			Title:       "Frontend 101",
			Description: "Intro to frontend",
			AuthorID:    authorID,
			Tags:        []string{"js", "frontend"},
		},
	}

	for _, dto := range courses {
		_, err := repo.Create(ctx, dto)
		assert.NoError(t, err)
	}

	t.Run("filters by title", func(t *testing.T) {
		query := "go"
		filter := FilterCoursesDTO{Title: &query}
		results, err := repo.Find(ctx, filter)
		assert.NoError(t, err)
		assert.Len(t, results, 2)
	})

	t.Run("filters by tags", func(t *testing.T) {
		filter := FilterCoursesDTO{Tags: []string{"frontend"}}
		results, err := repo.Find(ctx, filter)
		assert.NoError(t, err)
		assert.Len(t, results, 1)
		assert.Equal(t, "Frontend 101", results[0].Title)
	})

	t.Run("filters by author ID", func(t *testing.T) {
		filter := FilterCoursesDTO{AuthorID: &authorID}
		results, err := repo.Find(ctx, filter)
		assert.NoError(t, err)
		assert.Len(t, results, 3)
	})

	t.Run("filters with nonexistent author ID", func(t *testing.T) {
		id := primitive.NewObjectID().Hex()
		filter := FilterCoursesDTO{AuthorID: &id}
		results, err := repo.Find(ctx, filter)
		assert.NoError(t, err)
		assert.Len(t, results, 0)
	})

	t.Run("returns error on invalid author ID", func(t *testing.T) {
		id := "invalid-hex"
		filter := FilterCoursesDTO{AuthorID: &id}
		results, err := repo.Find(ctx, filter)
		assert.Nil(t, results)
		assert.EqualError(t, err, "invalid author id format")
	})
}
