package websocket

import "time"

// Message type constants
const (
	MsgFriendRequest  = "friend_request"
	MsgFriendAccepted = "friend_accepted"
	MsgActivityLiked  = "activity_liked"
	MsgActivityComment = "activity_comment"
)

// Message is the payload sent to WebSocket clients
type Message struct {
	Type      string    `json:"type"`
	UserID    int       `json:"user_id"`
	Payload   any       `json:"payload"`
	Timestamp time.Time `json:"timestamp"`
}
