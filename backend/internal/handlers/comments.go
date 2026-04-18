package handlers

import (
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"socialnetwork/backend/internal/models"
	wshandler "socialnetwork/backend/internal/ws"
	"strings"
	"time"

	"github.com/google/uuid"
)

func (h *HandlerStruct) GetCommentHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromRequest(r)
	if !ok {
		sendJSONError(w, "Not authenticated", http.StatusUnauthorized)
		return
	}

	postID := r.PathValue("postId")
	if postID == "" {
		sendJSONError(w, "postID is required", http.StatusBadRequest)
		return
	}

	post, err := h.PostRepo.GetSinglePost(postID, userID)
	if err != nil {
		log.Printf("GetSinglePost error during comments access: %v", err)
		sendJSONError(w, "internal server error", http.StatusInternalServerError)
		return
	}
	if post == nil {
		sendJSONError(w, "post not found or access denied", http.StatusNotFound)
		return
	}

	comments, err := h.CommentRepo.GetAllComments(postID)
	if err != nil {
		log.Printf("Repository Error: %v", err)
		sendJSONError(w, "Could not fetch comments from database", http.StatusInternalServerError)
		return
	}

	if comments == nil {
		comments = []models.CommentWithAuthor{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(comments)
}

func (h *HandlerStruct) PostCommentHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		sendJSONError(w, "Failed to parse form", http.StatusBadRequest)
		return
	}
	postID := r.PathValue("postId")
	if postID == "" {
		sendJSONError(w, "postID is required", http.StatusBadRequest)
		return
	}

	commentID := uuid.New().String()
	content := strings.TrimSpace(r.FormValue("content"))
	parentID := strings.TrimSpace(r.FormValue("parent_id"))
	_, _, hasImage := r.FormFile("image")
	if content == "" && hasImage != nil {
		sendJSONError(w, "Comment is required", http.StatusBadRequest)
		return
	}

	if len(content) > 0 && (len(content) < 3 || len(content) > 2500) {
		sendJSONError(w, "Comment should be between 3 to 2500 characters", http.StatusBadRequest)
		return
	}
	userID, ok := userIDFromRequest(r)
	if !ok {
		sendJSONError(w, "Not authenticated", http.StatusUnauthorized)
		return
	}

	// Verify user has access to this post
	post, err := h.PostRepo.GetSinglePost(postID, userID)
	if err != nil {
		log.Printf("GetSinglePost error during comment creation: %v", err)
		sendJSONError(w, "internal server error", http.StatusInternalServerError)
		return
	}
	if post == nil {
		sendJSONError(w, "post not found or access denied", http.StatusNotFound)
		return
	}

	var imagePath string
	file, fileHeader, err := r.FormFile("image")
	if err == nil {
		defer file.Close()

		buffer := make([]byte, 512)
		_, err := file.Read(buffer)
		if err != nil && err != io.EOF {
			sendJSONError(w, "Error reading file", http.StatusInternalServerError)
			return
		}

		if _, err := file.Seek(0, io.SeekStart); err != nil {
			sendJSONError(w, "Error processing file", http.StatusInternalServerError)
			return
		}

		SecureContentType := http.DetectContentType(buffer)

		exts := map[string]string{
			"image/jpeg": ".jpg",
			"image/png":  ".png",
			"image/gif":  ".gif",
		}
		ext, ok := exts[SecureContentType]
		if !ok {
			log.Printf("Invalid file type detected: %s (Client claimed: %s)", SecureContentType, fileHeader.Header.Get("Content-Type"))
			sendJSONError(w, "Invalid File Type [only .jpg, .png, .gif are allowed]", http.StatusBadRequest)
			return
		}

		fileName := uuid.New().String() + ext
		savePath := filepath.Join("uploads", "comments", fileName)

		// Create directories if they do not exist
		if err := os.MkdirAll(filepath.Dir(savePath), os.ModePerm); err != nil {
			sendJSONError(w, "Error saving file", http.StatusInternalServerError)
			return
		}

		// Save the file
		dst, err := os.Create(savePath)
		if err != nil {
			sendJSONError(w, "Error saving file", http.StatusInternalServerError)
			return
		}
		defer dst.Close()

		if _, err := io.Copy(dst, file); err != nil {
			sendJSONError(w, "Error saving file", http.StatusInternalServerError)
			return
		}

		// Prepend with / to make it an absolute path for the frontend
		imagePath = "/" + filepath.ToSlash(savePath)
	} else if err != http.ErrMissingFile {
		sendJSONError(w, "Error retrieving file", http.StatusInternalServerError)
		return
	}

	// Validate parent comment belongs to the same post
	if parentID != "" {
		parentAuthorID, parentPostID, err := h.CommentRepo.GetCommentAuthorAndPostID(parentID)
		if err != nil || parentPostID != postID {
			sendJSONError(w, "Invalid parent comment", http.StatusBadRequest)
			return
		}
		// Notify parent comment author if different from commenter
		if parentAuthorID != userID {
			notifPayload := `{"postId":"` + postID + `","commentId":"` + commentID + `"}`
			h.createNotification(parentAuthorID, userID, "new_comment_reply", notifPayload)
		}
	}

	newComment := models.Comment{
		ID:        commentID,
		PostID:    postID,
		UserID:    userID,
		Content:   content,
		ImagePath: imagePath,
		ParentID:  parentID,
		CreatedAt: time.Now().UTC(),
	}

	if err := h.CommentRepo.CreateComment(&newComment); err != nil {
		log.Printf("Database Error: %v", err)
		if imagePath != "" {
			deleteImageFile(imagePath)
		}
		sendJSONError(w, "Could not save comment", http.StatusInternalServerError)
		return
	}

	// Notify the post owner (skip if commenter is the post owner)
	if post.UserID != userID {
		notifPayload := `{"postId":"` + postID + `","commentId":"` + commentID + `"}`
		h.createNotification(post.UserID, userID, "new_comment", notifPayload)
	}

	// START WEBSOCKET
	// fixing early return on WS recipient error preventing HTTP response from being sent
	if strings.TrimSpace(post.GroupID) != "" {
		h.emitGroupCommentRealtime(post.GroupID, commentID)
	} else if h.NotifierMany != nil {
		var recipientIDs []string
		var recipientErr error

		switch post.Privacy {
		case "public":
			recipientIDs, recipientErr = h.PostRepo.GetFeedUserIDs(post.UserID)
		case "almost_private":
			recipientIDs, recipientErr = h.PostRepo.GetFollowersIDs(post.UserID)
		case "private":
			recipientIDs, recipientErr = h.PostRepo.GetAllowedUsers(post.ID)
		}

		if recipientErr != nil {
			log.Println("ws: failed to get comment recipients:", recipientErr)
		} else {
			recipientIDs = append(recipientIDs, post.UserID)

			commentWithAuthor, err := h.CommentRepo.GetCommentByID(commentID)
			if err != nil {
				log.Printf("failed to fetch comment with author: %v", err)
				commentWithAuthor = &models.CommentWithAuthor{
					ID:        newComment.ID,
					PostID:    newComment.PostID,
					Content:   newComment.Content,
					ImagePath: newComment.ImagePath,
					CreatedAt: newComment.CreatedAt,
					Author: models.CommentAuthor{
						UserID: userID,
					},
				}
			}

			h.NotifierMany.EmitToUsers(recipientIDs, wshandler.OutgoingEnvelope{
				Type:    "comment_event",
				Payload: commentWithAuthor,
			})
		}
	}
	//END WEBSOCKET
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newComment)
}

