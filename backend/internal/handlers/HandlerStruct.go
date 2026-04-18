package handlers

import (
	"encoding/json"
	"net/http"

	"socialnetwork/backend/internal/repository"
	wshandler "socialnetwork/backend/internal/ws"
)

type UserEventEmitter interface {
	EmitToUser(userID string, message wshandler.OutgoingEnvelope)
}
type UsersEventEmitter interface{
	EmitToUsers(userIDs []string, message wshandler.OutgoingEnvelope)
}
type HandlerStruct struct {
	CommentRepo  *repository.CommentRepository
	PostRepo     *repository.PostRepository
	UserRepo     *repository.UserRepository
	SessionStore *repository.SessionStore
	GroupRepo    *repository.GroupRepository
	LikeRepo     *repository.LikeRepository
	Notifier     UserEventEmitter
	NotifierMany UsersEventEmitter
}

func NewHandler(
	commentRepo *repository.CommentRepository,
	postRepo *repository.PostRepository,
	userRepo *repository.UserRepository,
	sessionStore *repository.SessionStore,
	groupRepo ...*repository.GroupRepository,
) *HandlerStruct {
	h := &HandlerStruct{
		CommentRepo: commentRepo,
		PostRepo:     postRepo,
		UserRepo:     userRepo,
		SessionStore: sessionStore,
	}
	if len(groupRepo) > 0 {
		h.GroupRepo = groupRepo[0]
	}
	return h
}

func (h *HandlerStruct) SetNotifier(notifier UserEventEmitter) {
	h.Notifier = notifier
}
func (h *HandlerStruct) SetNotifierMany(notifier UsersEventEmitter) {
	h.NotifierMany = notifier
}

// CreateNotification is the exported entry point for the ws layer to create
// persistent notifications without importing the handlers package.
func (h *HandlerStruct) CreateNotification(userID, actorID, notifType, payload string) {
	h.createNotification(userID, actorID, notifType, payload)
}
func sendJSONError(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}
