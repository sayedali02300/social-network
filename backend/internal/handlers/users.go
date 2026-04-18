package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"mime"
	"net/http"
	"socialnetwork/backend/internal/models"
	"socialnetwork/backend/internal/repository"
	"strings"
)

type patchMeRequest struct {
	FirstName   *string `json:"firstName"`
	LastName    *string `json:"lastName"`
	DateOfBirth *string `json:"dateOfBirth"`
	Avatar      *string `json:"avatar"`
	Nickname    *string `json:"nickname"`
	AboutMe     *string `json:"aboutMe"`
}

type patchMePrivacyRequest struct {
	IsPublic *bool `json:"isPublic"`
}

func (h *HandlerStruct) canAccessProfile(viewerID string, targetUser *models.User) (bool, error) {
	if targetUser.ID == viewerID || targetUser.IsPublic {
		return true, nil
	}

	return h.UserRepo.IsFollowing(viewerID, targetUser.ID)
}

func (h *HandlerStruct) GetMeHandler(w http.ResponseWriter, r *http.Request) {
	if h.UserRepo == nil {
		sendJSONError(w, "User repository is not configured", http.StatusInternalServerError)
		return
	}

	userID, ok := userIDFromRequest(r)
	if !ok {
		sendJSONError(w, "Not authenticated", http.StatusUnauthorized)
		return
	}

	user, err := h.UserRepo.GetUserByID(userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			sendJSONError(w, "User not found", http.StatusNotFound)
			return
		}
		log.Printf("Error loading current user: %v", err)
		sendJSONError(w, "Could not load current user", http.StatusInternalServerError)
		return
	}

	user.Password = ""
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(user)
}

