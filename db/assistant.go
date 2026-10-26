// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package db

import (
	"database/sql"
	"errors"
	"time"
)

const (
	AssistantStatusIdle    = "idle"
	AssistantStatusRunning = "running"
	AssistantStatusFailed  = "failed"

	AssistantMessageKindUser      = "user"
	AssistantMessageKindAssistant = "assistant"
)

// AssistantChat is a single row in the ai_assistant_chats table.
type AssistantChat struct {
	Id          Id
	WorkspaceId Id
	UserId      Id
	Title       string
	Status      string
	Preview     *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// AssistantMessage is a single row in the ai_assistant_messages table.
type AssistantMessage struct {
	Id        Id
	ChatId    Id
	Kind      string
	Content   string
	Meta      *string
	CreatedAt time.Time
}

// AssistantRepository is the interface for assistant chat persistence.
type AssistantRepository interface {
	CreateChat(chat *AssistantChat) error
	GetChat(workspaceId, userId, id Id) (*AssistantChat, error)
	ListChats(workspaceId, userId Id, limit, offset int) ([]*AssistantChat, error)
	UpdateChat(chat *AssistantChat) error
	DeleteChat(workspaceId, userId, id Id) (bool, error)
	CreateMessage(message *AssistantMessage) error
	ListMessages(chatId Id) ([]*AssistantMessage, error)
}

type AssistantRepositoryPostgres struct {
	db *sql.DB
}

// NewAssistantRepository returns an assistant repository.
func NewAssistantRepository(db *sql.DB) AssistantRepository {
	return &AssistantRepositoryPostgres{db: db}
}

// CreateChat inserts a chat row.
func (r *AssistantRepositoryPostgres) CreateChat(chat *AssistantChat) error {
	id, err := NewId()
	if err != nil {
		return err
	}

	chat.Id = id

	return r.db.QueryRow(
		`INSERT INTO ai_assistant_chats (id, workspace_id, user_id, title, status)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING created_at, updated_at`,
		chat.Id.String(),
		chat.WorkspaceId.String(),
		chat.UserId.String(),
		chat.Title,
		chat.Status,
	).Scan(&chat.CreatedAt, &chat.UpdatedAt)
}

// GetChat returns a user's chat in a workspace.
func (r *AssistantRepositoryPostgres) GetChat(workspaceId, userId, id Id) (*AssistantChat, error) {
	chat := &AssistantChat{}
	err := r.db.QueryRow(
		`SELECT id, workspace_id, user_id, title, status, created_at, updated_at
		FROM ai_assistant_chats
		WHERE workspace_id = $1 AND user_id = $2 AND id = $3`,
		workspaceId.String(),
		userId.String(),
		id.String(),
	).Scan(
		&chat.Id,
		&chat.WorkspaceId,
		&chat.UserId,
		&chat.Title,
		&chat.Status,
		&chat.CreatedAt,
		&chat.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	return chat, err
}

// ListChats returns a user's chats in a workspace with their last message, most recently active first.
func (r *AssistantRepositoryPostgres) ListChats(workspaceId, userId Id, limit, offset int) ([]*AssistantChat, error) {
	rows, err := r.db.Query(
		`SELECT
			c.id, c.workspace_id, c.user_id, c.title, c.status,
			(
				SELECT m.content
				FROM ai_assistant_messages m
				WHERE m.chat_id = c.id
				ORDER BY m.created_at DESC, m.id DESC
				LIMIT 1
			),
			c.created_at, c.updated_at
		FROM ai_assistant_chats c
		WHERE c.workspace_id = $1 AND c.user_id = $2
		ORDER BY c.updated_at DESC
		LIMIT $3 OFFSET $4`,
		workspaceId.String(),
		userId.String(),
		limit,
		offset,
	)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var list []*AssistantChat
	for rows.Next() {
		chat := &AssistantChat{}
		err := rows.Scan(
			&chat.Id,
			&chat.WorkspaceId,
			&chat.UserId,
			&chat.Title,
			&chat.Status,
			&chat.Preview,
			&chat.CreatedAt,
			&chat.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		list = append(list, chat)
	}

	return list, rows.Err()
}

// UpdateChat updates a chat's title and status and marks it active now.
func (r *AssistantRepositoryPostgres) UpdateChat(chat *AssistantChat) error {
	chat.UpdatedAt = time.Now().UTC()

	_, err := r.db.Exec(
		`UPDATE ai_assistant_chats
		SET title = $1, status = $2, updated_at = $3
		WHERE id = $4`,
		chat.Title,
		chat.Status,
		chat.UpdatedAt,
		chat.Id.String(),
	)

	return err
}

// DeleteChat removes a user's chat and its messages and reports whether it existed.
func (r *AssistantRepositoryPostgres) DeleteChat(workspaceId, userId, id Id) (bool, error) {
	result, err := r.db.Exec(
		`DELETE FROM ai_assistant_chats
		WHERE workspace_id = $1 AND user_id = $2 AND id = $3`,
		workspaceId.String(),
		userId.String(),
		id.String(),
	)
	if err != nil {
		return false, err
	}

	deleted, err := result.RowsAffected()

	return deleted > 0, err
}

// CreateMessage inserts a message row.
func (r *AssistantRepositoryPostgres) CreateMessage(message *AssistantMessage) error {
	id, err := NewId()
	if err != nil {
		return err
	}

	message.Id = id

	return r.db.QueryRow(
		`INSERT INTO ai_assistant_messages (id, chat_id, kind, content, meta)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING created_at`,
		message.Id.String(),
		message.ChatId.String(),
		message.Kind,
		message.Content,
		message.Meta,
	).Scan(&message.CreatedAt)
}

// ListMessages returns a chat's messages, oldest first.
func (r *AssistantRepositoryPostgres) ListMessages(chatId Id) ([]*AssistantMessage, error) {
	rows, err := r.db.Query(
		`SELECT id, chat_id, kind, content, meta, created_at
		FROM ai_assistant_messages
		WHERE chat_id = $1
		ORDER BY created_at, id`,
		chatId.String(),
	)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var list []*AssistantMessage
	for rows.Next() {
		message := &AssistantMessage{}
		err := rows.Scan(
			&message.Id,
			&message.ChatId,
			&message.Kind,
			&message.Content,
			&message.Meta,
			&message.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		list = append(list, message)
	}

	return list, rows.Err()
}
