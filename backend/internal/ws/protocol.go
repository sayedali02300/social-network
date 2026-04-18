package ws

import (
	"encoding/json"
	"errors"
	"strings"
)

const (
	maxMessageContentBytes = 65536

	TypePrivateMessageSend = "private_message_send"
	TypeGroupMessageSend   = "group_message_send"
	TypePing               = "ping"

	TypeAck            = "ack"
	TypeError          = "error"
	TypePrivateMessage = "private_message"
	TypeGroupMessage   = "group_message"

	TypeFollowerRemoved = "follower_removed"
	TypeFollowerAdded   = "follower_added"

	TypePrivateTyping = "private_typing"
	TypeGroupTyping   = "group_typing"
	TypeTypingEvent   = "typing_event"
)

type IncomingEnvelope struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type OutgoingEnvelope struct {
	Type    string `json:"type"`
	Payload any    `json:"payload"`
}

type ErrorPayload struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type AckPayload struct {
	For string `json:"for"`
}

type PrivateMessageSendPayload struct {
	ToUserID string `json:"to_user_id"`
	Content  string `json:"content"`
}

type PrivateMessageEventPayload struct {
	ID         string `json:"id"`
	SenderID   string `json:"sender_id"`
	ReceiverID string `json:"receiver_id"`
	Content    string `json:"content"`
	CreatedAt  string `json:"created_at"`
}

type GroupMessageEventPayload struct {
	ID        string `json:"id"`
	SenderID  string `json:"sender_id"`
	GroupID   string `json:"group_id"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
}

type GroupMessageSendPayload struct {
	GroupID string `json:"group_id"`
	Content string `json:"content"`
}

type PrivateTypingPayload struct {
	ToUserID string `json:"to_user_id"`
	IsTyping bool   `json:"is_typing"`
}

type GroupTypingPayload struct {
	GroupID  string `json:"group_id"`
	IsTyping bool   `json:"is_typing"`
}

type TypingEventPayload struct {
	SenderID    string `json:"sender_id"`
	ContextType string `json:"context_type"` // "private" or "group"
	ContextID   string `json:"context_id"`   // peer user ID (private) or group ID (group)
	IsTyping    bool   `json:"is_typing"`
}

func decodeIncoming(data []byte) (IncomingEnvelope, error) {
	var env IncomingEnvelope
	if err := json.Unmarshal(data, &env); err != nil {
		return IncomingEnvelope{}, errors.New("invalid JSON envelope")
	}
	env.Type = strings.TrimSpace(env.Type)
	if env.Type == "" {
		return IncomingEnvelope{}, errors.New("missing message type")
	}
	return env, nil
}

func parsePrivateMessagePayload(raw json.RawMessage) (PrivateMessageSendPayload, error) {
	var p PrivateMessageSendPayload
	if err := json.Unmarshal(raw, &p); err != nil {
		return PrivateMessageSendPayload{}, errors.New("invalid private message payload")
	}

	p.ToUserID = strings.TrimSpace(p.ToUserID)
	p.Content = strings.TrimSpace(p.Content)
	if p.ToUserID == "" {
		return PrivateMessageSendPayload{}, errors.New("to_user_id is required")
	}
	if p.Content == "" {
		return PrivateMessageSendPayload{}, errors.New("content is required")
	}
	if len(p.Content) > maxMessageContentBytes {
		return PrivateMessageSendPayload{}, errors.New("content exceeds maximum size")
	}
	return p, nil
}

func parseGroupMessagePayload(raw json.RawMessage) (GroupMessageSendPayload, error) {
	var p GroupMessageSendPayload
	if err := json.Unmarshal(raw, &p); err != nil {
		return GroupMessageSendPayload{}, errors.New("invalid group message payload")
	}

	p.GroupID = strings.TrimSpace(p.GroupID)
	p.Content = strings.TrimSpace(p.Content)
	if p.GroupID == "" {
		return GroupMessageSendPayload{}, errors.New("group_id is required")
	}
	if p.Content == "" {
		return GroupMessageSendPayload{}, errors.New("content is required")
	}
	if len(p.Content) > maxMessageContentBytes {
		return GroupMessageSendPayload{}, errors.New("content exceeds maximum size")
	}
	return p, nil
}
