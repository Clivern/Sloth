// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package module

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/clivern/sloth/conf"
	"github.com/clivern/sloth/db"
	"github.com/clivern/sloth/pkg/ai"

	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
)

const (
	AssistantMaxRounds    = 3
	AssistantMaxCalls     = 3
	AssistantOutputLength = 8000
	AssistantTitleLength  = 60
	AssistantListLimit    = 50
	AssistantReplyTimeout = 3 * time.Minute

	AssistantEventChat  = "chat"
	AssistantEventDelta = "delta"
	AssistantEventDone  = "done"
	AssistantEventError = "error"
)

var (
	ErrAssistantChatNotFound  = errors.New("assistant chat not found")
	ErrAssistantNotConfigured = errors.New("assistant is not configured")
	ErrAssistantBusy          = errors.New("assistant is still answering")
	ErrFailedAssistantRequest = errors.New("failed assistant request")
)

// Assistant is the module for the dashboard assistant.
type Assistant struct {
	AssistantRepository db.AssistantRepository
	Tools               []*AssistantTool
	Can                 AssistantPermissionCheck
	NewChat             func() ai.ChatClient
}

// AssistantEmit sends one named event with its data to the user while the assistant answers.
type AssistantEmit func(event string, data any)

// AssistantPermissionCheck reports whether a user holds a workspace permission.
type AssistantPermissionCheck func(workspaceId, userId db.Id, permission string) (bool, error)

// SendAssistantMessageRequest is the body for asking the assistant a question.
type SendAssistantMessageRequest struct {
	Content string `json:"content" validate:"required,max=4000" label:"Message"`
}

// AssistantPlan is the tool calls the model picked for a question.
type AssistantPlan struct {
	Calls []AssistantCall `json:"calls"`
}

// AssistantCall is one tool call in a plan.
type AssistantCall struct {
	Tool      string            `json:"tool"`
	Arguments map[string]string `json:"arguments"`
}

// AssistantChatResponse is a chat shaped for API responses.
type AssistantChatResponse struct {
	Id        db.Id                       `json:"id"`
	Title     string                      `json:"title"`
	Status    string                      `json:"status"`
	Preview   string                      `json:"preview,omitempty"`
	Messages  []*AssistantMessageResponse `json:"messages,omitempty"`
	CreatedAt string                      `json:"createdAt"`
	UpdatedAt string                      `json:"updatedAt"`
}

// AssistantMessageResponse is a message shaped for API responses.
type AssistantMessageResponse struct {
	Id        db.Id  `json:"id"`
	Kind      string `json:"kind"`
	Content   string `json:"content"`
	CreatedAt string `json:"createdAt"`
}

// NewAssistant creates an assistant module whose tools run under the user's workspace permissions.
func NewAssistant(assistant db.AssistantRepository, workspaces db.WorkspaceRepository, tools *AssistantTools) *Assistant {
	return &Assistant{
		AssistantRepository: assistant,
		Tools:               tools.List(),
		Can: func(workspaceId, userId db.Id, permission string) (bool, error) {
			perm := &Perm{WorkspaceRepository: workspaces}

			return perm.WithUser(&db.User{Id: userId}).WithWorkspaceId(workspaceId).Can(permission)
		},
		NewChat: ai.NewChatClient,
	}
}

// ListChats returns a user's recent chats in a workspace.
func (a *Assistant) ListChats(workspaceId, userId db.Id) ([]*AssistantChatResponse, error) {
	chats, err := a.AssistantRepository.ListChats(workspaceId, userId, AssistantListLimit, 0)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Msg("Failed to list assistant chats")
		return nil, fmt.Errorf("%w: %v", ErrFailedAssistantRequest, err)
	}

	return lo.Map(chats, func(chat *db.AssistantChat, _ int) *AssistantChatResponse {
		return ToAssistantChatResponse(chat, nil)
	}), nil
}

// CreateChat starts an empty chat for a user.
func (a *Assistant) CreateChat(workspaceId, userId db.Id) (*AssistantChatResponse, error) {
	chat := &db.AssistantChat{
		WorkspaceId: workspaceId,
		UserId:      userId,
		Status:      db.AssistantStatusIdle,
	}

	err := a.AssistantRepository.CreateChat(chat)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Msg("Failed to create assistant chat")
		return nil, fmt.Errorf("%w: %v", ErrFailedAssistantRequest, err)
	}

	return ToAssistantChatResponse(chat, []*db.AssistantMessage{}), nil
}

// GetChat returns a user's chat with its messages.
func (a *Assistant) GetChat(workspaceId, userId, id db.Id) (*AssistantChatResponse, error) {
	chat, err := a.AssistantRepository.GetChat(workspaceId, userId, id)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrFailedAssistantRequest, err)
	}
	if chat == nil {
		return nil, ErrAssistantChatNotFound
	}

	messages, err := a.AssistantRepository.ListMessages(chat.Id)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrFailedAssistantRequest, err)
	}

	return ToAssistantChatResponse(chat, messages), nil
}