func (h *HandlerStruct) PatchMeHandler(w http.ResponseWriter, r *http.Request) {
	if h.UserRepo == nil {
		sendJSONError(w, "User repository is not configured", http.StatusInternalServerError)
		return
	}

	userID, ok := userIDFromRequest(r)
	if !ok {
		sendJSONError(w, "Not authenticated", http.StatusUnauthorized)
		return
	}

	currentUser, err := h.UserRepo.GetUserByID(userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			sendJSONError(w, "User not found", http.StatusNotFound)
			return
		}
		log.Printf("Error loading current user before update: %v", err)
		sendJSONError(w, "Could not update profile", http.StatusInternalServerError)
		return
	}

	var req patchMeRequest
	var uploadedAvatarPath string

	contentType := strings.TrimSpace(r.Header.Get("Content-Type"))
	mediaType, _, _ := mime.ParseMediaType(contentType)

	if strings.EqualFold(mediaType, "multipart/form-data") {
		if err := r.ParseMultipartForm(8 << 20); err != nil {
			sendJSONError(w, "Invalid multipart form data", http.StatusBadRequest)
			return
		}

		if values, ok := r.MultipartForm.Value["firstName"]; ok && len(values) > 0 {
			value := values[0]
			req.FirstName = &value
		}
		if values, ok := r.MultipartForm.Value["lastName"]; ok && len(values) > 0 {
			value := values[0]
			req.LastName = &value
		}
		if values, ok := r.MultipartForm.Value["dateOfBirth"]; ok && len(values) > 0 {
			value := values[0]
			req.DateOfBirth = &value
		}
		if values, ok := r.MultipartForm.Value["nickname"]; ok && len(values) > 0 {
			value := values[0]
			req.Nickname = &value
		}
		if values, ok := r.MultipartForm.Value["aboutMe"]; ok && len(values) > 0 {
			value := values[0]
			req.AboutMe = &value
		}
		if values, ok := r.MultipartForm.Value["avatar"]; ok && len(values) > 0 {
			// Allows clearing avatar with an empty string from form field.
			value := values[0]
			req.Avatar = &value
		}

		file, fileHeader, err := r.FormFile("avatarFile")
		if err == nil {
			defer file.Close()

			savedPath, saveErr := saveAvatarFile(file, fileHeader)
			if saveErr != nil {
				sendJSONError(w, saveErr.Error(), http.StatusBadRequest)
				return
			}
			uploadedAvatarPath = savedPath
			req.Avatar = &savedPath
		} else if !errors.Is(err, http.ErrMissingFile) {
			sendJSONError(w, "Invalid avatar file", http.StatusBadRequest)
			return
		}
	} else {
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&req); err != nil {
			sendJSONError(w, "Invalid request body", http.StatusBadRequest)
			return
		}
	}

	if req.FirstName != nil {
		trimmed := strings.TrimSpace(*req.FirstName)
		if err := validateRequiredLettersOnlyName("firstName", trimmed); err != nil {
			sendJSONError(w, err.Error(), http.StatusBadRequest)
			return
		}
		req.FirstName = &trimmed
	}
	if req.LastName != nil {
		trimmed := strings.TrimSpace(*req.LastName)
		if err := validateRequiredLettersOnlyName("lastName", trimmed); err != nil {
			sendJSONError(w, err.Error(), http.StatusBadRequest)
			return
		}
		req.LastName = &trimmed
	}
	if req.DateOfBirth != nil {
		trimmed := strings.TrimSpace(*req.DateOfBirth)
		if trimmed == "" {
			sendJSONError(w, "dateOfBirth cannot be empty", http.StatusBadRequest)
			return
		}
		if err := validateRealisticDateOfBirth(trimmed); err != nil {
			sendJSONError(w, err.Error(), http.StatusBadRequest)
			return
		}
		req.DateOfBirth = &trimmed
	}
	if req.Avatar != nil {
		trimmed := strings.TrimSpace(*req.Avatar)
		req.Avatar = &trimmed
	}
	if req.Nickname != nil {
		trimmed := strings.TrimSpace(*req.Nickname)
		if err := validateOptionalNickname(trimmed); err != nil {
			sendJSONError(w, err.Error(), http.StatusBadRequest)
			return
		}

		if trimmed != "" {
			nicknameExists, err := h.UserRepo.NicknameExistsForOtherUser(trimmed, userID)
			if err != nil {
				if uploadedAvatarPath != "" {
					removeUploadedFile(uploadedAvatarPath)
				}
				log.Printf("Error checking nickname availability: %v", err)
				sendJSONError(w, "Could not update profile", http.StatusInternalServerError)
				return
			}
			if nicknameExists {
				if uploadedAvatarPath != "" {
					removeUploadedFile(uploadedAvatarPath)
				}
				sendJSONError(w, "Nickname is already taken", http.StatusConflict)
				return
			}
		}

		req.Nickname = &trimmed
	}
	if req.AboutMe != nil {
		trimmed := strings.TrimSpace(*req.AboutMe)
		if len(trimmed) > 255 {
			sendJSONError(w, "aboutMe must be 255 characters or fewer", http.StatusBadRequest)
			return
		}
		req.AboutMe = &trimmed
	}

	if req.FirstName == nil && req.LastName == nil && req.DateOfBirth == nil && req.Avatar == nil && req.Nickname == nil && req.AboutMe == nil {
		sendJSONError(w, "At least one field must be provided", http.StatusBadRequest)
		return
	}

	updates := repository.UserProfileUpdates{
		FirstName:   req.FirstName,
		LastName:    req.LastName,
		DateOfBirth: req.DateOfBirth,
		Avatar:      req.Avatar,
		Nickname:    req.Nickname,
		AboutMe:     req.AboutMe,
	}

	if err := h.UserRepo.UpdateUserProfile(userID, updates); err != nil {
		if uploadedAvatarPath != "" {
			removeUploadedFile(uploadedAvatarPath)
		}
		if errors.Is(err, sql.ErrNoRows) {
			sendJSONError(w, "User not found", http.StatusNotFound)
			return
		}
		log.Printf("Error updating current user: %v", err)
		sendJSONError(w, "Could not update profile", http.StatusInternalServerError)
		return
	}

	user, err := h.UserRepo.GetUserByID(userID)
	if err != nil {
		if uploadedAvatarPath != "" {
			removeUploadedFile(uploadedAvatarPath)
		}
		if errors.Is(err, sql.ErrNoRows) {
			sendJSONError(w, "User not found", http.StatusNotFound)
			return
		}
		log.Printf("Error loading updated current user: %v", err)
		sendJSONError(w, "Could not load updated profile", http.StatusInternalServerError)
		return
	}

	if uploadedAvatarPath != "" && shouldDeleteOldAvatar(currentUser.Avatar, user.Avatar) {
		removeUploadedFile(currentUser.Avatar)
	}

	user.Password = ""
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(user)
}

