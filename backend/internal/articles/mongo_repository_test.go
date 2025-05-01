package articles

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

func setupMongo(t *testing.T) (*ArticleRepositoryImpl, func()) {
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
	return NewMongoArticleRepository(db), func() {
		_ = client.Disconnect(ctx)
	}
}

func TestArticleRepository_Create(t *testing.T) {
	repo, disconnect := setupMongo(t)
	defer disconnect()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	t.Run("successfully creates article", func(t *testing.T) {
		dto := CreateArticleDTO{
			CourseID:   "64e21f6a2f1c8e1c0a3d4f5b",
			Title:      "Intro to Markdown",
			ContentMD:  "# Heading",
			ContentTxt: "Heading",
			Order:      1,
			Tags:       []string{"intro"},
		}
		result, err := repo.Create(ctx, dto)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, dto.Title, result.Title)
	})

	t.Run("returns error on invalid course ID", func(t *testing.T) {
		dto := CreateArticleDTO{
			CourseID:   "invalid-hex",
			Title:      "Bad",
			ContentMD:  "# Broken",
			ContentTxt: "Broken",
			Order:      1,
		}
		result, err := repo.Create(ctx, dto)
		assert.Nil(t, result)
		assert.Error(t, err)
	})
}

func TestArticleRepository_GetByID(t *testing.T) {
	repo, disconnect := setupMongo(t)
	defer disconnect()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	dto := CreateArticleDTO{
		CourseID:   "64e21f6a2f1c8e1c0a3d4f5b",
		Title:      "Read me",
		ContentMD:  "## Content",
		ContentTxt: "Content",
		Order:      1,
	}
	created, err := repo.Create(ctx, dto)
	assert.NoError(t, err)

	t.Run("successfully returns article by ID", func(t *testing.T) {
		result, err := repo.GetByID(ctx, created.ID)
		assert.NoError(t, err)
		assert.Equal(t, created.ID, result.ID)
	})

	t.Run("returns error on invalid ID", func(t *testing.T) {
		result, err := repo.GetByID(ctx, "not-object-id")
		assert.Nil(t, result)
		assert.Error(t, err)
	})

	t.Run("returns error if article not found", func(t *testing.T) {
		fake := primitive.NewObjectID().Hex()
		result, err := repo.GetByID(ctx, fake)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, ErrArticleNotFound)
	})
}

func TestArticleRepository_UpdateByID(t *testing.T) {
	repo, disconnect := setupMongo(t)
	defer disconnect()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	dto := CreateArticleDTO{
		CourseID:   "64e21f6a2f1c8e1c0a3d4f5b",
		Title:      "To Update",
		ContentMD:  "## Before",
		ContentTxt: "Before",
		Order:      1,
	}
	created, err := repo.Create(ctx, dto)
	assert.NoError(t, err)

	t.Run("successfully updates title", func(t *testing.T) {
		newTitle := "Updated"
		err := repo.UpdateByID(ctx, created.ID, UpdateArticleDTO{Title: &newTitle})
		assert.NoError(t, err)

		updated, err := repo.GetByID(ctx, created.ID)
		assert.NoError(t, err)
		assert.Equal(t, newTitle, updated.Title)
	})

	t.Run("returns error on invalid ID", func(t *testing.T) {
		err := repo.UpdateByID(ctx, "invalid-id", UpdateArticleDTO{})
		assert.Error(t, err)
	})

	t.Run("returns not found for nonexistent ID", func(t *testing.T) {
		id := primitive.NewObjectID().Hex()
		newTitle := "Doesn't Exist"
		err := repo.UpdateByID(ctx, id, UpdateArticleDTO{Title: &newTitle})
		assert.ErrorIs(t, err, ErrArticleNotFound)
	})
}

func TestArticleRepository_DeleteByID(t *testing.T) {
	repo, disconnect := setupMongo(t)
	defer disconnect()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	dto := CreateArticleDTO{
		CourseID:   "64e21f6a2f1c8e1c0a3d4f5b",
		Title:      "To Be Deleted",
		ContentMD:  "# Delete Me",
		ContentTxt: "Delete Me",
		Order:      1,
	}
	created, err := repo.Create(ctx, dto)
	assert.NoError(t, err)

	t.Run("successfully soft-deletes article", func(t *testing.T) {
		err := repo.DeleteByID(ctx, created.ID)
		assert.NoError(t, err)

		res, err := repo.GetByID(ctx, created.ID)
		assert.Nil(t, res)
		assert.ErrorIs(t, err, ErrArticleNotFound)
	})

	t.Run("returns error on invalid ID", func(t *testing.T) {
		err := repo.DeleteByID(ctx, "bad-id")
		assert.Error(t, err)
	})

	t.Run("returns error on non-existent article", func(t *testing.T) {
		fake := primitive.NewObjectID().Hex()
		err := repo.DeleteByID(ctx, fake)
		assert.ErrorIs(t, err, ErrArticleNotFound)
	})
}