// DeleteChat removes a user's chat and its messages. An answer still being written for it is dropped.
func (a *Assistant) DeleteChat(workspaceId, userId, id db.Id) error {
	deleted, err := a.AssistantRepository.DeleteChat(workspaceId, userId, id)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Str("chatId", id.String()).
			Msg("Failed to delete assistant chat")
		return fmt.Errorf("%w: %v", ErrFailedAssistantRequest, err)
	}
	if !deleted {
		return ErrAssistantChatNotFound
	}

	return nil
}

// SendMessage stores the user's question and answers it, emitting the chat and the answer deltas as they
// happen. Errors before the question is stored are returned; later ones are emitted, as the stream has started.
func (a *Assistant) SendMessage(workspaceId, userId, id db.Id, req *SendAssistantMessageRequest, emit AssistantEmit) error {
	if !ai.IsConfigured() {
		return ErrAssistantNotConfigured
	}

	chat, err := a.AssistantRepository.GetChat(workspaceId, userId, id)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrFailedAssistantRequest, err)
	}
	if chat == nil {
		return ErrAssistantChatNotFound
	}
	if chat.Status == db.AssistantStatusRunning {
		return ErrAssistantBusy
	}

	content := strings.TrimSpace(req.Content)

	err = a.AssistantRepository.CreateMessage(&db.AssistantMessage{
		ChatId:  chat.Id,
		Kind:    db.AssistantMessageKindUser,
		Content: content,
	})
	if err != nil {
		return fmt.Errorf("%w: %v", ErrFailedAssistantRequest, err)
	}

	if lo.IsEmpty(chat.Title) {
		chat.Title = lo.Substring(content, 0, AssistantTitleLength)
	}

	chat.Status = db.AssistantStatusRunning

	err = a.AssistantRepository.UpdateChat(chat)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrFailedAssistantRequest, err)
	}

	started, err := a.GetChat(workspaceId, userId, id)
	if err != nil {
		return err
	}

	emit(AssistantEventChat, started)

	a.Reply(chat, emit)

	return nil
}

// Reply answers the chat's latest question, marks the chat idle, or failed when answering broke, and emits the outcome.
// It runs detached from the request, so the answer is still stored when the user leaves mid-stream.
func (a *Assistant) Reply(chat *db.AssistantChat, emit AssistantEmit) {
	ctx, cancel := context.WithTimeout(context.Background(), AssistantReplyTimeout)
	defer cancel()

	err := a.Answer(ctx, chat, emit)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", chat.WorkspaceId.String()).
			Str("chatId", chat.Id.String()).
			Msg("Failed to answer assistant chat")
	}

	chat.Status = lo.Ternary(err == nil, db.AssistantStatusIdle, db.AssistantStatusFailed)

	err = a.AssistantRepository.UpdateChat(chat)
	if err != nil {
		log.Error().
			Err(err).
			Str("chatId", chat.Id.String()).
			Msg("Failed to update assistant chat")
	}

	finished, err := a.GetChat(chat.WorkspaceId, chat.UserId, chat.Id)
	if err != nil {
		emit(AssistantEventError, map[string]string{"status": db.AssistantStatusFailed})
		return
	}

	emit(lo.Ternary(finished.Status == db.AssistantStatusFailed, AssistantEventError, AssistantEventDone), finished)
}

// Answer runs tool rounds until the model needs no more results, then streams and stores its answer; tool results
// only live in the model's context for this answer.
func (a *Assistant) Answer(ctx context.Context, chat *db.AssistantChat, emit AssistantEmit) error {
	messages, err := a.AssistantRepository.ListMessages(chat.Id)
	if err != nil {
		return err
	}

	tools, err := a.AllowedTools(chat)
	if err != nil {
		return err
	}

	model := a.NewChat()
	history := AssistantHistory(messages)
	results := []string{}

	for round := 0; round < AssistantMaxRounds; round++ {
		plan, err := a.Plan(ctx, model, history, tools, results)
		if err != nil {
			return err
		}
		if len(plan.Calls) == 0 {
			break
		}

		for _, call := range plan.Calls {
			result, err := a.Call(chat, tools, call)
			if err != nil {
				return err
			}

			results = append(results, result)
		}
	}

	// Looked-up data goes in the instructions, not the chat, so the model answers from it instead of narrating it.
	prompt := fmt.Sprintf(
		"%s\n\nWhat you can do:\n%s\n\nWorkspace data:\n%s",
		conf.AssistantAnswerPrompt,
		lo.Ternary(len(tools) == 0, "- answer general questions about Sloth", strings.Join(lo.Map(tools, func(tool *AssistantTool, _ int) string {
			return "- " + tool.Description
		}), "\n")),
		lo.Ternary(len(results) == 0, "none needed", strings.Join(results, "\n")),
	)

	answer, _, err := model.Stream(ctx, append([]ai.Message{{Role: "system", Content: prompt}}, history...), func(delta string) {
		emit(AssistantEventDelta, map[string]string{"content": delta})
	})
	if err != nil {
		return err
	}

	return a.AssistantRepository.CreateMessage(&db.AssistantMessage{
		ChatId:  chat.Id,
		Kind:    db.AssistantMessageKindAssistant,
		Content: strings.TrimSpace(answer),
	})
}