func (h *HandlerStruct) PatchMePrivacyHandler(w http.ResponseWriter, r *http.Request) {
	if h.UserRepo == nil {
		sendJSONError(w, "User repository is not configured", http.StatusInternalServerError)
		return
	}

	userID, ok := userIDFromRequest(r)
	if !ok {
		sendJSONError(w, "Not authenticated", http.StatusUnauthorized)
		return
	}

	var req patchMePrivacyRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		sendJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.IsPublic == nil {
		sendJSONError(w, "isPublic is required", http.StatusBadRequest)
		return
	}

	if err := h.UserRepo.UpdateUserPrivacy(userID, *req.IsPublic); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			sendJSONError(w, "User not found", http.StatusNotFound)
			return
		}
		log.Printf("Error updating user privacy: %v", err)
		sendJSONError(w, "Could not update privacy", http.StatusInternalServerError)
		return
	}

	user, err := h.UserRepo.GetUserByID(userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			sendJSONError(w, "User not found", http.StatusNotFound)
			return
		}
		log.Printf("Error loading updated current user: %v", err)
		sendJSONError(w, "Could not load updated profile", http.StatusInternalServerError)
		return
	}

	user.Password = ""
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(user)
}

func (h *HandlerStruct) GetUserByIDHandler(w http.ResponseWriter, r *http.Request) {
	if h.UserRepo == nil {
		sendJSONError(w, "User repository is not configured", http.StatusInternalServerError)
		return
	}

	viewerID, ok := userIDFromRequest(r)
	if !ok {
		sendJSONError(w, "Not authenticated", http.StatusUnauthorized)
		return
	}

	userID := strings.TrimSpace(r.PathValue("userId"))
	if userID == "" {
		sendJSONError(w, "userId is required", http.StatusBadRequest)
		return
	}

	user, err := h.UserRepo.GetUserByID(userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			sendJSONError(w, "User not found", http.StatusNotFound)
			return
		}
		log.Printf("Error loading user by id: %v", err)
		sendJSONError(w, "Could not load user", http.StatusInternalServerError)
		return
	}

	canAccess, err := h.canAccessProfile(viewerID, user)
	if err != nil {
		log.Printf("Error checking profile visibility: %v", err)
		sendJSONError(w, "Could not enforce profile visibility", http.StatusInternalServerError)
		return
	}

	user.Password = ""
	if user.ID != viewerID && !canAccess {
		user.Email = ""
		user.DateOfBirth = ""
		// Private profile preview for non-followers: keep basic identity only.
		user.AboutMe = ""
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(user)
}

func (h *HandlerStruct) SearchUsersHandler(w http.ResponseWriter, r *http.Request) {
	if h.UserRepo == nil {
		sendJSONError(w, "User repository is not configured", http.StatusInternalServerError)
		return
	}

	viewerID, ok := userIDFromRequest(r)
	if !ok {
		sendJSONError(w, "Not authenticated", http.StatusUnauthorized)
		return
	}

	query := strings.TrimSpace(r.URL.Query().Get("q"))
	if query == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode([]models.User{})
		return
	}

	users, err := h.UserRepo.SearchUsers(query, viewerID, 20)
	if err != nil {
		log.Printf("Error searching users: %v", err)
		sendJSONError(w, "Could not search users", http.StatusInternalServerError)
		return
	}

	// Collect all other-user IDs for a single batch follow-status check.
	var otherIDs []string
	for i := range users {
		if users[i].ID != viewerID {
			otherIDs = append(otherIDs, users[i].ID)
		}
	}
	followStatus, err := h.UserRepo.FollowStatusAmong(viewerID, otherIDs)
	if err != nil {
		log.Printf("Error checking search result follow status: %v", err)
		sendJSONError(w, "Could not determine follow status", http.StatusInternalServerError)
		return
	}

	type searchResult struct {
		ID           string `json:"id"`
		Email        string `json:"email"`
		FirstName    string `json:"firstName"`
		LastName     string `json:"lastName"`
		Avatar       string `json:"avatar"`
		Nickname     string `json:"nickname"`
		IsPublic     bool   `json:"isPublic"`
		FollowStatus string `json:"followStatus"`
	}

	out := make([]searchResult, 0, len(users))
	for i := range users {
		status := followStatus[users[i].ID] // "following", "requested", or ""
		canAccess := users[i].ID == viewerID || users[i].IsPublic || status == "following"
		u := searchResult{
			ID:           users[i].ID,
			Avatar:       users[i].Avatar,
			IsPublic:     users[i].IsPublic,
			FollowStatus: status,
		}
		if canAccess {
			u.Email = users[i].Email
			u.FirstName = users[i].FirstName
			u.LastName = users[i].LastName
			u.Nickname = users[i].Nickname
		}
		out = append(out, u)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(out)
}

