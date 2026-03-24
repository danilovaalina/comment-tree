package repository

import (
	"context"
	"time"

	"comment-tree/internal/model"

	"github.com/cockroachdb/errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

type commentRow struct {
	ID       int64     `db:"id"`
	ParentID *int64    `db:"parent_id"`
	Content  string    `db:"content"`
	Created  time.Time `db:"created"`
}

func (r *Repository) toModel(row commentRow) model.Comment {
	return model.Comment{
		ID:       row.ID,
		ParentID: row.ParentID,
		Content:  row.Content,
		Created:  row.Created,
	}
}

func (r *Repository) CreateComment(ctx context.Context, c model.Comment) (*model.Comment, error) {
	query := `
		insert into comments (parent_id, content)
		values ($1, $2)
		returning id, parent_id, content, created
	`
	rows, err := r.pool.Query(ctx, query, c.ParentID, c.Content)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer rows.Close()

	row, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByNameLax[commentRow])
	if err != nil {
		return nil, errors.WithStack(err)
	}

	comment := r.toModel(row)

	return &comment, nil
}

func (r *Repository) DeleteComment(ctx context.Context, id int64) error {
	query := `delete from comments where id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return errors.WithStack(err)
	}

	if result.RowsAffected() == 0 {
		return model.ErrNotFound
	}

	return nil
}

func (r *Repository) CommentTree(ctx context.Context, rootID int64) ([]*model.Comment, error) {
	query := `
	with recursive comment_tree as (
		select id, parent_id, content, created
		from comments
		where id = $1
		union all

		select c.id, c.parent_id, c.content, c.created
		from comments c
		join comment_tree ct on c.parent_id = ct.id
	)
	select id, parent_id, content, created from comment_tree order by created asc;
	`

	rows, err := r.pool.Query(ctx, query, rootID)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer rows.Close()

	commentRows, err := pgx.CollectRows[commentRow](rows, pgx.RowToStructByNameLax[commentRow])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.WithStack(model.ErrNotFound)
		}
		return nil, errors.WithStack(err)
	}

	comments := make([]*model.Comment, 0, len(commentRows))
	for _, row := range commentRows {
		m := r.toModel(row)
		comments = append(comments, &m)
	}

	return comments, nil
}