// AllowedTools returns the tools the chat's user holds the permission for.
func (a *Assistant) AllowedTools(chat *db.AssistantChat) ([]*AssistantTool, error) {
	allowed := make([]*AssistantTool, 0, len(a.Tools))
	for _, tool := range a.Tools {
		can, err := a.Can(chat.WorkspaceId, chat.UserId, tool.Permission)
		if err != nil {
			return nil, err
		}
		if can {
			allowed = append(allowed, tool)
		}
	}

	return allowed, nil
}

// Plan asks the model which tools to run next; a reply that is not a plan means no calls.
func (a *Assistant) Plan(ctx context.Context, model ai.ChatClient, history []ai.Message, tools []*AssistantTool, results []string) (*AssistantPlan, error) {
	if len(tools) == 0 {
		return &AssistantPlan{}, nil
	}

	choices, _ := json.Marshal(tools)

	reply, _, err := model.Complete(ctx, append(
		append([]ai.Message{{Role: "system", Content: conf.AssistantPlanPrompt}}, history...),
		ai.Message{Role: "user", Content: fmt.Sprintf(
			"Tools you may run: %s\nData already looked up for this message:\n%s",
			choices,
			lo.Ternary(len(results) == 0, "none", strings.Join(results, "\n")),
		)},
	))
	if err != nil {
		return nil, err
	}

	plan := &AssistantPlan{}
	start, end := strings.Index(reply, "{"), strings.LastIndex(reply, "}")
	if start == -1 || end < start {
		return plan, nil
	}

	_ = json.Unmarshal([]byte(reply[start:end+1]), plan)

	names := lo.SliceToMap(tools, func(tool *AssistantTool) (string, bool) { return tool.Name, true })
	plan.Calls = lo.Filter(plan.Calls, func(call AssistantCall, _ int) bool { return names[call.Tool] })
	plan.Calls = lo.Slice(plan.Calls, 0, AssistantMaxCalls)

	return plan, nil
}

// Call runs one tool for the chat's user and returns its outcome as a line of workspace data for the model.
func (a *Assistant) Call(chat *db.AssistantChat, tools []*AssistantTool, call AssistantCall) (string, error) {
	tool, _ := lo.Find(tools, func(tool *AssistantTool) bool { return tool.Name == call.Tool })
	arguments := lo.CoalesceMapOrEmpty(call.Arguments)
	input, _ := json.Marshal(arguments)

	// The user's role may have changed since the tools were offered.
	can, err := a.Can(chat.WorkspaceId, chat.UserId, tool.Permission)
	if err != nil {
		return "", err
	}
	if !can {
		return fmt.Sprintf("%s %s: the user is not allowed to see this", call.Tool, input), nil
	}

	result, err := tool.Run(chat.WorkspaceId, chat.UserId, arguments)
	if err != nil {
		return fmt.Sprintf("%s %s: could not be read (%s)", call.Tool, input, err), nil
	}

	output, _ := json.Marshal(result)

	return fmt.Sprintf(
		"%s %s: %s",
		call.Tool,
		input,
		lo.Substring(string(output), 0, AssistantOutputLength),
	), nil
}

// AssistantHistory turns stored messages into model messages.
func AssistantHistory(messages []*db.AssistantMessage) []ai.Message {
	return lo.Map(messages, func(message *db.AssistantMessage, _ int) ai.Message {
		return ai.Message{Role: message.Kind, Content: message.Content}
	})
}

// ToAssistantChatResponse shapes a chat and, when given, its messages for API responses.
func ToAssistantChatResponse(chat *db.AssistantChat, messages []*db.AssistantMessage) *AssistantChatResponse {
	response := &AssistantChatResponse{
		Id:        chat.Id,
		Title:     chat.Title,
		Status:    chat.Status,
		Preview:   lo.FromPtr(chat.Preview),
		CreatedAt: chat.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt: chat.UpdatedAt.UTC().Format(time.RFC3339),
	}

	if messages != nil {
		response.Messages = lo.Map(messages, func(message *db.AssistantMessage, _ int) *AssistantMessageResponse {
			return &AssistantMessageResponse{
				Id:        message.Id,
				Kind:      message.Kind,
				Content:   message.Content,
				CreatedAt: message.CreatedAt.UTC().Format(time.RFC3339),
			}
		})
	}

	return response
}
