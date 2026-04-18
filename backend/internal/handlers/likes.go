package handlers

import (
	"encoding/json"
	"net/http"
)

func (h *HandlerStruct) LikePostHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromRequest(r)
	if !ok {
		sendJSONError(w, "Not authenticated", http.StatusUnauthorized)
		return
	}
	postID := r.PathValue("postId")
	if postID == "" {
		sendJSONError(w, "postId is required", http.StatusBadRequest)
		return
	}

	var body struct {
		Value int `json:"value"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || (body.Value != 1 && body.Value != -1) {
		sendJSONError(w, "value must be 1 or -1", http.StatusBadRequest)
		return
	}

	// Verify post access
	post, err := h.PostRepo.GetSinglePost(postID, userID)
	if err != nil {
		sendJSONError(w, "internal server error", http.StatusInternalServerError)
		return
	}
	if post == nil {
		sendJSONError(w, "post not found or access denied", http.StatusNotFound)
		return
	}

	if err := h.LikeRepo.UpsertLike(postID, userID, body.Value); err != nil {
		sendJSONError(w, "Could not save reaction", http.StatusInternalServerError)
		return
	}

	likes, dislikes, err := h.LikeRepo.GetLikeCounts(postID)
	if err != nil {
		sendJSONError(w, "Could not fetch counts", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"postId":     postID,
		"likes":      likes,
		"dislikes":   dislikes,
		"myReaction": body.Value,
	})
}

func (h *HandlerStruct) UnlikePostHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromRequest(r)
	if !ok {
		sendJSONError(w, "Not authenticated", http.StatusUnauthorized)
		return
	}
	postID := r.PathValue("postId")
	if postID == "" {
		sendJSONError(w, "postId is required", http.StatusBadRequest)
		return
	}

	if err := h.LikeRepo.DeleteLike(postID, userID); err != nil {
		sendJSONError(w, "Could not remove reaction", http.StatusInternalServerError)
		return
	}

	likes, dislikes, err := h.LikeRepo.GetLikeCounts(postID)
	if err != nil {
		sendJSONError(w, "Could not fetch counts", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"postId":     postID,
		"likes":      likes,
		"dislikes":   dislikes,
		"myReaction": 0,
	})
}