func (h *HandlerStruct) DeleteCommentHandler(w http.ResponseWriter, r *http.Request) {
	requesterID, ok := userIDFromRequest(r)
	if !ok {
		sendJSONError(w, "Not authenticated", http.StatusUnauthorized)
		return
	}
	CommentID := r.PathValue("commentID")
	if CommentID == "" {
		sendJSONError(w, "Comment ID is required", http.StatusBadRequest)
		return
	}
	ownerID, err := h.CommentRepo.GetCommentOwnerID(CommentID)
	if err != nil {
		log.Printf("Error getting comment owner: %v", err)
		sendJSONError(w, "Comment not found", http.StatusNotFound)
		return
	}
	if ownerID != requesterID {
		sendJSONError(w, "You are not allowed to delete this comment", http.StatusForbidden)
		return
	}

	imagePath, err := h.CommentRepo.GetCommentImagePath(CommentID)
	if err != nil {
		log.Printf("Error getting image path: %v", err)
		sendJSONError(w, "Comment not found", http.StatusNotFound)
		return
	}

	postID, err := h.CommentRepo.GetPostIDByCommentID(CommentID)
	if err != nil {
		log.Printf("Error getting post ID for comment: %v", err)
		sendJSONError(w, "Could not find associated post", http.StatusInternalServerError)
		return
	}
	err = h.CommentRepo.DeleteComment(CommentID)
	if err != nil {
		log.Printf("Error deleting comment from DB: %v", err)
		sendJSONError(w, "Could not delete comment", http.StatusInternalServerError)
		return
	}

	if imagePath != "" {
		deleteImageFile(imagePath)
	}
	// START WEBSOCKET
	if postID != "" {
		// Need to get recipients for this post
		// Optimization: we could fetch the post to get its privacy/author, or use existing methods
		post, err := h.PostRepo.GetSinglePost(postID, requesterID)
		if err == nil && post != nil {
			if strings.TrimSpace(post.GroupID) != "" {
				h.emitGroupCommentDeleteRealtime(post.GroupID, postID, CommentID)
			} else if h.NotifierMany != nil {
				var recipientIDs []string
				switch post.Privacy {
				case "public":
					recipientIDs, _ = h.PostRepo.GetFeedUserIDs(post.UserID)
					recipientIDs = append(recipientIDs, post.UserID)
				case "almost_private":
					recipientIDs, _ = h.PostRepo.GetFollowersIDs(post.UserID)
					recipientIDs = append(recipientIDs, post.UserID)
				case "private":
					recipientIDs, _ = h.PostRepo.GetAllowedUsers(post.ID)
					recipientIDs = append(recipientIDs, post.UserID)
				}
				// fixing inconsistent key name: "commentid" should be camelCase "commentId"
				h.NotifierMany.EmitToUsers(recipientIDs, wshandler.OutgoingEnvelope{
					Type: "delete_comment",
					Payload: map[string]string{
						"commentId": CommentID,
						"postId":    postID,
					},
				})
			}
		}
	}
	// END WEBSOCKET
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Comment deleted successfully"})
}

