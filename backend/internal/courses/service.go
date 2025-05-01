package courses

import (
	"context"

	"github.com/shekshuev/codeed/backend/internal/logger"
)

// CourseService defines business operations related to courses.
//
// It abstracts logic for creating, retrieving, updating, deleting and filtering courses,
// and delegates persistence to the underlying Repository interface.
type CourseService interface {
	// CreateCourse creates a new course and returns the created DTO.
	// Returns an error if validation or persistence fails.
	CreateCourse(ctx context.Context, dto CreateCourseDTO) (*ReadCourseDTO, error)

	// GetCourseByID retrieves a course by its ObjectID string.
	// Returns ErrCourseNotFound or error if lookup fails.
	GetCourseByID(ctx context.Context, id string) (*ReadCourseDTO, error)

	// UpdateCourse updates an existing course identified by ID using the provided DTO.
	// Only non-nil fields are applied. Returns error if course is not found or update fails.
	UpdateCourse(ctx context.Context, id string, dto UpdateCourseDTO) error

	// DeleteCourse performs a soft-delete of the course by setting deleted_at.
	// Returns error if the course does not exist or deletion fails.
	DeleteCourse(ctx context.Context, id string) error

	// FindCourses retrieves courses matching filter criteria.
	// Supports filtering by title, tags, author, and publication status.
	FindCourses(ctx context.Context, filter FilterCoursesDTO) ([]*ReadCourseDTO, error)
}

type CourseServiceImpl struct {
	repo CourseRepository
	log  *logger.Logger
}

// NewService creates a new course service backed by a Repository and logger.
func NewService(repo CourseRepository) CourseService {
	return &CourseServiceImpl{
		repo: repo,
		log:  logger.NewLogger(),
	}
}

// CreateCourse creates a new course and logs the operation.
func (s *CourseServiceImpl) CreateCourse(ctx context.Context, dto CreateCourseDTO) (*ReadCourseDTO, error) {
	s.log.Sugar.Infof("creating course with title: %s", dto.Title)
	course, err := s.repo.Create(ctx, dto)
	if err != nil {
		s.log.Sugar.Errorf("failed to create course: %v", err)
		return nil, err
	}
	s.log.Sugar.Infof("course created: %s", course.ID)
	return course, nil
}

// GetCourseByID retrieves a course by ID and logs the operation.
func (s *CourseServiceImpl) GetCourseByID(ctx context.Context, id string) (*ReadCourseDTO, error) {
	s.log.Sugar.Infof("retrieving course: %s", id)
	course, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.log.Sugar.Warnf("failed to get course %s: %v", id, err)
		return nil, err
	}
	s.log.Sugar.Debugf("retrieved course: %+v", course)
	return course, nil
}

// UpdateCourse updates fields of an existing course and logs the result.
func (s *CourseServiceImpl) UpdateCourse(ctx context.Context, id string, dto UpdateCourseDTO) error {
	s.log.Sugar.Infof("updating course: %s", id)
	err := s.repo.UpdateByID(ctx, id, dto)
	if err != nil {
		s.log.Sugar.Warnf("failed to update course %s: %v", id, err)
		return err
	}
	s.log.Sugar.Infof("course updated: %s", id)
	return nil
}

// DeleteCourse performs a soft-delete and logs the operation.
func (s *CourseServiceImpl) DeleteCourse(ctx context.Context, id string) error {
	s.log.Sugar.Infof("deleting course: %s", id)
	err := s.repo.DeleteByID(ctx, id)
	if err != nil {
		s.log.Sugar.Warnf("failed to delete course %s: %v", id, err)
		return err
	}
	s.log.Sugar.Infof("course deleted: %s", id)
	return nil
}

// FindCourses applies filters and returns matching courses with log output.
func (s *CourseServiceImpl) FindCourses(ctx context.Context, filter FilterCoursesDTO) ([]*ReadCourseDTO, error) {
	s.log.Sugar.Infof("finding courses with filter: %+v", filter)
	courses, err := s.repo.Find(ctx, filter)
	if err != nil {
		s.log.Sugar.Errorf("failed to find courses: %v", err)
		return nil, err
	}
	s.log.Sugar.Infof("found %d course(s)", len(courses))
	return courses, nil
}
