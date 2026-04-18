# Backend — Social Network API

Go HTTP server providing the REST API and WebSocket layer for the social network platform. Uses SQLite for persistence and Gorilla WebSocket for real-time communication.

---

## Tech Stack

| Concern | Choice |
|---------|--------|
| Language | Go 1.24 |
| HTTP router | `net/http` (Go 1.22 pattern routing) |
| Database | SQLite via `go-sqlite3` |
| Migrations | File-based, applied on startup |
| WebSocket | `gorilla/websocket` |
| Auth | Session cookies (`gorilla/securecookie` via session store) |
| Password hashing | `golang.org/x/crypto` (bcrypt) |
| IDs | `google/uuid` |

---

## Project Structure

```
backend/
├── cmd/server/          # Entry point — wires config, DB, repos, handlers, routes
├── internal/
│   ├── handlers/        # HTTP handlers and middleware (auth, posts, groups, notifications …)
│   ├── models/          # Shared domain structs
│   └── repository/      # Database access layer (users, posts, groups, sessions, comments)
├── pkg/
│   ├── config/          # Environment-based config loader
│   └── db/sqlite/       # SQLite open helper and migration runner
├── utils/               # Background utilities (orphaned image cleanup)
└── Dockerfile           # Multi-stage build (builder → debian:bookworm-slim)
```

---

## Running Locally

**Prerequisites:** Go 1.22+, GCC (required by `go-sqlite3` for CGO)

```bash
cd backend
go run ./cmd/server
```

The server starts on `:8080` by default.

---

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `SERVER_ADDR` | `:8080` | Address and port the server listens on |
| `DB_PATH` | `social-network.db` | Path to the SQLite database file |
| `MIGRATIONS_DIR` | `pkg/db/migrations/sqlite` | Directory containing numbered migration files |
| `SEED_DB` | `false` | Set to `true` to seed the database on startup |

---

## Running with Docker

A multi-stage `Dockerfile` is included. The recommended way is via `docker-compose` from the repo root:

```bash
docker-compose up --build
```

To run the backend container directly:

```bash
docker build -t social-network-backend .
docker run -p 8080:8080 social-network-backend
```

---

## API Overview

All routes under `/api/` require a valid session cookie unless noted. The WebSocket endpoint is `/ws`.

### Auth
| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/auth/register` | Register a new user |
| POST | `/api/auth/login` | Log in and receive a session cookie |
| POST | `/api/auth/logout` | Invalidate the current session |
| GET | `/api/auth/session` | Return the current session user |

### Users & Follows
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/users/me` | Get the authenticated user's profile |
| PATCH | `/api/users/me` | Update profile fields |
| PATCH | `/api/users/me/privacy` | Toggle public / private account |
| GET | `/api/users/search` | Search users by name or nickname |
| GET | `/api/users/{userId}` | Get a user's public profile |
| GET | `/api/users/{userId}/followers` | List a user's followers |
| GET | `/api/users/{userId}/following` | List who a user follows |
| POST | `/api/follow-requests` | Send a follow request |
| GET | `/api/follow-requests/incoming` | List incoming follow requests |
| GET | `/api/follow-requests/outgoing` | List outgoing follow requests |
| PATCH | `/api/follow-requests/{requestId}` | Accept or decline a follow request |
| DELETE | `/api/follow-requests/{requestId}` | Cancel an outgoing follow request |
| DELETE | `/api/followers/{userId}` | Remove a follower |
| DELETE | `/api/following/{userId}` | Unfollow a user |

### Posts & Comments
| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/posts` | Create a post (supports image upload) |
| GET | `/api/posts/feed` | Get the authenticated user's feed |
| GET | `/api/posts/{postId}` | Get a single post |
| PATCH | `/api/posts/{postId}` | Edit a post |
| DELETE | `/api/posts/{postId}` | Delete a post |
| POST/DELETE/GET | `/api/posts/{postId}/allowed-users` | Manage close-friends visibility list |
| POST | `/api/posts/{postId}/comments` | Add a comment |
| GET | `/api/posts/{postId}/comments` | List comments on a post |
| PATCH | `/api/comments/{commentId}` | Edit a comment |
| DELETE | `/api/comments/{commentId}` | Delete a comment |

### Groups & Events
| Method | Path | Description |
|--------|------|-------------|
| GET/POST | `/api/groups` | List all groups or create a new group |
| GET/PATCH/DELETE | `/api/groups/{groupId}` | Get, update, or delete a group |
| GET/POST | `/api/groups/{groupId}/invites` | List or send group invites |
| PATCH | `/api/group-invites/{inviteId}` | Accept or decline an invite |
| GET/POST | `/api/groups/{groupId}/requests` | List or send join requests |
| PATCH | `/api/group-requests/{requestId}` | Approve or decline a join request |
| GET/DELETE | `/api/groups/{groupId}/members` | List members or remove a member |
| DELETE | `/api/groups/{groupId}/members/{userId}` | Remove a specific member |
| GET/POST | `/api/groups/{groupId}/events` | List or create group events |
| GET/POST | `/api/groups/{groupId}/posts` | List or create group posts |
| GET/PATCH/DELETE | `/api/events/{eventId}` | Get, update, or delete an event |
| GET/POST | `/api/events/{eventId}/responses` | Get or submit an RSVP response |

### Notifications
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/notifications` | List all notifications |
| GET | `/api/notifications/unread-count` | Get unread notification count |
| PATCH | `/api/notifications/{notificationId}/read` | Mark a notification as read |
| PATCH | `/api/notifications/read-all` | Mark all notifications as read |

### Chat
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/chats/private/{chatId}/messages` | Fetch private chat history |
| GET | `/api/chats/groups/{chatId}/messages` | Fetch group chat history |
| GET | `/api/chats/unread-counts` | Get unread message counts per chat |
| POST/DELETE | `/api/chats/{chatType}/{chatId}/mute` | Mute or unmute a chat |

### WebSocket
| Path | Description |
|------|-------------|
| `/ws` | Persistent connection for real-time messaging, notifications, and group events |

### Other
| Path | Description |
|------|-------------|
| `GET /health` | Health check — returns `{"status":"ok"}` |
| `GET /uploads/{type}/{filename}` | Serve uploaded images (avatars, posts, comments) |

---

## WebSocket Protocol

Messages are JSON envelopes with a `type` field and a `payload` field:

```json
{ "type": "event_type", "payload": { ... } }
```

Key outgoing event types pushed from the server:

| Type | Description |
|------|-------------|
| `notification` | New notification for the user |
| `notification_unread_count` | Updated unread notification badge count |
| `group_summary_event` | Updated group summary pushed to all members |
| `group_state_refresh` | Signal to re-fetch group state |
| `group_post_event` | New or updated group post |
| `group_post_delete` | Group post deleted |
| `group_comment_event` | New or updated group comment |
| `group_comment_delete` | Group comment deleted |
| `group_calendar_event` | New or updated group event |
| `group_calendar_delete` | Group event deleted |
| `private_message` | New private chat message |
| `group_message` | New group chat message |

---

## Database

SQLite with foreign keys enabled (`_foreign_keys=on`). Migrations are numbered SQL files under `pkg/db/migrations/sqlite/` and are applied automatically on every startup via `ApplyMigrations()`.

Uploaded files are stored under `uploads/` (avatars, posts, comments) and served through the authenticated `/uploads/` route. A background goroutine runs on startup to clean up orphaned image files.
