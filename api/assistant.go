// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/clivern/sloth/db"
	"github.com/clivern/sloth/locale"
	"github.com/clivern/sloth/module"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
)

// ListAssistantChatsAction lists the user's assistant chats in a workspace.
func (a *API) ListAssistantChatsAction(w http.ResponseWriter, r *http.Request) {
	wid := chi.URLParam(r, "workspaceId")
	user, _ := a.GetUser(r)

	chats, err := a.Assistant.ListChats(db.Id(wid), user.Id)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", wid).
			Msg("Failed to list assistant chats")
		a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
			"errorMessage": locale.TR(r, "failed_assistant_request"),
		})

		return
	}

	a.WriteJSON(w, http.StatusOK, map[string]any{
		"chats": chats,
	})
}

// CreateAssistantChatAction starts an empty assistant chat for the user.
func (a *API) CreateAssistantChatAction(w http.ResponseWriter, r *http.Request) {
	wid := chi.URLParam(r, "workspaceId")
	user, _ := a.GetUser(r)

	chat, err := a.Assistant.CreateChat(db.Id(wid), user.Id)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", wid).
			Msg("Failed to create assistant chat")
		a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
			"errorMessage": locale.TR(r, "failed_assistant_request"),
		})

		return
	}

	a.WriteJSON(w, http.StatusCreated, chat)
}

// GetAssistantChatAction returns one of the user's chats with its messages.
func (a *API) GetAssistantChatAction(w http.ResponseWriter, r *http.Request) {
	wid := chi.URLParam(r, "workspaceId")
	user, _ := a.GetUser(r)

	chatId := chi.URLParam(r, "chatId")
	if lo.IsEmpty(chatId) {
		a.WriteJSON(w, http.StatusBadRequest, map[string]any{
			"errorMessage": locale.TR(r, "invalid_assistant_chat_id"),
		})
		return
	}

	chat, err := a.Assistant.GetChat(db.Id(wid), user.Id, db.Id(chatId))
	if err != nil {
		if a.WriteError(w, err, []ErrorMapping{
			{module.ErrAssistantChatNotFound, http.StatusNotFound, locale.TR(r, "assistant_chat_not_found")},
		}) {
			return
		}

		log.Error().
			Err(err).
			Str("workspaceId", wid).
			Str("chatId", chatId).
			Msg("Failed to get assistant chat")
		a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
			"errorMessage": locale.TR(r, "failed_assistant_request"),
		})

		return
	}

	a.WriteJSON(w, http.StatusOK, chat)
}

// DeleteAssistantChatAction removes one of the user's chats with its messages.
func (a *API) DeleteAssistantChatAction(w http.ResponseWriter, r *http.Request) {
	wid := chi.URLParam(r, "workspaceId")
	user, _ := a.GetUser(r)

	chatId := chi.URLParam(r, "chatId")
	if lo.IsEmpty(chatId) {
		a.WriteJSON(w, http.StatusBadRequest, map[string]any{
			"errorMessage": locale.TR(r, "invalid_assistant_chat_id"),
		})
		return
	}

	log.Info().
		Str("workspaceId", wid).
		Str("chatId", chatId).
		Msg("Deleting assistant chat")

	err := a.Assistant.DeleteChat(db.Id(wid), user.Id, db.Id(chatId))
	if err != nil {
		if a.WriteError(w, err, []ErrorMapping{
			{module.ErrAssistantChatNotFound, http.StatusNotFound, locale.TR(r, "assistant_chat_not_found")},
		}) {
			return
		}

		log.Error().
			Err(err).
			Str("workspaceId", wid).
			Str("chatId", chatId).
			Msg("Failed to delete assistant chat")
		a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
			"errorMessage": locale.TR(r, "failed_assistant_request"),
		})

		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// SendAssistantMessageAction stores the user's question and streams the answer as server-sent events.
func (a *API) SendAssistantMessageAction(w http.ResponseWriter, r *http.Request) {
	wid := chi.URLParam(r, "workspaceId")
	user, _ := a.GetUser(r)

	chatId := chi.URLParam(r, "chatId")
	if lo.IsEmpty(chatId) {
		a.WriteJSON(w, http.StatusBadRequest, map[string]any{
			"errorMessage": locale.TR(r, "invalid_assistant_chat_id"),
		})
		return
	}

	var req module.SendAssistantMessageRequest
	err := a.DecodeAndValidate(r, &req)
	if err != nil {
		a.WriteValidationError(w, err)
		return
	}

	// An answer can outlive the server write timeout, so this response gets its own deadline.
	controller := http.NewResponseController(w)
	_ = controller.SetWriteDeadline(time.Now().Add(module.AssistantReplyTimeout + time.Minute))

	streaming := false
	emit := func(event string, data any) {
		if !streaming {
			w.Header().Set("Content-Type", "text/event-stream")
			w.Header().Set("Cache-Control", "no-cache")
			w.Header().Set("X-Accel-Buffering", "no")
			w.WriteHeader(http.StatusOK)
			streaming = true
		}

		body, _ := json.Marshal(data)
		fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, body)
		_ = controller.Flush()
	}

	err = a.Assistant.SendMessage(db.Id(wid), user.Id, db.Id(chatId), &req, emit)
	if err != nil {
		if a.WriteError(w, err, []ErrorMapping{
			{module.ErrAssistantChatNotFound, http.StatusNotFound, locale.TR(r, "assistant_chat_not_found")},
			{module.ErrAssistantNotConfigured, http.StatusServiceUnavailable, locale.TR(r, "assistant_not_configured")},
			{module.ErrAssistantBusy, http.StatusConflict, locale.TR(r, "assistant_busy")},
		}) {
			return
		}

		log.Error().
			Err(err).
			Str("workspaceId", wid).
			Str("chatId", chatId).
			Msg("Failed to send assistant message")
		a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
			"errorMessage": locale.TR(r, "failed_assistant_request"),
		})
	}
}