func (h *HandlerStruct) PatchCommentHandler(w http.ResponseWriter, r *http.Request) {
	requesterID, ok := userIDFromRequest(r)
	if !ok {
		sendJSONError(w, "Not authenticated", http.StatusUnauthorized)
		return
	}

	commentID := strings.TrimSpace(r.PathValue("commentID"))
	if commentID == "" {
		sendJSONError(w, "Comment ID is required", http.StatusBadRequest)
		return
	}

	ownerID, err := h.CommentRepo.GetCommentOwnerID(commentID)
	if err != nil {
		log.Printf("Error getting comment owner: %v", err)
		sendJSONError(w, "Comment not found", http.StatusNotFound)
		return
	}
	if ownerID != requesterID {
		sendJSONError(w, "You are not allowed to edit this comment", http.StatusForbidden)
		return
	}

	var payload struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		sendJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	content := strings.TrimSpace(payload.Content)
	if len(content) < 3 || len(content) > 2500 {
		sendJSONError(w, "Comment should be between 3 to 2500 characters", http.StatusBadRequest)
		return
	}

	if err := h.CommentRepo.UpdateComment(commentID, content); err != nil {
		if err == sql.ErrNoRows {
			sendJSONError(w, "Comment not found", http.StatusNotFound)
			return
		}
		log.Printf("Error updating comment: %v", err)
		sendJSONError(w, "Could not update comment", http.StatusInternalServerError)
		return
	}

	updatedComment, err := h.CommentRepo.GetCommentByID(commentID)
	if err != nil {
		log.Printf("Error loading updated comment: %v", err)
		sendJSONError(w, "Comment updated, but failed to reload it", http.StatusInternalServerError)
		return
	}
	if updatedComment == nil {
		sendJSONError(w, "Comment updated, but no longer accessible", http.StatusNotFound)
		return
	}

	postID, err := h.CommentRepo.GetPostIDByCommentID(commentID)
	if err == nil && strings.TrimSpace(postID) != "" {
		post, postErr := h.PostRepo.GetSinglePost(postID, requesterID)
		if postErr == nil && isGroupPost(post) {
			h.emitGroupCommentRealtime(post.GroupID, commentID)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedComment)
}