func shouldDeleteOldAvatar(oldAvatarPath, newAvatarPath string) bool {
	if oldAvatarPath == "" || oldAvatarPath == newAvatarPath {
		return false
	}
	return strings.HasPrefix(oldAvatarPath, "/uploads/avatars/")
}

func (h *HandlerStruct) GetFollowersHandler(w http.ResponseWriter, r *http.Request) {
	if h.UserRepo == nil {
		sendJSONError(w, "User repository is not configured", http.StatusInternalServerError)
		return
	}

	viewerID, ok := userIDFromRequest(r)
	if !ok {
		sendJSONError(w, "Not authenticated", http.StatusUnauthorized)
		return
	}

	userID := strings.TrimSpace(r.PathValue("userId"))
	if userID == "" {
		sendJSONError(w, "userId is required", http.StatusBadRequest)
		return
	}

	profileUser, err := h.UserRepo.GetUserByID(userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			sendJSONError(w, "User not found", http.StatusNotFound)
			return
		}
		log.Printf("Error loading user by id for followers: %v", err)
		sendJSONError(w, "Could not load followers", http.StatusInternalServerError)
		return
	}

	canAccess, err := h.canAccessProfile(viewerID, profileUser)
	if err != nil {
		log.Printf("Error checking followers visibility: %v", err)
		sendJSONError(w, "Could not enforce profile visibility", http.StatusInternalServerError)
		return
	}
	if !canAccess {
		sendJSONError(w, "This profile is private", http.StatusForbidden)
		return
	}

	followers, err := h.UserRepo.GetFollowersByUserID(userID)
	if err != nil {
		log.Printf("Error loading followers: %v", err)
		sendJSONError(w, "Could not load followers", http.StatusInternalServerError)
		return
	}
	if followers == nil {
		followers = []models.User{}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(followers)
}

func (h *HandlerStruct) GetFollowingHandler(w http.ResponseWriter, r *http.Request) {
	if h.UserRepo == nil {
		sendJSONError(w, "User repository is not configured", http.StatusInternalServerError)
		return
	}

	viewerID, ok := userIDFromRequest(r)
	if !ok {
		sendJSONError(w, "Not authenticated", http.StatusUnauthorized)
		return
	}

	userID := strings.TrimSpace(r.PathValue("userId"))
	if userID == "" {
		sendJSONError(w, "userId is required", http.StatusBadRequest)
		return
	}

	profileUser, err := h.UserRepo.GetUserByID(userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			sendJSONError(w, "User not found", http.StatusNotFound)
			return
		}
		log.Printf("Error loading user by id for following: %v", err)
		sendJSONError(w, "Could not load following", http.StatusInternalServerError)
		return
	}

	canAccess, err := h.canAccessProfile(viewerID, profileUser)
	if err != nil {
		log.Printf("Error checking following visibility: %v", err)
		sendJSONError(w, "Could not enforce profile visibility", http.StatusInternalServerError)
		return
	}
	if !canAccess {
		sendJSONError(w, "This profile is private", http.StatusForbidden)
		return
	}

	following, err := h.UserRepo.GetFollowingByUserID(userID)
	if err != nil {
		log.Printf("Error loading following: %v", err)
		sendJSONError(w, "Could not load following", http.StatusInternalServerError)
		return
	}
	if following == nil {
		following = []models.User{}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(following)
}
