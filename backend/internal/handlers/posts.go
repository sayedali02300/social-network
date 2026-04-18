package handlers

import (
	"database/sql"
	"encoding/json"
	"github.com/google/uuid"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"socialnetwork/backend/internal/models"
	wshandler "socialnetwork/backend/internal/ws"
	"strings"
	"time"
	"unicode/utf8"
)

func (h *HandlerStruct) GetPostsHandler(w http.ResponseWriter, r *http.Request) {
	viewerID, ok := userIDFromRequest(r)
	if !ok {
		sendJSONError(w, "Not authenticated", http.StatusUnauthorized)
		return
	}

	posts, err := h.PostRepo.GetAllPosts(viewerID)
	if err != nil {
		log.Printf("Repository Error: %v", err)
		sendJSONError(w, "Could not fetch posts from database", http.StatusInternalServerError)
		return
	}
	if posts == nil {
		posts = []models.Post{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(posts)
}

func (h *HandlerStruct) GetGroupPostsHandler(w http.ResponseWriter, r *http.Request) {
	viewerID, ok := userIDFromRequest(r)
	if !ok {
		sendJSONError(w, "Not authenticated", http.StatusUnauthorized)
		return
	}

	groupID := strings.TrimSpace(r.PathValue("groupId"))
	if groupID == "" {
		sendJSONError(w, "groupId is required", http.StatusBadRequest)
		return
	}

	posts, err := h.PostRepo.GetGroupPosts(groupID, viewerID)
	if err != nil {
		if err == sql.ErrNoRows {
			sendJSONError(w, "only group members can view group posts", http.StatusForbidden)
			return
		}
		log.Printf("Repository Error: %v", err)
		sendJSONError(w, "Could not fetch group posts", http.StatusInternalServerError)
		return
	}
	if posts == nil {
		posts = []models.Post{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(posts)
}

func (h *HandlerStruct) PostingHandler(w http.ResponseWriter, r *http.Request) {
	newPost, err := h.createPostFromMultipartRequest(r, "")
	if err != nil {
		sendRequestError(w, err)
		return
	}

	if err := h.PostRepo.CreatePost(newPost); err != nil {
		log.Printf("Database Error: %v", err)
		if newPost.ImagePath != "" {
			deleteImageFile(newPost.ImagePath)
		}
		sendJSONError(w, "Could not save post", http.StatusInternalServerError)
		return
	}

	if newPost.Privacy == "private" {
		if err := h.applyAllowedFollowersToPost(r, newPost.ID, newPost.UserID, newPost.ImagePath); err != nil {
			sendRequestError(w, err)
			return
		}
	}

	// START WEBSOCKET
	if h.NotifierMany != nil {
		var recipientIDs []string

		switch newPost.Privacy {
		case "public":
			recipientIDs, err = h.PostRepo.GetFeedUserIDs(newPost.UserID)
		case "almost_private":
			recipientIDs, err = h.PostRepo.GetFollowersIDs(newPost.UserID)
		case "private":
			recipientIDs, err = h.PostRepo.GetAllowedUsers(newPost.ID)
		}

		if err != nil {
			log.Println("ws: failed to get recipients:", err)
		} else {
			recipientIDs = append(recipientIDs, newPost.UserID)

			postWithAuthor, err := h.PostRepo.GetSinglePost(newPost.ID, newPost.UserID)
			if err != nil {
				log.Printf("ws: failed to fetch post with author: %v", err)
				postWithAuthor = newPost // fall back to the bare post
			}

			h.NotifierMany.EmitToUsers(recipientIDs, wshandler.OutgoingEnvelope{
				Type:    "post_event",
				Payload: postWithAuthor,
			})
		}
	}
	// END WEBSOCKET

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newPost)
}

func (h *HandlerStruct) CreateGroupPostHandler(w http.ResponseWriter, r *http.Request) {
	groupID := strings.TrimSpace(r.PathValue("groupId"))
	if groupID == "" {
		sendJSONError(w, "groupId is required", http.StatusBadRequest)
		return
	}

	newPost, err := h.createPostFromMultipartRequest(r, groupID)
	if err != nil {
		sendRequestError(w, err)
		return
	}

	if _, err := h.PostRepo.GetGroupPosts(groupID, newPost.UserID); err != nil {
		if err == sql.ErrNoRows {
			if newPost.ImagePath != "" {
				deleteImageFile(newPost.ImagePath)
			}
			sendJSONError(w, "only group members can create group posts", http.StatusForbidden)
			return
		}
		if newPost.ImagePath != "" {
			deleteImageFile(newPost.ImagePath)
		}
		sendJSONError(w, "Could not validate group membership", http.StatusInternalServerError)
		return
	}

	newPost.Privacy = "private"
	if err := h.PostRepo.CreatePost(newPost); err != nil {
		log.Printf("Database Error: %v", err)
		if newPost.ImagePath != "" {
			deleteImageFile(newPost.ImagePath)
		}
		sendJSONError(w, "Could not save group post", http.StatusInternalServerError)
		return
	}
	h.emitGroupPostRealtime(groupID, newPost.ID, newPost.UserID)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newPost)
}

func (h *HandlerStruct) createPostFromMultipartRequest(r *http.Request, groupID string) (*models.Post, error) {
	// 10 MB limit for file uploads
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		return nil, newRequestError("Unable to parse form", http.StatusBadRequest)
	}

	title := r.FormValue("title")
	body := r.FormValue("body")
	privacy := r.FormValue("privacy")
	postID := uuid.New().String()

	title, body, err = sanitizeAndValidatePostFields(title, body)
	if err != nil {
		return nil, err
	}
	// check if the privacy option is one of the thee only
	allowedPrivacy := map[string]bool{"public": true, "private": true, "almost_private": true}
	if !allowedPrivacy[privacy] {
		privacy = "public"
	}

	userID, ok := userIDFromRequest(r)
	if !ok {
		return nil, newRequestError("Not authenticated", http.StatusUnauthorized)
	}

	var imagePath string
	file, fileHeader, err := r.FormFile("image")
	if err == nil {
		defer file.Close()

		buffer := make([]byte, 512)
		_, err := file.Read(buffer)
		if err != nil && err != io.EOF {
			return nil, newRequestError("Error reading file", http.StatusInternalServerError)
		}

		if _, err := file.Seek(0, io.SeekStart); err != nil {
			return nil, newRequestError("Error processing file", http.StatusInternalServerError)
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
			return nil, newRequestError("Invalid File Type [only .jpg, .png, .gif are allowed]", http.StatusBadRequest)
		}

		fileName := uuid.New().String() + ext
		savePath := filepath.Join("uploads", "posts", fileName)

		// Create directories if they do not exist
		if err := os.MkdirAll(filepath.Dir(savePath), os.ModePerm); err != nil {
			return nil, newRequestError("Error saving file", http.StatusInternalServerError)
		}

		// Save the file
		dst, err := os.Create(savePath)
		if err != nil {
			return nil, newRequestError("Error saving file", http.StatusInternalServerError)
		}
		defer dst.Close()

		if _, err := io.Copy(dst, file); err != nil {
			return nil, newRequestError("Error saving file", http.StatusInternalServerError)
		}

		// Prepend with / to make it an absolute path for the frontend
		imagePath = "/" + filepath.ToSlash(savePath)
	} else if err != http.ErrMissingFile {
		// Error other than missing file
		return nil, newRequestError("Error retrieving file", http.StatusInternalServerError)
	}

	// Create Post object
	newPost := models.Post{
		ID:        postID,
		UserID:    userID,
		GroupID:   groupID,
		Title:     title,
		Content:   body,
		ImagePath: imagePath,
		Privacy:   privacy,
		CreatedAt: time.Now().UTC(),
	}

	return &newPost, nil
}

func sanitizeAndValidatePostFields(title, body string) (string, string, error) {
	title = strings.TrimSpace(title)
	body = strings.TrimSpace(body)


	if title == "" || body == "" {
		return "", "", newRequestError("Title and Body are required", http.StatusBadRequest)
	}

	if utf8.RuneCountInString(title) < 3 || utf8.RuneCountInString(title) > 60 {
		return "", "", newRequestError("title should be between 3 to 60 characters", http.StatusBadRequest)
	}

	if utf8.RuneCountInString(body) < 3 || utf8.RuneCountInString(body) > 5000 {
		return "", "", newRequestError("body should be between 3 to 5000 characters", http.StatusBadRequest)
	}

	return title, body, nil
}

func (h *HandlerStruct) applyAllowedFollowersToPost(r *http.Request, postID, userID, imagePath string) error {
	followersJSON := r.FormValue("allowed_followers")
	if len(followersJSON) == 0 {
		return nil
	}

	var selectedFollowers []string
	if err := json.Unmarshal([]byte(followersJSON), &selectedFollowers); err != nil {
		log.Printf("Error parsing allowed_followers JSON: %v", err)
		h.PostRepo.DeletePost(postID)
		if imagePath != "" {
			deleteImageFile(imagePath)
		}
		return newRequestError("Invalid allowed_followers format", http.StatusBadRequest)
	}

	if err := h.PostRepo.AllowFollowers(selectedFollowers, postID, userID); err != nil {
		log.Printf("Error setting allowed followers: %v", err)
		h.PostRepo.DeletePost(postID)
		if imagePath != "" {
			deleteImageFile(imagePath)
		}
		return newRequestError("Could not set allowed followers", http.StatusInternalServerError)
	}

	return nil
}

func (h *HandlerStruct) GetOnePostHandler(w http.ResponseWriter, r *http.Request) {
	postID := r.PathValue("postId")
	if postID == "" {
		sendJSONError(w, "missing post ID", http.StatusBadRequest)
		return
	}

	currentUserID, ok := userIDFromRequest(r)
	if !ok {
		sendJSONError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	post, err := h.PostRepo.GetSinglePost(postID, currentUserID)
	if err != nil {
		log.Printf("GetSinglePost error: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	if post == nil {
		http.Error(w, "post not found or access denied", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(post)
}

func (h *HandlerStruct) PatchPostHandler(w http.ResponseWriter, r *http.Request) {
	requesterID, ok := userIDFromRequest(r)
	if !ok {
		sendJSONError(w, "Not authenticated", http.StatusUnauthorized)
		return
	}

	postID := strings.TrimSpace(r.PathValue("postId"))
	if postID == "" {
		sendJSONError(w, "Post ID is required", http.StatusBadRequest)
		return
	}

	ownerID, err := h.PostRepo.GetPostOwnerID(postID)
	if err != nil {
		log.Printf("Error getting post owner: %v", err)
		sendJSONError(w, "Post not found", http.StatusNotFound)
		return
	}
	if ownerID != requesterID {
		sendJSONError(w, "You are not allowed to edit this post", http.StatusForbidden)
		return
	}

	var payload struct {
		Title   string `json:"title"`
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		sendJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	title, content, err := sanitizeAndValidatePostFields(payload.Title, payload.Content)
	if err != nil {
		sendRequestError(w, err)
		return
	}

	if err := h.PostRepo.UpdatePost(postID, title, content); err != nil {
		if err == sql.ErrNoRows {
			sendJSONError(w, "Post not found", http.StatusNotFound)
			return
		}
		log.Printf("Error updating post: %v", err)
		sendJSONError(w, "Could not update post", http.StatusInternalServerError)
		return
	}

	updatedPost, err := h.PostRepo.GetSinglePost(postID, requesterID)
	if err != nil {
		log.Printf("Error loading updated post: %v", err)
		sendJSONError(w, "Post updated, but failed to reload it", http.StatusInternalServerError)
		return
	}
	if updatedPost == nil {
		sendJSONError(w, "Post updated, but no longer accessible", http.StatusNotFound)
		return
	}
	if isGroupPost(updatedPost) {
		h.emitGroupPostRealtime(updatedPost.GroupID, updatedPost.ID, requesterID)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedPost)
}

func (h *HandlerStruct) DeletePostHandler(w http.ResponseWriter, r *http.Request) {
	requesterID, ok := userIDFromRequest(r)
	if !ok {
		sendJSONError(w, "Not authenticated", http.StatusUnauthorized)
		return
	}

	postID := r.PathValue("postId")
	if postID == "" {
		sendJSONError(w, "Post ID is required", http.StatusBadRequest)
		return
	}

	ownerID, err := h.PostRepo.GetPostOwnerID(postID)
	if err != nil {
		log.Printf("Error getting post owner: %v", err)
		sendJSONError(w, "Post not found", http.StatusNotFound)
		return
	}
	if ownerID != requesterID {
		sendJSONError(w, "You are not allowed to delete this post", http.StatusForbidden)
		return
	}

	imagePath, err := h.PostRepo.GetPostImagePath(postID)
	if err != nil {
		log.Printf("Error getting image path: %v", err)
		sendJSONError(w, "Post not found", http.StatusNotFound)
		return
	}

	// fixing delete post WS ignoring privacy: fetch post privacy before deletion so we can broadcast to correct recipients
	var postPrivacy string
	var postGroupID string
	postBeforeDelete, err := h.PostRepo.GetSinglePost(postID, requesterID)
	if err == nil && postBeforeDelete != nil {
		postPrivacy = postBeforeDelete.Privacy
		postGroupID = postBeforeDelete.GroupID
	}

	// Resolve recipients BEFORE deletion — post_allowed_users has ON DELETE CASCADE,
	// so GetAllowedUsers would return empty after DeletePost.
	var recipientIDs []string
	if strings.TrimSpace(postGroupID) == "" {
		switch postPrivacy {
		case "almost_private":
			recipientIDs, err = h.PostRepo.GetFollowersIDs(ownerID)
		case "private":
			recipientIDs, err = h.PostRepo.GetAllowedUsers(postID)
		default:
			recipientIDs, err = h.PostRepo.GetFeedUserIDs(ownerID)
		}
		if err != nil {
			log.Println("ws: failed to get recipients:", err)
		}
	}

	err = h.PostRepo.DeletePost(postID)
	if err != nil {
		log.Printf("Error deleting post from DB: %v", err)
		sendJSONError(w, "Could not delete post", http.StatusInternalServerError)
		return
	}

	if imagePath != "" {
		deleteImageFile(imagePath)
	}

	//START WEBSOCKET
	if strings.TrimSpace(postGroupID) != "" {
		h.emitGroupPostDeleteRealtime(postGroupID, postID)
	} else if h.NotifierMany != nil {
		recipientIDs = append(recipientIDs, ownerID)
		h.NotifierMany.EmitToUsers(recipientIDs, wshandler.OutgoingEnvelope{
			Type:    "delete_post",
			Payload: map[string]string{"postId": postID},
		})
	}
	//END WEBSOCKET

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Post deleted successfully"})
}

func deleteImageFile(imagePathFromDB string) {
	localFilePath := strings.TrimPrefix(imagePathFromDB, "/")

	err := os.Remove(localFilePath)
	if err != nil {
		log.Printf("Warning: Failed to delete image file %s: %v", localFilePath, err)
	} else {
		log.Printf("Successfully deleted image: %s", localFilePath)
	}
}

type requestError struct {
	message string
	code    int
}

func (e requestError) Error() string {
	return e.message
}

func newRequestError(message string, code int) error {
	return requestError{message: message, code: code}
}

func sendRequestError(w http.ResponseWriter, err error) {
	if reqErr, ok := err.(requestError); ok {
		sendJSONError(w, reqErr.message, reqErr.code)
		return
	}
	sendJSONError(w, err.Error(), http.StatusBadRequest)
}

func (h *HandlerStruct) validateRecipientManagedPost(postID, requesterID, ownerAction string) error {
	ownerID, err := h.PostRepo.GetPostOwnerID(postID)
	if err != nil {
		return newRequestError("Post not found", http.StatusNotFound)
	}
	if ownerID != requesterID {
		return newRequestError("Only the post owner can "+ownerAction, http.StatusForbidden)
	}

	post, err := h.PostRepo.GetSinglePost(postID, requesterID)
	if err != nil || post == nil {
		return newRequestError("Post not found", http.StatusNotFound)
	}
	if strings.TrimSpace(post.GroupID) != "" {
		return newRequestError("Group posts do not support recipient management", http.StatusBadRequest)
	}
	if post.Privacy != "private" {
		return newRequestError("Can only manage recipients on private posts", http.StatusBadRequest)
	}

	return nil
}

func (h *HandlerStruct) AddAllowedUsersHandler(w http.ResponseWriter, r *http.Request) {
	requesterID, ok := userIDFromRequest(r)
	if !ok {
		sendJSONError(w, "Not authenticated", http.StatusUnauthorized)
		return
	}

	postID := r.PathValue("postId")
	if postID == "" {
		sendJSONError(w, "Post ID is required", http.StatusBadRequest)
		return
	}

	if err := h.validateRecipientManagedPost(postID, requesterID, "add recipients"); err != nil {
		sendRequestError(w, err)
		return
	}

	var body struct {
		UserIDs []string `json:"user_ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || len(body.UserIDs) == 0 {
		sendJSONError(w, "user_ids is required", http.StatusBadRequest)
		return
	}

	added, err := h.PostRepo.AddAllowedUsers(postID, requesterID, body.UserIDs)
	if err != nil {
		log.Printf("Error adding allowed users: %v", err)
		sendJSONError(w, "Could not add recipients", http.StatusInternalServerError)
		return
	}

	if h.NotifierMany != nil && len(added) > 0 {
		postWithAuthor, err := h.PostRepo.GetSinglePost(postID, requesterID)
		if err == nil && postWithAuthor != nil {
			h.NotifierMany.EmitToUsers(added, wshandler.OutgoingEnvelope{
				Type:    "post_event",
				Payload: postWithAuthor,
			})
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Recipients added successfully",
		"added":   len(added),
	})
}

func (h *HandlerStruct) RemoveAllowedUsersHandler(w http.ResponseWriter, r *http.Request) {
	requesterID, ok := userIDFromRequest(r)
	if !ok {
		sendJSONError(w, "Not authenticated", http.StatusUnauthorized)
		return
	}

	postID := r.PathValue("postId")
	if postID == "" {
		sendJSONError(w, "Post ID is required", http.StatusBadRequest)
		return
	}

	if err := h.validateRecipientManagedPost(postID, requesterID, "remove recipients"); err != nil {
		sendRequestError(w, err)
		return
	}

	var body struct {
		UserIDs []string `json:"user_ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || len(body.UserIDs) == 0 {
		sendJSONError(w, "user_ids is required", http.StatusBadRequest)
		return
	}

	if h.NotifierMany != nil {
		h.NotifierMany.EmitToUsers(body.UserIDs, wshandler.OutgoingEnvelope{
			Type:    "delete_post",
			Payload: map[string]string{"postId": postID},
		})
	}

	removed, err := h.PostRepo.RemoveAllowedUsers(postID, body.UserIDs)
	if err != nil {
		log.Printf("Error removing allowed users: %v", err)
		sendJSONError(w, "Could not remove recipients", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Recipients removed successfully",
		"removed": removed,
	})
}

func (h *HandlerStruct) GetAllowedUsersHandler(w http.ResponseWriter, r *http.Request) {
	requesterID, ok := userIDFromRequest(r)
	if !ok {
		sendJSONError(w, "Not authenticated", http.StatusUnauthorized)
		return
	}

	postID := r.PathValue("postId")
	if postID == "" {
		sendJSONError(w, "Post ID is required", http.StatusBadRequest)
		return
	}

	if err := h.validateRecipientManagedPost(postID, requesterID, "view recipients"); err != nil {
		sendRequestError(w, err)
		return
	}

	userIDs, err := h.PostRepo.GetAllowedUsers(postID)
	if err != nil {
		log.Printf("Error getting allowed users: %v", err)
		sendJSONError(w, "Could not get recipients", http.StatusInternalServerError)
		return
	}
	if userIDs == nil {
		userIDs = []string{}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"user_ids": userIDs,
	})
}
