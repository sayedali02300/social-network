package handlers

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"log"
	"mime"
	"mime/multipart"
	"net"
	"net/http"
	"net/mail"
	"os"
	"path/filepath"
	"socialnetwork/backend/internal/models"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type registerRequest struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	FirstName   string `json:"firstName"`
	LastName    string `json:"lastName"`
	DateOfBirth string `json:"dateOfBirth"`
	Avatar      string `json:"avatar"`
	Nickname    string `json:"nickname"`
	AboutMe     string `json:"aboutMe"`
	IsPublic    *bool  `json:"isPublic"`
}

type loginRequest struct {
	Identifier string `json:"identifier"`
	Email      string `json:"email"`
	Password   string `json:"password"`
}

func (h *HandlerStruct) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if h.UserRepo == nil {
		sendJSONError(w, "User repository is not configured", http.StatusInternalServerError)
		return
	}

	var req registerRequest
	var uploadedAvatarPath string
	contentType := strings.TrimSpace(r.Header.Get("Content-Type"))
	mediaType, _, _ := mime.ParseMediaType(contentType)

	if strings.EqualFold(mediaType, "multipart/form-data") {
		if err := r.ParseMultipartForm(6 << 20); err != nil {
			sendJSONError(w, "Invalid multipart form data", http.StatusBadRequest)
			return
		}

		req.Email = r.FormValue("email")
		req.Password = r.FormValue("password")
		req.FirstName = r.FormValue("firstName")
		req.LastName = r.FormValue("lastName")
		req.DateOfBirth = r.FormValue("dateOfBirth")
		req.Nickname = r.FormValue("nickname")
		req.AboutMe = r.FormValue("aboutMe")

		if isPublicRaw := strings.TrimSpace(r.FormValue("isPublic")); isPublicRaw != "" {
			parsed, err := parseFormBool(isPublicRaw)
			if err != nil {
				sendJSONError(w, "isPublic must be a boolean value", http.StatusBadRequest)
				return
			}
			req.IsPublic = &parsed
		}

		file, fileHeader, err := r.FormFile("avatar")
		if err == nil {
			defer file.Close()

			savedPath, saveErr := saveAvatarFile(file, fileHeader)
			if saveErr != nil {
				sendJSONError(w, saveErr.Error(), http.StatusBadRequest)
				return
			}

			uploadedAvatarPath = savedPath
			req.Avatar = savedPath
		} else if !errors.Is(err, http.ErrMissingFile) {
			sendJSONError(w, "Invalid avatar file", http.StatusBadRequest)
			return
		}
	} else {
		decoder := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
		decoder.DisallowUnknownFields()

		if err := decoder.Decode(&req); err != nil {
			sendJSONError(w, "Invalid request body", http.StatusBadRequest)
			return
		}
	}

	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	req.Password = strings.TrimSpace(req.Password)
	req.FirstName = strings.TrimSpace(req.FirstName)
	req.LastName = strings.TrimSpace(req.LastName)
	req.DateOfBirth = strings.TrimSpace(req.DateOfBirth)
	req.Avatar = strings.TrimSpace(req.Avatar)
	req.Nickname = strings.TrimSpace(req.Nickname)
	req.AboutMe = strings.TrimSpace(req.AboutMe)

	if err := validateRequiredLettersOnlyName("firstName", req.FirstName); err != nil {
		sendJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := validateRequiredLettersOnlyName("lastName", req.LastName); err != nil {
		sendJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := validateOptionalNickname(req.Nickname); err != nil {
		sendJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	if len(req.AboutMe) > 255 {
		sendJSONError(w, "aboutMe must be 255 characters or fewer", http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Password == "" || req.FirstName == "" || req.LastName == "" || req.DateOfBirth == "" {
		sendJSONError(w, "email, password, firstName, lastName and dateOfBirth are required", http.StatusBadRequest)
		return
	}

	if !isValidEmail(req.Email) {
		sendJSONError(w, "Invalid email format", http.StatusBadRequest)
		return
	}

	if len(req.Password) < 8 {
		sendJSONError(w, "Password must be at least 8 characters", http.StatusBadRequest)
		return
	}

	if err := validateRealisticDateOfBirth(req.DateOfBirth); err != nil {
		sendJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	emailExists, err := h.UserRepo.EmailExists(req.Email)
	if err != nil {
		log.Printf("Error checking user email: %v", err)
		sendJSONError(w, "Could not process registration", http.StatusInternalServerError)
		return
	}
	if emailExists {
		sendJSONError(w, "Email is already registered", http.StatusConflict)
		return
	}

	if req.Nickname != "" {
		nicknameExists, err := h.UserRepo.NicknameExists(req.Nickname)
		if err != nil {
			log.Printf("Error checking user nickname: %v", err)
			sendJSONError(w, "Could not process registration", http.StatusInternalServerError)
			return
		}
		if nicknameExists {
			sendJSONError(w, "Nickname is already taken", http.StatusConflict)
			return
		}
	}

	passwordHash, err := hashPassword(req.Password)
	if err != nil {
		log.Printf("Error hashing password: %v", err)
		sendJSONError(w, "Could not process registration", http.StatusInternalServerError)
		return
	}

	user := models.User{
		ID:          uuid.NewString(),
		Email:       req.Email,
		Password:    passwordHash,
		FirstName:   req.FirstName,
		LastName:    req.LastName,
		DateOfBirth: req.DateOfBirth,
		Avatar:      req.Avatar,
		Nickname:    req.Nickname,
		AboutMe:     req.AboutMe,
		IsPublic:    true,
		CreatedAt:   time.Now().UTC(),
	}

	if req.IsPublic != nil {
		user.IsPublic = *req.IsPublic
	}

	if err := h.UserRepo.CreateUser(&user); err != nil {
		if uploadedAvatarPath != "" {
			removeUploadedFile(uploadedAvatarPath)
		}
		log.Printf("Error creating user: %v", err)
		sendJSONError(w, "Could not create user", http.StatusInternalServerError)
		return
	}

	user.Password = ""
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(user)
}

func (h *HandlerStruct) LoginHandler(w http.ResponseWriter, r *http.Request) {
	if h.UserRepo == nil || h.SessionStore == nil {
		sendJSONError(w, "Auth services are not configured", http.StatusInternalServerError)
		return
	}

	var req loginRequest
	decoder := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		sendJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	req.Identifier = strings.ToLower(strings.TrimSpace(req.Identifier))
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	req.Password = strings.TrimSpace(req.Password)

	identifier := req.Identifier
	if identifier == "" {
		identifier = req.Email
	}

	if identifier == "" || req.Password == "" {
		sendJSONError(w, "identifier and password are required", http.StatusBadRequest)
		return
	}

	user, err := h.UserRepo.GetUserByEmailOrNickname(identifier)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			sendJSONError(w, "Invalid email, nickname, or password", http.StatusUnauthorized)
			return
		}
		log.Printf("Error loading user by identifier: %v", err)
		sendJSONError(w, "Could not process login", http.StatusInternalServerError)
		return
	}

	if !verifyPassword(req.Password, user.Password) {
		sendJSONError(w, "Invalid email, nickname, or password", http.StatusUnauthorized)
		return
	}

	if err := h.SessionStore.DeleteSessionsByUserID(user.ID); err != nil {
		log.Printf("Error rotating user sessions: %v", err)
		sendJSONError(w, "Could not process login", http.StatusInternalServerError)
		return
	}

	sessionToken, err := generateSessionToken()
	if err != nil {
		log.Printf("Error generating session token: %v", err)
		sendJSONError(w, "Could not process login", http.StatusInternalServerError)
		return
	}

	session := models.Session{
		ID:        sessionToken,
		UserID:    user.ID,
		CreatedAt: time.Now().UTC(),
		ExpiresAt: time.Now().UTC().Add(24 * time.Hour),
		IPAddress: extractClientIP(r),
		UserAgent: r.Header.Get("User-Agent"),
	}
	if err := h.SessionStore.CreateSession(&session); err != nil {
		log.Printf("Error creating session: %v", err)
		sendJSONError(w, "Could not process login", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    session.ID,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   isSecureRequest(r),
		Expires:  session.ExpiresAt,
	})

	user.Password = ""
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(user)
}

func (h *HandlerStruct) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	if h.SessionStore == nil {
		sendJSONError(w, "Auth services are not configured", http.StatusInternalServerError)
		return
	}

	cookie, err := r.Cookie("session_id")
	if err != nil && !errors.Is(err, http.ErrNoCookie) {
		sendJSONError(w, "Invalid session cookie", http.StatusBadRequest)
		return
	}

	if err == nil && strings.TrimSpace(cookie.Value) != "" {
		if deleteErr := h.SessionStore.DeleteSession(cookie.Value); deleteErr != nil {
			log.Printf("Error deleting session: %v", deleteErr)
			sendJSONError(w, "Could not process logout", http.StatusInternalServerError)
			return
		}
	}

	clearSessionCookie(w, isSecureRequest(r))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"message": "Logged out successfully"})
}

