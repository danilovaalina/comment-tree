package service

import (
	"context"

	"comment-tree/internal/model"
)

type Repository interface {
	CreateComment(ctx context.Context, c model.Comment) (*model.Comment, error)
	DeleteComment(ctx context.Context, id int64) error
	CommentTree(ctx context.Context, rootID int64) ([]*model.Comment, error)
	SearchComments(ctx context.Context, query string) ([]*model.Comment, error)
	CommentsTree(ctx context.Context, filter model.CommentFilter) ([]*model.Comment, error)
}

type Service struct {
	repo Repository
	opts Options
}

type Options struct {
	DefaultLimit uint64
	MaxLimit     uint64
}

func New(repo Repository, opts Options) *Service {
	return &Service{repo: repo, opts: opts}
}

func (s *Service) CreateComment(ctx context.Context, c model.Comment) (*model.Comment, error) {
	return s.repo.CreateComment(ctx, c)
}

func (s *Service) DeleteComment(ctx context.Context, id int64) error {
	return s.repo.DeleteComment(ctx, id)
}

func (s *Service) CommentTree(ctx context.Context, rootID int64) (*model.Comment, error) {
	// Получаем плоский список (корень + его потомки)
	tree, err := s.repo.CommentTree(ctx, rootID)
	if err != nil {
		return nil, err
	}

	// Создаем карту для быстрого доступа к узлам по ID
	// Используем указатели, чтобы изменения в детях отражались в основном дереве
	nodes := make(map[int64]*model.Comment)
	for _, c := range tree {
		c.Children = []*model.Comment{}
		nodes[c.ID] = c
	}

	var root *model.Comment
	// 3. Распределяем детей по родителям
	for _, c := range nodes {
		if c.ID == rootID {
			root = c // Нашли вершину нашего дерева
			continue
		}

		// Если у комментария есть родитель и он присутствует в нашей выборке
		if c.ParentID != nil {
			if parent, ok := nodes[*c.ParentID]; ok {
				// разыменовываем c, так как в модели Children хранятся значения
				parent.Children = append(parent.Children, c)
			}
		}
	}

	return root, nil
}

func (s *Service) CommentsTree(ctx context.Context, filter model.CommentFilter) ([]*model.Comment, error) {
	if filter.Limit <= 0 {
		filter.Limit = s.opts.DefaultLimit
	}

	if filter.Limit > s.opts.MaxLimit {
		filter.Limit = s.opts.MaxLimit
	}

	if filter.Page < 1 {
		filter.Page = 1
	}

	filter.Offset = (filter.Page - 1) * filter.Limit

	// Получаем плоский список (корни + их потомки)
	tree, err := s.repo.CommentsTree(ctx, filter)
	if err != nil {
		return nil, err
	}

	nodes := make(map[int64]*model.Comment)
	for _, c := range tree {
		// Инициализируем слайс детей, чтобы в JSON не было null
		c.Children = []*model.Comment{}
		nodes[c.ID] = c
	}

	// 3. Слайс для хранения только корневых комментариев
	var roots []*model.Comment

	for _, c := range nodes {
		if c.ParentID == nil {
			// Если родителя нет, это корень дерева, добавляем в итоговый список
			roots = append(roots, c)
		} else {
			// Если родитель есть, ищем его в мапе и добавляем текущий коммент ему в "дети"
			if parent, ok := nodes[*c.ParentID]; ok {
				parent.Children = append(parent.Children, c)
			}
		}
	}

	return roots, nil
}

func (s *Service) SearchComments(ctx context.Context, query string) ([]*model.Comment, error) {
	if query == "" {
		return []*model.Comment{}, nil
	}

	return s.repo.SearchComments(ctx, query)
}
