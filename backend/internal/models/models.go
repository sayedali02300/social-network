package models

import (
	"time"
)

type User struct {
	ID          string    `json:"id"`
	Email       string    `json:"email"`
	Password    string    `json:"-"`
	FirstName   string    `json:"firstName"`
	LastName    string    `json:"lastName"`
	DateOfBirth string    `json:"dateOfBirth"`
	Avatar      string    `json:"avatar"`
	Nickname    string    `json:"nickname"`
	AboutMe     string    `json:"aboutMe"`
	IsPublic    bool      `json:"isPublic"`
	CreatedAt   time.Time `json:"createdAt"`
}

type Session struct {
	ID        string    `json:"id"`
	UserID    string    `json:"userId"`
	ExpiresAt time.Time `json:"expiresAt"`
	CreatedAt time.Time `json:"createdAt"`
	IPAddress string    `json:"ipAddress"`
	UserAgent string    `json:"userAgent"`
}

// Follower is the user that follows, Following is who he follows
type Follower struct {
	FollowerID  string    `json:"followerId"`
	FollowingID string    `json:"followingId"`
	CreatedAt   time.Time `json:"createdAt"`
}

type FollowRequest struct {
	ID         string    `json:"id"`
	SenderID   string    `json:"senderId"`
	ReceiverID string    `json:"receiverId"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"createdAt"`
}

type Post struct {
	ID           string    `json:"id"`
	UserID       string    `json:"userId"`
	GroupID      string    `json:"groupId,omitempty"`
	Title        string    `json:"title"`
	Content      string    `json:"content"`
	ImagePath    string    `json:"imagePath"`
	Privacy      string    `json:"privacy"` // public, almost_private, private
	CreatedAt    time.Time `json:"createdAt"`
	Author       User      `json:"author"`
	LikeCount    int       `json:"likeCount"`
	DislikeCount int       `json:"dislikeCount"`
	MyReaction   int       `json:"myReaction"`
}

type PostLike struct {
	PostID    string    `json:"postId"`
	UserID    string    `json:"userId"`
	Value     int       `json:"value"`
	CreatedAt time.Time `json:"createdAt"`
}

type Comment struct {
	ID        string    `json:"id"`
	PostID    string    `json:"postId"`
	UserID    string    `json:"userId"`
	Content   string    `json:"content"`
	ImagePath string    `json:"imagePath"`
	ParentID  string    `json:"parentId,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
}

type CommentWithAuthor struct {
	ID        string        `json:"id"`
	PostID    string        `json:"postId"`
	Content   string        `json:"content"`
	ImagePath string        `json:"imagePath,omitempty"`
	ParentID  string        `json:"parentId,omitempty"`
	CreatedAt time.Time     `json:"createdAt"`
	Author    CommentAuthor `json:"author"`
}

type CommentAuthor struct {
	UserID    string `json:"user_id,omitempty"`
	Avatar    string `json:"avatar,omitempty"`
	Nickname  string `json:"nickname,omitempty"`
	FirstName string `json:"firstName,omitempty"`
	LastName  string `json:"lastName,omitempty"`
}

type Message struct {
	ID         string    `json:"id"`
	SenderID   string    `json:"senderId"`
	ReceiverID string    `json:"receiverId"` // NULL if group message
	GroupID    string    `json:"groupId"`    // NULL if private message
	Content    string    `json:"content"`
	CreatedAt  time.Time `json:"createdAt"`
}

type Group struct {
	ID          string    `json:"id"`
	CreatorID   string    `json:"creatorId"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"createdAt"`
}

type GroupSummary struct {
	ID                string    `json:"id"`
	CreatorID         string    `json:"creatorId"`
	CreatorName       string    `json:"creatorName"`
	Title             string    `json:"title"`
	Description       string    `json:"description"`
	MembersCount      int       `json:"membersCount"`
	IsMember          bool      `json:"isMember"`
	HasPendingInvite  bool      `json:"hasPendingInvite"`
	HasPendingRequest bool      `json:"hasPendingRequest"`
	CreatedAt         time.Time `json:"createdAt"`
}

type GroupMember struct {
	UserID   string    `json:"userId"`
	Nickname string    `json:"nickname"`
	Avatar   string    `json:"avatar"`
	Role     string    `json:"role"`
	JoinedAt time.Time `json:"joinedAt"`
}

type GroupInvite struct {
	ID         string    `json:"id"`
	GroupID    string    `json:"groupId"`
	SenderID   string    `json:"senderId"`
	ReceiverID string    `json:"receiverId"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"createdAt"`
}

type GroupJoinRequest struct {
	ID        string    `json:"id"`
	GroupID   string    `json:"groupId"`
	UserID    string    `json:"userId"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
}

type Event struct {
	ID            string    `json:"id"`
	GroupID       string    `json:"groupId"`
	CreatorID     string    `json:"creatorId"`
	Title         string    `json:"title"`
	Description   string    `json:"description"`
	EventTime     time.Time `json:"eventTime"`
	CreatedAt     time.Time `json:"createdAt"`
	GoingCount    int       `json:"goingCount"`
	NotGoingCount int       `json:"notGoingCount"`
	MyResponse    string    `json:"myResponse"`
}

type EventResponse struct {
	EventID   string    `json:"eventId"`
	UserID    string    `json:"userId"`
	Response  string    `json:"response"`
	CreatedAt time.Time `json:"createdAt"`
}

type Notification struct {
	ID        string    `json:"id"`
	UserID    string    `json:"userId"`
	ActorID   string    `json:"actorId"`
	Type      string    `json:"type"`
	Payload   string    `json:"payload"`
	IsRead    bool      `json:"isRead"`
	CreatedAt time.Time `json:"createdAt"`
}
