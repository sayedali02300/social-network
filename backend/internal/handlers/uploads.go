package handlers

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func (h *HandlerStruct) ServeUploadsHandler(w http.ResponseWriter, r *http.Request) {
	uploadType := r.PathValue("type")
	filename := r.PathValue("filename")

	if uploadType == "" || filename == "" {
		http.Error(w, "Invalid upload path", http.StatusBadRequest)
		return
	}

	// Calculate logical path stored in DB
	imagePath := "/uploads/" + uploadType + "/" + filename

	// Avatars are always public
	if uploadType == "avatars" {
		serveFile(w, r, imagePath)
		return
	}

	// For posts and comments, check authentication
	currentUserID, ok := userIDFromRequest(r)
	if !ok {
		http.Error(w, "Unauthorized access to image", http.StatusUnauthorized)
		return
	}

	if uploadType == "posts" {
		post, err := h.PostRepo.GetPostByImagePath(imagePath, currentUserID)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		if post == nil {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		serveFile(w, r, imagePath)
		return
	}

	if uploadType == "comments" {
		postID, err := h.CommentRepo.GetPostIDByImagePath(imagePath)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		if postID == "" {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}

		post, err := h.PostRepo.GetSinglePost(postID, currentUserID)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		if post == nil {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		serveFile(w, r, imagePath)
		return
	}

	http.Error(w, "Not Found", http.StatusNotFound)
}

func serveFile(w http.ResponseWriter, r *http.Request, imagePath string) {
	localFilePath := strings.TrimPrefix(imagePath, "/")
	localFilePath = filepath.Clean(localFilePath)

	if _, err := os.Stat(localFilePath); os.IsNotExist(err) {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	http.ServeFile(w, r, localFilePath)
}
