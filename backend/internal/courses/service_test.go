package courses

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestCourseService_CreateCourse(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockRepository(ctrl)
	service := NewService(mockRepo)

	ctx := context.Background()
	createDTO := CreateCourseDTO{
		Title:       "Test",
		Description: "Desc",
		AuthorID:    "64e21f6a2f1c8e1c0a3d4f5b",
		Tags:        []string{"go"},
	}
	readDTO := &ReadCourseDTO{
		ID:          "abc123",
		Title:       "Test",
		Description: "Desc",
		AuthorID:    "64e21f6a2f1c8e1c0a3d4f5b",
		Tags:        []string{"go"},
		IsPublished: false,
		CreatedAt:   "2024-01-01T00:00:00Z",
		UpdatedAt:   "2024-01-01T00:00:00Z",
	}

	t.Run("success", func(t *testing.T) {
		mockRepo.EXPECT().Create(ctx, createDTO).Return(readDTO, nil).Times(1)

		result, err := service.CreateCourse(ctx, createDTO)
		assert.NoError(t, err)
		assert.Equal(t, readDTO, result)
	})

	t.Run("repo error", func(t *testing.T) {
		mockRepo.EXPECT().Create(ctx, createDTO).Return(nil, errors.New("db error")).Times(1)

		result, err := service.CreateCourse(ctx, createDTO)
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestCourseService_GetCourseByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockRepository(ctrl)
	service := NewService(mockRepo)

	ctx := context.Background()
	id := "abc123"
	readDTO := &ReadCourseDTO{ID: id, Title: "Test"}

	t.Run("success", func(t *testing.T) {
		mockRepo.EXPECT().GetByID(ctx, id).Return(readDTO, nil).Times(1)

		result, err := service.GetCourseByID(ctx, id)
		assert.NoError(t, err)
		assert.Equal(t, readDTO, result)
	})

	t.Run("not found", func(t *testing.T) {
		mockRepo.EXPECT().GetByID(ctx, id).Return(nil, ErrCourseNotFound).Times(1)

		result, err := service.GetCourseByID(ctx, id)
		assert.ErrorIs(t, err, ErrCourseNotFound)
		assert.Nil(t, result)
	})
}

func TestCourseService_UpdateCourse(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockRepository(ctrl)
	service := NewService(mockRepo)

	ctx := context.Background()
	id := "abc123"
	dto := UpdateCourseDTO{Title: ptrStr("New Title")}

	t.Run("success", func(t *testing.T) {
		mockRepo.EXPECT().UpdateByID(ctx, id, dto).Return(nil).Times(1)

		err := service.UpdateCourse(ctx, id, dto)
		assert.NoError(t, err)
	})

	t.Run("error", func(t *testing.T) {
		mockRepo.EXPECT().UpdateByID(ctx, id, dto).Return(errors.New("db fail")).Times(1)

		err := service.UpdateCourse(ctx, id, dto)
		assert.Error(t, err)
	})
}

func TestCourseService_DeleteCourse(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockRepository(ctrl)
	service := NewService(mockRepo)

	ctx := context.Background()
	id := "abc123"

	t.Run("success", func(t *testing.T) {
		mockRepo.EXPECT().DeleteByID(ctx, id).Return(nil).Times(1)

		err := service.DeleteCourse(ctx, id)
		assert.NoError(t, err)
	})

	t.Run("error", func(t *testing.T) {
		mockRepo.EXPECT().DeleteByID(ctx, id).Return(errors.New("db err")).Times(1)

		err := service.DeleteCourse(ctx, id)
		assert.Error(t, err)
	})
}

func TestCourseService_FindCourses(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockRepository(ctrl)
	service := NewService(mockRepo)

	ctx := context.Background()
	filter := FilterCoursesDTO{Tags: []string{"go"}}
	expected := []*ReadCourseDTO{
		{ID: "1", Title: "Course1"},
		{ID: "2", Title: "Course2"},
	}

	t.Run("success", func(t *testing.T) {
		mockRepo.EXPECT().Find(ctx, filter).Return(expected, nil).Times(1)

		result, err := service.FindCourses(ctx, filter)
		assert.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("error", func(t *testing.T) {
		mockRepo.EXPECT().Find(ctx, filter).Return(nil, errors.New("find failed")).Times(1)

		result, err := service.FindCourses(ctx, filter)
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func ptrStr(s string) *string {
	return &s
}
