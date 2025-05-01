package articles

import (
	"context"

	"github.com/shekshuev/codeed/backend/internal/logger"
)

// ArticleService defines business operations related to articles.
type ArticleService interface {
	// CreateArticle adds a new article and returns the created result.
	// Returns an error if validation or DB insert fails.
	CreateArticle(ctx context.Context, dto CreateArticleDTO) (*ReadArticleDTO, error)

	// GetArticleByID retrieves a single article by ID.
	// Returns ErrArticleNotFound or error on lookup failure.
	GetArticleByID(ctx context.Context, id string) (*ReadArticleDTO, error)

	// UpdateArticle modifies an article’s fields.
	// Only non-nil fields are applied. Returns error if the article is not found.
	UpdateArticle(ctx context.Context, id string, dto UpdateArticleDTO) error

	// DeleteArticle marks an article as deleted (soft-delete).
	// Returns error if the article does not exist or deletion fails.
	DeleteArticle(ctx context.Context, id string) error

	// FindArticles lists articles by optional filters (course, draft, tags).
	// Returns slice of matched articles or empty slice if none found.
	FindArticles(ctx context.Context, filter FilterArticlesDTO) ([]*ReadArticleDTO, error)

	// UpdateWithVersioning clones article with incremented version, applies updates to the clone,
	// and returns the updated copy.
	UpdateWithVersioning(ctx context.Context, id string, dto UpdateArticleDTO) (*ReadArticleDTO, error)
}

// ArticleServiceImpl is the concrete implementation of ArticleService.
type ArticleServiceImpl struct {
	repo ArticleRepository
	log  *logger.Logger
}

// NewArticleService creates a new ArticleService instance with injected repo and logger.
func NewArticleService(repo ArticleRepository) ArticleService {
	return &ArticleServiceImpl{
		repo: repo,
		log:  logger.NewLogger(),
	}
}

// CreateArticle adds a new article to the repository.
func (s *ArticleServiceImpl) CreateArticle(ctx context.Context, dto CreateArticleDTO) (*ReadArticleDTO, error) {
	s.log.Sugar.Infof("creating article for course %s: %s", dto.CourseID, dto.Title)

	article, err := s.repo.Create(ctx, dto)
	if err != nil {
		s.log.Sugar.Errorf("failed to create article: %v", err)
		return nil, err
	}

	s.log.Sugar.Infof("article created: %s", article.ID)
	return article, nil
}

// GetArticleByID retrieves a single article by its ID.
func (s *ArticleServiceImpl) GetArticleByID(ctx context.Context, id string) (*ReadArticleDTO, error) {
	s.log.Sugar.Infof("retrieving article: %s", id)

	article, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.log.Sugar.Warnf("failed to get article %s: %v", id, err)
		return nil, err
	}

	s.log.Sugar.Debugf("retrieved article: %+v", article)
	return article, nil
}

// UpdateArticle modifies an existing article's fields.
func (s *ArticleServiceImpl) UpdateArticle(ctx context.Context, id string, dto UpdateArticleDTO) error {
	s.log.Sugar.Infof("updating article: %s", id)

	err := s.repo.UpdateByID(ctx, id, dto)
	if err != nil {
		s.log.Sugar.Warnf("failed to update article %s: %v", id, err)
		return err
	}

	s.log.Sugar.Infof("article updated: %s", id)
	return nil
}

// DeleteArticle performs a soft-delete on the specified article.
func (s *ArticleServiceImpl) DeleteArticle(ctx context.Context, id string) error {
	s.log.Sugar.Infof("deleting article: %s", id)

	err := s.repo.DeleteByID(ctx, id)
	if err != nil {
		s.log.Sugar.Warnf("failed to delete article %s: %v", id, err)
		return err
	}

	s.log.Sugar.Infof("article deleted: %s", id)
	return nil
}

// FindArticles retrieves a list of articles based on provided filters.
func (s *ArticleServiceImpl) FindArticles(ctx context.Context, filter FilterArticlesDTO) ([]*ReadArticleDTO, error) {
	s.log.Sugar.Infof("finding articles with filter: %+v", filter)

	articles, err := s.repo.Find(ctx, filter)
	if err != nil {
		s.log.Sugar.Errorf("failed to find articles: %v", err)
		return nil, err
	}

	if len(articles) == 0 {
		s.log.Sugar.Infof("no articles matched the given filter")
	} else {
		s.log.Sugar.Infof("found %d article(s)", len(articles))
	}

	return articles, nil
}

// UpdateWithVersioning clones article with incremented version, then applies updates to the clone.
// Returns the updated copy or error.
func (s *ArticleServiceImpl) UpdateWithVersioning(ctx context.Context, id string, dto UpdateArticleDTO) (*ReadArticleDTO, error) {
	s.log.Sugar.Infof("cloning article with version bump: %s", id)

	cloned, err := s.repo.CloneWithIncrementedVersion(ctx, id)
	if err != nil {
		s.log.Sugar.Errorf("failed to clone article %s: %v", id, err)
		return nil, err
	}

	s.log.Sugar.Infof("cloned article %s as %s (v%d)", id, cloned.ID, cloned.Version)

	err = s.repo.UpdateByID(ctx, cloned.ID, dto)
	if err != nil {
		s.log.Sugar.Errorf("failed to apply update to clone %s: %v", cloned.ID, err)
		return nil, err
	}

	s.log.Sugar.Infof("updated cloned article %s", cloned.ID)

	return s.repo.GetByID(ctx, cloned.ID)
}
