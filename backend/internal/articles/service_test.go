package articles

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestArticleService_CreateArticle(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockArticleRepository(ctrl)
	service := NewArticleService(mockRepo)

	dto := CreateArticleDTO{
		CourseID:   "64e21f6a2f1c8e1c0a3d4f5b",
		Title:      "Test Article",
		ContentMD:  "Some *markdown* content",
		ContentTxt: "Some markdown content",
		Order:      1,
		Tags:       []string{"go"},
	}
	expected := &ReadArticleDTO{ID: "abc123", Title: dto.Title, CourseID: dto.CourseID}

	t.Run("success", func(t *testing.T) {
		mockRepo.EXPECT().Create(gomock.Any(), dto).Return(expected, nil)

		result, err := service.CreateArticle(context.Background(), dto)
		assert.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("fail", func(t *testing.T) {
		mockRepo.EXPECT().Create(gomock.Any(), dto).Return(nil, errors.New("insert error"))

		result, err := service.CreateArticle(context.Background(), dto)
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestArticleService_GetArticleByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockArticleRepository(ctrl)
	service := NewArticleService(mockRepo)

	id := "abc123"
	expected := &ReadArticleDTO{ID: id}

	t.Run("success", func(t *testing.T) {
		mockRepo.EXPECT().GetByID(gomock.Any(), id).Return(expected, nil)

		result, err := service.GetArticleByID(context.Background(), id)
		assert.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("not found", func(t *testing.T) {
		mockRepo.EXPECT().GetByID(gomock.Any(), id).Return(nil, ErrArticleNotFound)

		result, err := service.GetArticleByID(context.Background(), id)
		assert.ErrorIs(t, err, ErrArticleNotFound)
		assert.Nil(t, result)
	})
}

func TestArticleService_UpdateArticle(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockArticleRepository(ctrl)
	service := NewArticleService(mockRepo)

	id := "abc123"
	title := "Updated"
	dto := UpdateArticleDTO{
		Title: &title,
	}

	t.Run("success", func(t *testing.T) {
		mockRepo.EXPECT().UpdateByID(gomock.Any(), id, dto).Return(nil)

		err := service.UpdateArticle(context.Background(), id, dto)
		assert.NoError(t, err)
	})

	t.Run("error", func(t *testing.T) {
		mockRepo.EXPECT().UpdateByID(gomock.Any(), id, dto).Return(errors.New("update error"))

		err := service.UpdateArticle(context.Background(), id, dto)
		assert.Error(t, err)
	})
}

func TestArticleService_DeleteArticle(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockArticleRepository(ctrl)
	service := NewArticleService(mockRepo)

	id := "abc123"

	t.Run("success", func(t *testing.T) {
		mockRepo.EXPECT().DeleteByID(gomock.Any(), id).Return(nil)

		err := service.DeleteArticle(context.Background(), id)
		assert.NoError(t, err)
	})

	t.Run("not found", func(t *testing.T) {
		mockRepo.EXPECT().DeleteByID(gomock.Any(), id).Return(ErrArticleNotFound)

		err := service.DeleteArticle(context.Background(), id)
		assert.ErrorIs(t, err, ErrArticleNotFound)
	})
}

func TestArticleService_FindArticles(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockArticleRepository(ctrl)
	service := NewArticleService(mockRepo)

	filter := FilterArticlesDTO{}
	expected := []*ReadArticleDTO{
		{ID: "a1", Title: "A"},
		{ID: "a2", Title: "B"},
	}

	t.Run("success", func(t *testing.T) {
		mockRepo.EXPECT().Find(gomock.Any(), filter).Return(expected, nil)

		result, err := service.FindArticles(context.Background(), filter)
		assert.NoError(t, err)
		assert.Len(t, result, 2)
	})

	t.Run("error", func(t *testing.T) {
		mockRepo.EXPECT().Find(gomock.Any(), filter).Return(nil, errors.New("db error"))

		result, err := service.FindArticles(context.Background(), filter)
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestArticleService_UpdateWithVersioning(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockArticleRepository(ctrl)
	service := NewArticleService(mockRepo)

	ctx := context.Background()
	originalID := "abc123"
	clonedID := "def456"

	cloned := &ReadArticleDTO{
		ID:      clonedID,
		Title:   "Old Title",
		Version: 2,
	}

	updated := &ReadArticleDTO{
		ID:      clonedID,
		Title:   "New Title",
		Version: 2,
	}
	title := "New Title"
	dto := UpdateArticleDTO{Title: &title}

	t.Run("successfully clones and updates article", func(t *testing.T) {
		mockRepo.EXPECT().CloneWithIncrementedVersion(ctx, originalID).Return(cloned, nil)
		mockRepo.EXPECT().UpdateByID(ctx, clonedID, dto).Return(nil)
		mockRepo.EXPECT().GetByID(ctx, clonedID).Return(updated, nil)

		result, err := service.(*ArticleServiceImpl).UpdateWithVersioning(ctx, originalID, dto)
		assert.NoError(t, err)
		assert.Equal(t, updated.ID, result.ID)
		assert.Equal(t, "New Title", result.Title)
	})

	t.Run("fails to clone", func(t *testing.T) {
		mockRepo.EXPECT().CloneWithIncrementedVersion(ctx, originalID).Return(nil, errors.New("clone error"))

		result, err := service.(*ArticleServiceImpl).UpdateWithVersioning(ctx, originalID, dto)
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("fails to update clone", func(t *testing.T) {
		mockRepo.EXPECT().CloneWithIncrementedVersion(ctx, originalID).Return(cloned, nil)
		mockRepo.EXPECT().UpdateByID(ctx, clonedID, dto).Return(errors.New("update error"))

		result, err := service.(*ArticleServiceImpl).UpdateWithVersioning(ctx, originalID, dto)
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}
