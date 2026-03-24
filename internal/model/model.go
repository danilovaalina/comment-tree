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

var (
	ErrNotFound = errors.New("comment not found")
)