func (h *HandlerStruct) SessionHandler(w http.ResponseWriter, r *http.Request) {
	if h.SessionStore == nil || h.UserRepo == nil {
		sendJSONError(w, "Auth services are not configured", http.StatusInternalServerError)
		return
	}

	cookie, err := r.Cookie("session_id")
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			sendJSONError(w, "Not authenticated", http.StatusUnauthorized)
			return
		}
		sendJSONError(w, "Invalid session cookie", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(cookie.Value) == "" {
		sendJSONError(w, "Not authenticated", http.StatusUnauthorized)
		return
	}

	session, err := h.SessionStore.GetSessionByID(cookie.Value)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			clearSessionCookie(w, isSecureRequest(r))
			sendJSONError(w, "Session expired or invalid", http.StatusUnauthorized)
			return
		}
		log.Printf("Error loading session: %v", err)
		sendJSONError(w, "Could not load session", http.StatusInternalServerError)
		return
	}

	user, err := h.UserRepo.GetUserByID(session.UserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			clearSessionCookie(w, isSecureRequest(r))
			sendJSONError(w, "Session user not found", http.StatusUnauthorized)
			return
		}
		log.Printf("Error loading session user: %v", err)
		sendJSONError(w, "Could not load session", http.StatusInternalServerError)
		return
	}

	user.Password = ""
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"session": map[string]interface{}{
			"id":        cookie.Value,
			"userId":    session.UserID,
			"expiresAt": session.ExpiresAt,
			"createdAt": session.CreatedAt,
		},
		"user": user,
	})
}

func isValidEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

func hashPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashed), nil
}

func verifyPassword(password, stored string) bool {
	// Prefer bcrypt for all new passwords.
	if strings.HasPrefix(stored, "$2a$") || strings.HasPrefix(stored, "$2b$") || strings.HasPrefix(stored, "$2y$") {
		return bcrypt.CompareHashAndPassword([]byte(stored), []byte(password)) == nil
	}

	// Legacy format fallback for existing records.
	parts := strings.Split(stored, "$")
	if len(parts) != 3 || parts[0] != "sha256" {
		return false
	}

	salt, err := hex.DecodeString(parts[1])
	if err != nil {
		return false
	}
	expectedHash, err := hex.DecodeString(parts[2])
	if err != nil {
		return false
	}

	computed := sha256Legacy(append(salt, []byte(password)...))
	return subtleConstantTimeCompare(computed, expectedHash)
}

func generateSessionToken() (string, error) {
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", err
	}

	return hex.EncodeToString(tokenBytes), nil
}

func sha256Legacy(data []byte) []byte {
	h := sha256.Sum256(data)
	return h[:]
}

func subtleConstantTimeCompare(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	result := byte(0)
	for i := range a {
		result |= a[i] ^ b[i]
	}
	return result == 0
}

// extractClientIP returns the real client IP, respecting common reverse-proxy headers.
func extractClientIP(r *http.Request) string {
	if xff := strings.TrimSpace(r.Header.Get("X-Forwarded-For")); xff != "" {
		return strings.TrimSpace(strings.SplitN(xff, ",", 2)[0])
	}
	if xri := strings.TrimSpace(r.Header.Get("X-Real-IP")); xri != "" {
		return xri
	}
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

func isSecureRequest(r *http.Request) bool {
	if r.TLS != nil {
		return true
	}
	return strings.EqualFold(strings.TrimSpace(r.Header.Get("X-Forwarded-Proto")), "https")
}

func clearSessionCookie(w http.ResponseWriter, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   secure,
		MaxAge:   -1,
		Expires:  time.Unix(0, 0).UTC(),
	})
}

func parseFormBool(raw string) (bool, error) {
	normalized := strings.ToLower(strings.TrimSpace(raw))
	switch normalized {
	case "on", "yes":
		return true, nil
	}
	return strconv.ParseBool(normalized)
}

func saveAvatarFile(file multipart.File, fileHeader *multipart.FileHeader) (string, error) {
	buffer := make([]byte, 512)
	readBytes, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return "", errors.New("could not read avatar file")
	}

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return "", errors.New("could not process avatar file")
	}

	contentType := http.DetectContentType(buffer[:readBytes])
	allowed := map[string]string{
		"image/jpeg": ".jpg",
		"image/png":  ".png",
		"image/gif":  ".gif",
		"image/webp": ".webp",
	}
	ext, ok := allowed[contentType]
	if !ok {
		return "", errors.New("avatar must be jpg, png, gif, or webp")
	}

	if fileHeader != nil && fileHeader.Size > 5*1024*1024 {
		return "", errors.New("avatar file size must be 5MB or less")
	}

	fileName := uuid.NewString() + ext
	savePath := filepath.Join("uploads", "avatars", fileName)

	if err := os.MkdirAll(filepath.Dir(savePath), os.ModePerm); err != nil {
		return "", errors.New("could not save avatar file")
	}

	dst, err := os.Create(savePath)
	if err != nil {
		return "", errors.New("could not save avatar file")
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		return "", errors.New("could not save avatar file")
	}

	return "/" + filepath.ToSlash(savePath), nil
}

func removeUploadedFile(publicPath string) {
	localPath := strings.TrimPrefix(publicPath, "/")
	if localPath == "" {
		return
	}
	_ = os.Remove(localPath)
}