func TestArticleRepository_Find(t *testing.T) {
	repo, disconnect := setupMongo(t)
	defer disconnect()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	courseID := "64e21f6a2f1c8e1c0a3d4f5b"
	articles := []CreateArticleDTO{
		{
			CourseID:   courseID,
			Title:      "Intro",
			ContentMD:  "# Go",
			ContentTxt: "Go",
			Order:      1,
			Tags:       []string{"go"},
		},
		{
			CourseID:   courseID,
			Title:      "Advanced",
			ContentMD:  "# Advanced",
			ContentTxt: "Advanced",
			Order:      2,
			Tags:       []string{"advanced"},
		},
	}

	for _, dto := range articles {
		_, err := repo.Create(ctx, dto)
		assert.NoError(t, err)
	}

	t.Run("filters by course ID", func(t *testing.T) {
		filter := FilterArticlesDTO{CourseID: &courseID}
		result, err := repo.Find(ctx, filter)
		assert.NoError(t, err)
		assert.Len(t, result, 2)
	})

	t.Run("filters by tags", func(t *testing.T) {
		filter := FilterArticlesDTO{Tags: []string{"advanced"}}
		result, err := repo.Find(ctx, filter)
		assert.NoError(t, err)
		assert.Len(t, result, 1)
	})

	t.Run("invalid course ID", func(t *testing.T) {
		id := "invalid"
		filter := FilterArticlesDTO{CourseID: &id}
		result, err := repo.Find(ctx, filter)
		assert.Nil(t, result)
		assert.Error(t, err)
	})
}

func TestArticleRepository_CloneWithIncrementedVersion(t *testing.T) {
	repo, disconnect := setupMongo(t)
	defer disconnect()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	t.Run("successfully clones and increments version", func(t *testing.T) {
		createDTO := CreateArticleDTO{
			CourseID:   "64e21f6a2f1c8e1c0a3d4f5b",
			Title:      "Intro to Markdown",
			ContentMD:  "# Markdown",
			ContentTxt: "Markdown",
			Order:      1,
			Tags:       []string{"markdown"},
		}

		original, err := repo.Create(ctx, createDTO)
		assert.NoError(t, err)

		cloned, err := repo.CloneWithIncrementedVersion(ctx, original.ID)
		assert.NoError(t, err)
		assert.NotEqual(t, original.ID, cloned.ID)
		assert.Equal(t, original.Title, cloned.Title)
		assert.Equal(t, original.CourseID, cloned.CourseID)
		assert.Equal(t, original.Version+1, cloned.Version)
	})

	t.Run("returns error for invalid ID format", func(t *testing.T) {
		cloned, err := repo.CloneWithIncrementedVersion(ctx, "invalid-object-id")
		assert.Nil(t, cloned)
		assert.Error(t, err)
		assert.EqualError(t, err, "invalid article id format")
	})

	t.Run("returns error if original not found", func(t *testing.T) {
		cloned, err := repo.CloneWithIncrementedVersion(ctx, "64e21f6a2f1c8e1c0a3d4f5c")
		assert.Nil(t, cloned)
		assert.ErrorIs(t, err, ErrArticleNotFound)
	})
}

func TestArticleRepository_FindAllVersions(t *testing.T) {
	repo, disconnect := setupMongo(t)
	defer disconnect()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	courseID := primitive.NewObjectID()
	title := "Intro to Markdown"

	for i := 1; i <= 3; i++ {
		article := Article{
			ID:         primitive.NewObjectID(),
			CourseID:   courseID,
			Title:      title,
			ContentMD:  "# Example " + string(rune(i)),
			ContentTxt: "Example " + string(rune(i)),
			Order:      1,
			Version:    i,
			Tags:       []string{"test"},
			IsDraft:    i < 3,
			CreatedAt:  time.Now().UTC(),
			UpdatedAt:  time.Now().UTC(),
		}
		_, err := repo.collection.InsertOne(ctx, article)
		assert.NoError(t, err)
	}

	t.Run("returns all versions for courseID + title", func(t *testing.T) {
		results, err := repo.FindAllVersions(ctx, courseID.Hex(), title)
		assert.NoError(t, err)
		assert.Len(t, results, 3)
	})

	t.Run("returns empty for nonexistent title", func(t *testing.T) {
		results, err := repo.FindAllVersions(ctx, courseID.Hex(), "Nope")
		assert.NoError(t, err)
		assert.Len(t, results, 0)
	})

	t.Run("returns error on invalid courseID", func(t *testing.T) {
		results, err := repo.FindAllVersions(ctx, "bad-hex", title)
		assert.Nil(t, results)
		assert.Error(t, err)
		assert.EqualError(t, err, "invalid course id format")
	})
}
