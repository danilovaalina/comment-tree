package service

import (
	"context"

	"comment-tree/internal/model"
)

type Repository interface {
	CreateComment(ctx context.Context, c model.Comment) (*model.Comment, error)
	DeleteComment(ctx context.Context, id int64) error
	CommentTree(ctx context.Context, rootID int64) ([]*model.Comment, error)
}

type Service struct {
	repo Repository
}

func New(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateComment(ctx context.Context, c model.Comment) (*model.Comment, error) {
	return s.repo.CreateComment(ctx, c)
}

func (s *Service) DeleteComment(ctx context.Context, id int64) error {
	return s.repo.DeleteComment(ctx, id)
}

func (s *Service) CommentTree(ctx context.Context, rootID int64) (*model.Comment, error) {
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
