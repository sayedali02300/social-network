package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"socialnetwork/backend/internal/models"
	"socialnetwork/backend/internal/repository"
	wshandler "socialnetwork/backend/internal/ws"
	"strings"
	"time"
)

type createFollowRequestRequest struct {
	ReceiverID string `json:"receiverId"`
}

type patchFollowRequestRequest struct {
	Status string `json:"status"`
}

func (h *HandlerStruct) createNotification(userID, actorID, notificationType, payload string) {
	if h == nil || h.UserRepo == nil {
		return
	}

	notification, err := h.UserRepo.CreateNotification(userID, actorID, notificationType, payload)
	if err != nil {
		log.Printf("Failed to create notification type=%s user=%s actor=%s: %v", notificationType, userID, actorID, err)
		return
	}

	h.emitNotificationRealtime(notification)
	h.emitNotificationUnreadCountRealtime(userID)
}

type followRequestUser struct {
	ID        string `json:"id"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Nickname  string `json:"nickname"`
	Avatar    string `json:"avatar"`
}

type followRequestResponse struct {
	ID         string             `json:"id"`
	SenderID   string             `json:"senderId"`
	ReceiverID string             `json:"receiverId"`
	Status     string             `json:"status"`
	CreatedAt  string             `json:"createdAt"`
	Sender     *followRequestUser `json:"sender,omitempty"`
	Receiver   *followRequestUser `json:"receiver,omitempty"`
}

func toFollowRequestUser(user *models.User) *followRequestUser {
	if user == nil {
		return nil
	}
	return &followRequestUser{
		ID:        user.ID,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Nickname:  user.Nickname,
		Avatar:    user.Avatar,
	}
}

func (h *HandlerStruct) CreateFollowRequestHandler(w http.ResponseWriter, r *http.Request) {
	if h.UserRepo == nil {
		sendJSONError(w, "User repository is not configured", http.StatusInternalServerError)
		return
	}

	senderID, ok := userIDFromRequest(r)
	if !ok {
		sendJSONError(w, "Not authenticated", http.StatusUnauthorized)
		return
	}

	var req createFollowRequestRequest
	decoder := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		sendJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	req.ReceiverID = strings.TrimSpace(req.ReceiverID)
	if req.ReceiverID == "" {
		sendJSONError(w, "receiverId is required", http.StatusBadRequest)
		return
	}
	if req.ReceiverID == senderID {
		sendJSONError(w, "You cannot follow yourself", http.StatusBadRequest)
		return
	}

	receiver, err := h.UserRepo.GetUserByID(req.ReceiverID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			sendJSONError(w, "Receiver user not found", http.StatusNotFound)
			return
		}
		log.Printf("Error loading receiver user: %v", err)
		sendJSONError(w, "Could not create follow request", http.StatusInternalServerError)
		return
	}

	followRequest, err := h.UserRepo.StartFollowRequest(senderID, req.ReceiverID, receiver.IsPublic)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrAlreadyFollowing):
			sendJSONError(w, "You are already following this user", http.StatusConflict)
			return
		case errors.Is(err, repository.ErrFollowRequestAlreadyPending):
			sendJSONError(w, "Follow request already exists", http.StatusConflict)
			return
		default:
			log.Printf("Error creating follow request: %v", err)
			sendJSONError(w, "Could not create follow request", http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(followRequest)

	if followRequest.Status == "accepted" {
		h.createNotification(
			req.ReceiverID,
			senderID,
			"new_follower",
			followRequest.ID,
		)
		// Emit follower_added via WebSocket to both users
		if h.Notifier != nil {
			payload := wshandler.OutgoingEnvelope{
				Type: wshandler.TypeFollowerAdded,
				Payload: map[string]string{
					"followerId":  senderID,
					"followingId": req.ReceiverID,
				},
			}
			h.Notifier.EmitToUser(senderID, payload)
			h.Notifier.EmitToUser(req.ReceiverID, payload)
		}
	} else {
		h.createNotification(
			req.ReceiverID,
			senderID,
			"follow_request_received",
			followRequest.ID,
		)
	}
}

func (h *HandlerStruct) GetIncomingFollowRequestsHandler(w http.ResponseWriter, r *http.Request) {
	if h.UserRepo == nil {
		sendJSONError(w, "User repository is not configured", http.StatusInternalServerError)
		return
	}

	userID, ok := userIDFromRequest(r)
	if !ok {
		sendJSONError(w, "Not authenticated", http.StatusUnauthorized)
		return
	}

	requests, err := h.UserRepo.GetIncomingFollowRequestsByUserID(userID)
	if err != nil {
		log.Printf("Error loading incoming follow requests: %v", err)
		sendJSONError(w, "Could not load incoming follow requests", http.StatusInternalServerError)
		return
	}
	responses := make([]followRequestResponse, 0, len(requests))
	for _, req := range requests {
		item := followRequestResponse{
			ID:         req.ID,
			SenderID:   req.SenderID,
			ReceiverID: req.ReceiverID,
			Status:     req.Status,
			CreatedAt:  req.CreatedAt.UTC().Format(time.RFC3339),
		}

		sender, senderErr := h.UserRepo.GetUserByID(req.SenderID)
		if senderErr != nil {
			log.Printf("Error loading sender user %s: %v", req.SenderID, senderErr)
		} else {
			item.Sender = toFollowRequestUser(sender)
		}

		responses = append(responses, item)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(responses)
}

func (h *HandlerStruct) GetOutgoingFollowRequestsHandler(w http.ResponseWriter, r *http.Request) {
	if h.UserRepo == nil {
		sendJSONError(w, "User repository is not configured", http.StatusInternalServerError)
		return
	}

	userID, ok := userIDFromRequest(r)
	if !ok {
		sendJSONError(w, "Not authenticated", http.StatusUnauthorized)
		return
	}

	requests, err := h.UserRepo.GetOutgoingFollowRequestsByUserID(userID)
	if err != nil {
		log.Printf("Error loading outgoing follow requests: %v", err)
		sendJSONError(w, "Could not load outgoing follow requests", http.StatusInternalServerError)
		return
	}
	responses := make([]followRequestResponse, 0, len(requests))
	for _, req := range requests {
		item := followRequestResponse{
			ID:         req.ID,
			SenderID:   req.SenderID,
			ReceiverID: req.ReceiverID,
			Status:     req.Status,
			CreatedAt:  req.CreatedAt.UTC().Format(time.RFC3339),
		}

		receiver, receiverErr := h.UserRepo.GetUserByID(req.ReceiverID)
		if receiverErr != nil {
			log.Printf("Error loading receiver user %s: %v", req.ReceiverID, receiverErr)
		} else {
			item.Receiver = toFollowRequestUser(receiver)
		}

		responses = append(responses, item)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(responses)
}

func (h *HandlerStruct) PatchFollowRequestHandler(w http.ResponseWriter, r *http.Request) {
	if h.UserRepo == nil {
		sendJSONError(w, "User repository is not configured", http.StatusInternalServerError)
		return
	}

	receiverID, ok := userIDFromRequest(r)
	if !ok {
		sendJSONError(w, "Not authenticated", http.StatusUnauthorized)
		return
	}

	requestID := strings.TrimSpace(r.PathValue("requestId"))
	if requestID == "" {
		sendJSONError(w, "requestId is required", http.StatusBadRequest)
		return
	}

	var req patchFollowRequestRequest
	decoder := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		sendJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	status := strings.ToLower(strings.TrimSpace(req.Status))
	if status != "accepted" && status != "declined" {
		sendJSONError(w, "status must be accepted or declined", http.StatusBadRequest)
		return
	}

	if err := h.UserRepo.ResolveFollowRequest(requestID, receiverID, status); err != nil {
		switch {
		case errors.Is(err, repository.ErrFollowRequestNotFound):
			sendJSONError(w, "Follow request not found", http.StatusNotFound)
			return
		case errors.Is(err, repository.ErrFollowRequestForbidden):
			sendJSONError(w, "You cannot modify this follow request", http.StatusForbidden)
			return
		case errors.Is(err, repository.ErrFollowRequestNotPending):
			sendJSONError(w, "Follow request is already processed", http.StatusConflict)
			return
		default:
			log.Printf("Error resolving follow request: %v", err)
			sendJSONError(w, "Could not update follow request", http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"id":     requestID,
		"status": status,
	})

	if status == "accepted" {
		followRequest, err := h.UserRepo.GetFollowRequestByID(requestID)
		if err != nil {
			log.Printf("Error loading follow request after accept for notifications: %v", err)
			return
		}

		h.createNotification(
			followRequest.SenderID,
			receiverID,
			"follow_request_accepted",
			followRequest.ID,
		)

		// Emit follower_added via WebSocket to both users
		if h.Notifier != nil {
			payload := wshandler.OutgoingEnvelope{
				Type: wshandler.TypeFollowerAdded,
				Payload: map[string]string{
					"followerId":  followRequest.SenderID,
					"followingId": receiverID,
				},
			}
			h.Notifier.EmitToUser(followRequest.SenderID, payload)
			h.Notifier.EmitToUser(receiverID, payload)
		}
	}
}

func (h *HandlerStruct) DeleteFollowerHandler(w http.ResponseWriter, r *http.Request) {
	if h.UserRepo == nil {
		sendJSONError(w, "User repository is not configured", http.StatusInternalServerError)
		return
	}

	profileOwnerID, ok := userIDFromRequest(r)
	if !ok {
		sendJSONError(w, "Not authenticated", http.StatusUnauthorized)
		return
	}

	followerUserID := strings.TrimSpace(r.PathValue("userId"))
	if followerUserID == "" {
		sendJSONError(w, "userId is required", http.StatusBadRequest)
		return
	}
	if followerUserID == profileOwnerID {
		sendJSONError(w, "You cannot remove yourself as a follower", http.StatusBadRequest)
		return
	}

	if _, err := h.UserRepo.GetUserByID(followerUserID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			sendJSONError(w, "User not found", http.StatusNotFound)
			return
		}
		log.Printf("Error loading follower user: %v", err)
		sendJSONError(w, "Could not remove follower", http.StatusInternalServerError)
		return
	}

	removed, err := h.UserRepo.RemoveFollower(profileOwnerID, followerUserID)
	if err != nil {
		log.Printf("Error removing follower: %v", err)
		sendJSONError(w, "Could not remove follower", http.StatusInternalServerError)
		return
	}
	if !removed {
		sendJSONError(w, "Follower relationship not found", http.StatusNotFound)
		return
	}

	// Clean up post_allowed_users for the removed follower
	if h.PostRepo != nil {
		if err := h.PostRepo.RemoveUserFromAllowedPosts(profileOwnerID, followerUserID); err != nil {
			log.Printf("Error cleaning up allowed posts after follower removal: %v", err)
		}
	}

	// Emit follower_removed via WebSocket to both users
	if h.Notifier != nil {
		payload := wshandler.OutgoingEnvelope{
			Type: wshandler.TypeFollowerRemoved,
			Payload: map[string]string{
				"followerId":  followerUserID,
				"followingId": profileOwnerID,
			},
		}
		h.Notifier.EmitToUser(followerUserID, payload)
		h.Notifier.EmitToUser(profileOwnerID, payload)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"message": "Follower removed successfully"})
}

func (h *HandlerStruct) DeleteFollowingHandler(w http.ResponseWriter, r *http.Request) {
	if h.UserRepo == nil {
		sendJSONError(w, "User repository is not configured", http.StatusInternalServerError)
		return
	}

	followerID, ok := userIDFromRequest(r)
	if !ok {
		sendJSONError(w, "Not authenticated", http.StatusUnauthorized)
		return
	}

	followingID := strings.TrimSpace(r.PathValue("userId"))
	if followingID == "" {
		sendJSONError(w, "userId is required", http.StatusBadRequest)
		return
	}
	if followingID == followerID {
		sendJSONError(w, "You cannot unfollow yourself", http.StatusBadRequest)
		return
	}

	if _, err := h.UserRepo.GetUserByID(followingID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			sendJSONError(w, "User not found", http.StatusNotFound)
			return
		}
		log.Printf("Error loading following user: %v", err)
		sendJSONError(w, "Could not unfollow user", http.StatusInternalServerError)
		return
	}

	removed, err := h.UserRepo.RemoveFollowing(followerID, followingID)
	if err != nil {
		log.Printf("Error unfollowing user: %v", err)
		sendJSONError(w, "Could not unfollow user", http.StatusInternalServerError)
		return
	}
	if !removed {
		sendJSONError(w, "You are not following this user", http.StatusNotFound)
		return
	}

	// Clean up post_allowed_users for the unfollowed user's private posts
	if h.PostRepo != nil {
		if err := h.PostRepo.RemoveUserFromAllowedPosts(followingID, followerID); err != nil {
			log.Printf("Error cleaning up allowed posts after unfollow: %v", err)
		}
	}

	// Emit follower_removed via WebSocket to both users
	if h.Notifier != nil {
		payload := wshandler.OutgoingEnvelope{
			Type: wshandler.TypeFollowerRemoved,
			Payload: map[string]string{
				"followerId":  followerID,
				"followingId": followingID,
			},
		}
		h.Notifier.EmitToUser(followerID, payload)
		h.Notifier.EmitToUser(followingID, payload)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"message": "Unfollowed successfully"})
}

func (h *HandlerStruct) DeleteFollowRequestHandler(w http.ResponseWriter, r *http.Request) {
	if h.UserRepo == nil {
		sendJSONError(w, "User repository is not configured", http.StatusInternalServerError)
		return
	}

	senderID, ok := userIDFromRequest(r)
	if !ok {
		sendJSONError(w, "Not authenticated", http.StatusUnauthorized)
		return
	}

	requestID := strings.TrimSpace(r.PathValue("requestId"))
	if requestID == "" {
		sendJSONError(w, "requestId is required", http.StatusBadRequest)
		return
	}

	if err := h.UserRepo.CancelOutgoingFollowRequest(requestID, senderID); err != nil {
		switch {
		case errors.Is(err, repository.ErrFollowRequestNotFound):
			sendJSONError(w, "Follow request not found", http.StatusNotFound)
			return
		case errors.Is(err, repository.ErrFollowRequestForbidden):
			sendJSONError(w, "You cannot cancel this follow request", http.StatusForbidden)
			return
		case errors.Is(err, repository.ErrFollowRequestNotPending):
			sendJSONError(w, "Only pending follow requests can be withdrawn", http.StatusConflict)
			return
		default:
			log.Printf("Error cancelling follow request: %v", err)
			sendJSONError(w, "Could not cancel follow request", http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"id":     requestID,
		"status": "cancelled",
	})
}
