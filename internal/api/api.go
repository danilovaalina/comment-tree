package api

import (
	"context"
	"net/http"
	"time"

	"comment-tree/internal/model"

	"github.com/cockroachdb/errors"
	"github.com/labstack/echo/v4"
)

type Service interface {
	CreateComment(ctx context.Context, c model.Comment) (*model.Comment, error)
	DeleteComment(ctx context.Context, id int64) error
	CommentTree(ctx context.Context, rootID int64) (*model.Comment, error)
	CommentsTree(ctx context.Context, filter model.CommentFilter) ([]*model.Comment, error)
	SearchComments(ctx context.Context, query string) ([]*model.Comment, error)
}

type API struct {
	*echo.Echo
	service Service
}

func New(service Service) *API {
	a := &API{
		Echo:    echo.New(),
		service: service,
	}
	a.Validator = NewCustomValidator()
	a.HideBanner = true

	a.Static("/", "web")

	g := a.Group("/comments")

	g.POST("", a.createComment)
	g.GET("", a.commentsTree)
	g.GET("/search", a.searchComments)
	g.GET("/:id", a.commentTree)
	g.DELETE("/:id", a.deleteComment)

	return a
}

type createCommentRequest struct {
	ParentID *int64 `json:"parent_id" validate:"omitempty"`
	Content  string `json:"content" validate:"required,min=1"`
}

func (a *API) createComment(c echo.Context) error {
	var req createCommentRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid format"})
	}

	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}

	newComment := model.Comment{
		ParentID: req.ParentID,
		Content:  req.Content,
	}

	res, err := a.service.CreateComment(c.Request().Context(), newComment)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "failed to create comment"})
	}

	return c.JSON(http.StatusCreated, a.commentToResponse(res))
}

type commentParam struct {
	ID int64 `param:"id" validate:"gt=0"`
}

func (a *API) deleteComment(c echo.Context) error {
	var param commentParam
	if err := c.Bind(&param); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid id format"})
	}

	if err := c.Validate(&param); err != nil {
		return c.JSON(http.StatusBadRequest, a.validationError(err))
	}

	if err := a.service.DeleteComment(c.Request().Context(), param.ID); err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return c.JSON(http.StatusNotFound, echo.Map{"error": "comment not found"})
		}
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "failed to delete comment"})
	}

	return c.NoContent(http.StatusNoContent)
}

func (a *API) commentTree(c echo.Context) error {
	var param commentParam
	if err := c.Bind(&param); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid id format"})
	}

	if err := c.Validate(&param); err != nil {
		return c.JSON(http.StatusBadRequest, a.validationError(err))
	}

	tree, err := a.service.CommentTree(c.Request().Context(), param.ID)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return c.JSON(http.StatusNotFound, echo.Map{"error": "comment not found"})
		}
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "failed to fetch comment tree"})
	}

	return c.JSON(http.StatusOK, a.commentToResponse(tree))
}

type searchParams struct {
	Query string `query:"q" validate:"required,min=2"`
}

func (a *API) searchComments(c echo.Context) error {
	var params searchParams
	if err := c.Bind(&params); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid search params"})
	}
	if err := c.Validate(&params); err != nil {
		return c.JSON(http.StatusBadRequest, a.validationError(err))
	}

	results, err := a.service.SearchComments(c.Request().Context(), params.Query)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "failed to search comments"})
	}

	response := make([]*commentResponse, 0, len(results))
	for _, res := range results {
		response = append(response, a.commentToResponse(res))
	}

	return c.JSON(http.StatusOK, response)
}

type commentsParams struct {
	Page  uint64 `query:"page" validate:"omitempty,gt=0"`
	Limit uint64 `query:"limit" validate:"omitempty,gt=0"`
}

func (a *API) commentsTree(c echo.Context) error {
	var params commentsParams

	if err := c.Bind(&params); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid params"})
	}

	if err := c.Validate(&params); err != nil {
		return c.JSON(http.StatusBadRequest, a.validationError(err))
	}

	filter := model.CommentFilter{
		Page:  params.Page,
		Limit: params.Limit,
	}

	trees, err := a.service.CommentsTree(c.Request().Context(), filter)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "failed to fetch comments"})
	}

	response := make([]*commentResponse, 0, len(trees))
	for _, t := range trees {
		response = append(response, a.commentToResponse(t))
	}

	return c.JSON(http.StatusOK, response)
}

func (a *API) commentToResponse(c *model.Comment) *commentResponse {
	resp := &commentResponse{
		ID:       c.ID,
		ParentID: c.ParentID,
		Content:  c.Content,
		// Форматируем время в строку
		Created: c.Created,
		// Сразу инициализируем пустым слайсом указателей
		Children: []*commentResponse{},
	}

	for _, child := range c.Children {
		// Рекурсивный вызов: передаем указатель на ребенка
		// и добавляем результат (тоже указатель) в слайс Children ответа
		resp.Children = append(resp.Children, a.commentToResponse(child))
	}

	return resp
}

type commentResponse struct {
	ID       int64              `json:"id"`
	ParentID *int64             `json:"parent_id,omitempty"`
	Content  string             `json:"content"`
	Created  time.Time          `json:"created"`
	Children []*commentResponse `json:"children"`
}
