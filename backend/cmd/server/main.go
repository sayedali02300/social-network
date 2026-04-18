package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"socialnetwork/backend/internal/handlers"
	"socialnetwork/backend/internal/repository"
	wshandler "socialnetwork/backend/internal/ws"
	"socialnetwork/backend/pkg/config"
	"socialnetwork/backend/pkg/db/sqlite"
	"socialnetwork/backend/utils"
)

func main() {
	cfg := config.Load()

	db, err := sqlite.Open(cfg.DBPath)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := sqlite.ApplyMigrations(ctx, db, cfg.MigrationsDir); err != nil {
		log.Fatalf("apply migrations: %v", err)
	}
	sqlite.EnsureSessionColumns(db)
	go utils.CleanUpOrphanedImages(db, "uploads/posts", "posts")
	go utils.CleanUpOrphanedImages(db, "uploads/comments", "comments")
	// Initialize reposotiries and Handlers
	postRepo := repository.NewPostRepository(db)
	userRepo := repository.NewUserRepository(db)
	repoSessionStore := repository.NewSessionStore(db)
	groupRepo := repository.NewGroupRepository(db)
	commentRepo := repository.NewCommentRepository(db)
	likeRepo := repository.NewLikeRepository(db)
	h := handlers.NewHandler(commentRepo, postRepo, userRepo, repoSessionStore, groupRepo)
	h.LikeRepo = likeRepo
	wsHub := wshandler.NewHub()
	wsSessionStore := wshandler.NewSessionStore(db)
	chatStore := wshandler.NewChatStore(db)
	go wsHub.Run()
	h.SetNotifierMany(wsHub)
	h.SetNotifier(wsHub)
	h.StartEventDueNotifier()

	mux := http.NewServeMux()

	registerRoutes(mux, h, wsHub, wsSessionStore, chatStore)

	// Images
	mux.HandleFunc("GET /uploads/{type}/{filename}", h.RequireAuth(h.ServeUploadsHandler))

	server := &http.Server{
		Addr:         cfg.ServerAddr,
		Handler:      enableCORS(mux),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	log.Printf("server listening on %s", cfg.ServerAddr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
}

func registerRoutes(mux *http.ServeMux, h *handlers.HandlerStruct, wsHub *wshandler.Hub, wsSessionStore *wshandler.SessionStore, chatStore *wshandler.ChatStore) {
	// API ROUTES
	mux.HandleFunc("POST /api/auth/register", h.RegisterHandler)
	mux.HandleFunc("POST /api/auth/login", h.LoginHandler)
	mux.HandleFunc("POST /api/auth/logout", h.LogoutHandler)
	mux.HandleFunc("GET /api/auth/session", h.SessionHandler)
	mux.HandleFunc("GET /api/users/me", h.RequireAuth(h.GetMeHandler))
	mux.HandleFunc("PATCH /api/users/me", h.RequireAuth(h.PatchMeHandler))
	mux.HandleFunc("PATCH /api/users/me/privacy", h.RequireAuth(h.PatchMePrivacyHandler))
	mux.HandleFunc("GET /api/users/search", h.RequireAuth(h.SearchUsersHandler))
	mux.HandleFunc("GET /api/users/{userId}", h.RequireAuth(h.GetUserByIDHandler))
	mux.HandleFunc("GET /api/users/{userId}/followers", h.RequireAuth(h.GetFollowersHandler))
	mux.HandleFunc("GET /api/users/{userId}/following", h.RequireAuth(h.GetFollowingHandler))
	mux.HandleFunc("POST /api/follow-requests", h.RequireAuth(h.CreateFollowRequestHandler))
	mux.HandleFunc("GET /api/follow-requests/incoming", h.RequireAuth(h.GetIncomingFollowRequestsHandler))
	mux.HandleFunc("GET /api/follow-requests/outgoing", h.RequireAuth(h.GetOutgoingFollowRequestsHandler))
	mux.HandleFunc("PATCH /api/follow-requests/{requestId}", h.RequireAuth(h.PatchFollowRequestHandler))
	mux.HandleFunc("DELETE /api/follow-requests/{requestId}", h.RequireAuth(h.DeleteFollowRequestHandler))
	mux.HandleFunc("DELETE /api/followers/{userId}", h.RequireAuth(h.DeleteFollowerHandler))
	mux.HandleFunc("DELETE /api/following/{userId}", h.RequireAuth(h.DeleteFollowingHandler))
	mux.HandleFunc("GET /api/notifications", h.RequireAuth(h.GetNotificationsHandler))
	mux.HandleFunc("GET /api/notifications/unread-count", h.RequireAuth(h.GetNotificationsUnreadCountHandler))
	mux.HandleFunc("PATCH /api/notifications/{notificationId}/read", h.RequireAuth(h.PatchNotificationReadHandler))
	mux.HandleFunc("PATCH /api/notifications/read-all", h.RequireAuth(h.PatchNotificationsReadAllHandler))
	mux.HandleFunc("POST /api/posts", h.RequireAuth(h.PostingHandler))
	mux.HandleFunc("GET /api/posts/feed", h.RequireAuth(h.GetPostsHandler))
	mux.HandleFunc("GET /api/posts/{postId}", h.RequireAuth(h.GetOnePostHandler))
	mux.HandleFunc("PATCH /api/posts/{postId}", h.RequireAuth(h.PatchPostHandler))
	mux.HandleFunc("DELETE /api/posts/{postId}", h.RequireAuth(h.DeletePostHandler))
	mux.HandleFunc("POST /api/posts/{postId}/allowed-users", h.RequireAuth(h.AddAllowedUsersHandler))
	mux.HandleFunc("DELETE /api/posts/{postId}/allowed-users", h.RequireAuth(h.RemoveAllowedUsersHandler))
	mux.HandleFunc("GET /api/posts/{postId}/allowed-users", h.RequireAuth(h.GetAllowedUsersHandler))
	mux.HandleFunc("POST /api/posts/{postId}/like", h.RequireAuth(h.LikePostHandler))
	mux.HandleFunc("DELETE /api/posts/{postId}/like", h.RequireAuth(h.UnlikePostHandler))
	mux.HandleFunc("POST /api/posts/{postId}/comments", h.RequireAuth(h.PostCommentHandler))
	mux.HandleFunc("GET /api/posts/{postId}/comments", h.RequireAuth(h.GetCommentHandler))
	mux.HandleFunc("PATCH /api/comments/{commentID}", h.RequireAuth(h.PatchCommentHandler))
	mux.HandleFunc("DELETE /api/comments/{commentID}", h.RequireAuth(h.DeleteCommentHandler))
	mux.HandleFunc("/api/groups", h.RequireAuth(h.GroupsCollectionHandler))
	mux.HandleFunc("/api/groups/{groupId}", h.RequireAuth(h.GroupHandler))
	mux.HandleFunc("/api/groups/{groupId}/invites", h.RequireAuth(h.GroupInvitesHandler))
	mux.HandleFunc("/api/group-invites/{inviteId}", h.RequireAuth(h.GroupInviteResponseHandler))
	mux.HandleFunc("/api/groups/{groupId}/requests", h.RequireAuth(h.GroupRequestsHandler))
	mux.HandleFunc("/api/group-requests/{requestId}", h.RequireAuth(h.GroupRequestResponseHandler))
	mux.HandleFunc("/api/groups/{groupId}/members", h.RequireAuth(h.GroupMembersHandler))
	mux.HandleFunc("/api/groups/{groupId}/members/{userId}", h.RequireAuth(h.GroupMemberHandler))
	mux.HandleFunc("/api/groups/{groupId}/events", h.RequireAuth(h.GroupEventsHandler))
	mux.HandleFunc("/api/groups/{groupId}/posts", h.RequireAuth(h.GroupPostsHandler))
	mux.HandleFunc("/api/events/{eventId}", h.RequireAuth(h.EventHandler))
	mux.HandleFunc("/api/events/{eventId}/responses", h.RequireAuth(h.EventResponseHandler))
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/ws", wshandler.NewHandler(wsHub, wsSessionStore, chatStore, h.PostRepo, h.CommentRepo, h.CreateNotification))
	mux.HandleFunc("/api/chats/private/{chatId}/messages", wshandler.NewPrivateHistoryHandler(wsSessionStore, chatStore))
	mux.HandleFunc("/api/chats/groups/{chatId}/messages", wshandler.NewGroupHistoryHandler(wsSessionStore, chatStore))
	mux.HandleFunc("GET /api/chats/unread-counts", wshandler.NewUnreadCountsHandler(wsSessionStore, chatStore))
	mux.HandleFunc("/api/chats/{chatType}/{chatId}/mute", wshandler.NewChatMuteHandler(wsSessionStore, chatStore))
}

func healthHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := strings.TrimSpace(r.Header.Get("Origin"))
		allowOrigin := "http://localhost:5173"
		// Allow local frontend dev servers (e.g. :5173, :5174) on localhost/127.0.0.1.
		if strings.HasPrefix(origin, "http://localhost:") || strings.HasPrefix(origin, "http://127.0.0.1:") {
			allowOrigin = origin
		}
		w.Header().Set("Access-Control-Allow-Origin", allowOrigin)
		w.Header().Set("Vary", "Origin")
		// methods allowed
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, PATCH, DELETE")
		// allow send authorization header
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// TODO when cookies we have to add [Access-Control-Allow-Credentials]
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		// before the site send POST or DELETE request it asks the server
		// am i allowed? thoguh opions
		//here we send status 200 meaning yes the browser is allowed
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
