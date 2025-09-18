package domain

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

// Comment представляет комментарий к инициативе
type Comment struct {
	ID        uuid.UUID `json:"id"`
	Text      string    `json:"text"`
	Author    UserBrief `json:"author"`
	CreatedAt time.Time `json:"createdAt"`
}

// CommentCreate запрос на создание комментария
type CommentCreate struct {
	Text string `json:"text"`
}

// Validate проверяет корректность создаваемого комментария
func (c *CommentCreate) Validate() error {
	text := strings.TrimSpace(c.Text)
	if len(text) < 1 || len(text) > 1000 {
		return ErrInvalidField("text", "must be between 1 and 1000 characters")
	}
	return nil
}

// CommentsList список комментариев с пагинацией
type CommentsList struct {
	Items  []Comment `json:"items"`
	Total  int       `json:"total"`
	Limit  int       `json:"limit"`
	Offset int       `json:"offset"`
}
