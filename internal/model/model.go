package model

import (
	"time"

	"github.com/cockroachdb/errors"
)

type Comment struct {
	ID       int64
	ParentID *int64
	Content  string
	Created  time.Time
	Children []*Comment // Для формирования дерева в памяти (если нужно)
}

type CommentFilter struct {
	Page   uint64
	Limit  uint64
	Offset uint64
}

var (
	ErrNotFound = errors.New("comment not found")
)
