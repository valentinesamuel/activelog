# MONTH 8: Social Features & Real-time

**Weeks:** 29-32
**Phase:** Social Interaction & Real-time Communication
**Theme:** Connect users and enable real-time updates

---

## Overview

This month adds social features to your application. You'll implement a friend system, activity feeds, real-time notifications with WebSockets, and a feature flags system for gradual rollouts. By the end, users can connect with friends, see their activities in real-time, and interact through likes and comments.

---

## Learning Path

### Week 29: Friend System
- Friend request/accept/reject flow
- Many-to-many relationship
- Friendship status management
- Friend list queries

### Week 30: Activity Feed + Feature Flags System (60 min)
- Aggregate friends' activities
- Pagination and infinite scroll
- Feed algorithms (chronological)
- **NEW:** Feature flags for gradual rollouts

### Week 31: WebSocket Integration
- WebSocket protocol basics
- Hub pattern for managing connections
- Broadcasting messages
- Real-time notifications

### Week 32: Comments + Likes
- Comment system implementation
- Like/unlike functionality
- Notification triggers
- Activity engagement metrics

---

## Features

### API Endpoints
```
# Friend System
POST   /api/v1/friends/request/:id    # Send friend request
POST   /api/v1/friends/accept/:id     # Accept request
POST   /api/v1/friends/reject/:id     # Reject request
DELETE /api/v1/friends/:id            # Remove friend
GET    /api/v1/friends                # List friends
GET    /api/v1/friends/requests       # Pending requests

# Activity Feed
GET    /api/v1/feed                   # Friends' activity feed

# Interactions
POST   /api/v1/activities/:id/comment # Add comment
DELETE /api/v1/activities/:id/comment/:commentId # Delete comment
POST   /api/v1/activities/:id/like    # Like activity
DELETE /api/v1/activities/:id/like    # Unlike activity

# WebSocket
WS     /ws                            # WebSocket connection
```

---

## Database Schema

```sql
-- Friendship table (self-referencing many-to-many)
CREATE TABLE friendships (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    friend_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    status VARCHAR(20) NOT NULL, -- pending, accepted, rejected
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, friend_id)
);

CREATE INDEX idx_friendships_user ON friendships(user_id, status);
CREATE INDEX idx_friendships_friend ON friendships(friend_id, status);

-- Comments
CREATE TABLE comments (
    id SERIAL PRIMARY KEY,
    activity_id INTEGER REFERENCES activities(id) ON DELETE CASCADE,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_comments_activity ON comments(activity_id);

-- Likes
CREATE TABLE likes (
    id SERIAL PRIMARY KEY,
    activity_id INTEGER REFERENCES activities(id) ON DELETE CASCADE,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(activity_id, user_id)
);

CREATE INDEX idx_likes_activity ON likes(activity_id);
CREATE INDEX idx_likes_user ON likes(user_id);

-- Notifications
CREATE TABLE notifications (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL, -- friend_request, comment, like, etc.
    title VARCHAR(255) NOT NULL,
    message TEXT,
    is_read BOOLEAN DEFAULT FALSE,
    related_id INTEGER, -- ID of related entity (friend_id, comment_id, etc.)
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_notifications_user ON notifications(user_id, is_read);
```

---

## Friend System Implementation

```go
type FriendService struct {
    repo *repository.FriendRepository
}

// Send friend request
func (s *FriendService) SendRequest(ctx context.Context, userID, friendID int) error {
    // Check if already friends or request exists
    exists, err := s.repo.FriendshipExists(ctx, userID, friendID)
    if err != nil {
        return err
    }
    if exists {
        return ErrFriendshipExists
    }

    // Create friendship with pending status
    friendship := &models.Friendship{
        UserID:   userID,
        FriendID: friendID,
        Status:   "pending",
    }

    if err := s.repo.Create(ctx, friendship); err != nil {
        return err
    }

    // Create notification for recipient
    notification := &models.Notification{
        UserID:    friendID,
        Type:      "friend_request",
        Title:     "New friend request",
        Message:   fmt.Sprintf("User %d sent you a friend request", userID),
        RelatedID: userID,
    }

    return s.notificationRepo.Create(ctx, notification)
}

// Accept friend request
func (s *FriendService) AcceptRequest(ctx context.Context, userID, requesterID int) error {
    // Update status to accepted
    if err := s.repo.UpdateStatus(ctx, requesterID, userID, "accepted"); err != nil {
        return err
    }

    // Create reciprocal friendship
    friendship := &models.Friendship{
        UserID:   userID,
        FriendID: requesterID,
        Status:   "accepted",
    }

    if err := s.repo.Create(ctx, friendship); err != nil {
        return err
    }

    // Notify requester
    notification := &models.Notification{
        UserID:    requesterID,
        Type:      "friend_accepted",
        Title:     "Friend request accepted",
        Message:   fmt.Sprintf("User %d accepted your friend request", userID),
        RelatedID: userID,
    }

    return s.notificationRepo.Create(ctx, notification)
}

// Get friends list
func (r *FriendRepository) GetFriends(ctx context.Context, userID int) ([]*models.User, error) {
    query := `
        SELECT u.id, u.email, u.name, u.created_at
        FROM users u
        INNER JOIN friendships f ON u.id = f.friend_id
        WHERE f.user_id = $1 AND f.status = 'accepted'
        ORDER BY u.name
    `

    rows, err := r.db.QueryContext(ctx, query, userID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    friends := []*models.User{}
    for rows.Next() {
        var user models.User
        if err := rows.Scan(&user.ID, &user.Email, &user.Name, &user.CreatedAt); err != nil {
            return nil, err
        }
        friends = append(friends, &user)
    }

    return friends, nil
}
```

---

## Activity Feed

```go
// Get activity feed from friends
func (s *FeedService) GetFeed(ctx context.Context, userID int, limit, offset int) ([]*models.Activity, error) {
    query := `
        SELECT
            a.id, a.user_id, a.activity_type, a.duration_minutes, a.distance_km,
            a.notes, a.activity_date, a.created_at,
            u.name as user_name,
            COUNT(DISTINCT l.id) as like_count,
            COUNT(DISTINCT c.id) as comment_count
        FROM activities a
        INNER JOIN friendships f ON a.user_id = f.friend_id
        INNER JOIN users u ON a.user_id = u.id
        LEFT JOIN likes l ON a.id = l.activity_id
        LEFT JOIN comments c ON a.id = c.activity_id
        WHERE f.user_id = $1 AND f.status = 'accepted' AND a.deleted_at IS NULL
        GROUP BY a.id, u.name
        ORDER BY a.created_at DESC
        LIMIT $2 OFFSET $3
    `

    rows, err := s.repo.db.QueryContext(ctx, query, userID, limit, offset)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    activities := []*models.Activity{}
    for rows.Next() {
        var activity models.Activity
        err := rows.Scan(
            &activity.ID, &activity.UserID, &activity.Type, &activity.Duration,
            &activity.Distance, &activity.Notes, &activity.Date, &activity.CreatedAt,
            &activity.UserName, &activity.LikeCount, &activity.CommentCount,
        )
        if err != nil {
            return nil, err
        }
        activities = append(activities, &activity)
    }

    return activities, nil
}
```

---

## WebSocket Hub

```go
import "github.com/gorilla/websocket"

type Hub struct {
    // Registered clients (user_id -> *Client)
    clients map[int]*Client

    // Broadcast channel
    broadcast chan Message

    // Register client
    register chan *Client

    // Unregister client
    unregister chan *Client

    mu sync.RWMutex
}

type Client struct {
    hub    *Hub
    conn   *websocket.Conn
    userID int
    send   chan []byte
}

type Message struct {
    Type      string      `json:"type"`
    UserID    int         `json:"user_id,omitempty"`
    Payload   interface{} `json:"payload"`
    Timestamp time.Time   `json:"timestamp"`
}

func NewHub() *Hub {
    return &Hub{
        clients:    make(map[int]*Client),
        broadcast:  make(chan Message),
        register:   make(chan *Client),
        unregister: make(chan *Client),
    }
}

func (h *Hub) Run() {
    for {
        select {
        case client := <-h.register:
            h.mu.Lock()
            h.clients[client.userID] = client
            h.mu.Unlock()

        case client := <-h.unregister:
            h.mu.Lock()
            if _, ok := h.clients[client.userID]; ok {
                delete(h.clients, client.userID)
                close(client.send)
            }
            h.mu.Unlock()

        case message := <-h.broadcast:
            h.mu.RLock()
            for _, client := range h.clients {
                select {
                case client.send <- encodeMessage(message):
                default:
                    // Client's send buffer is full, disconnect
                    close(client.send)
                    delete(h.clients, client.userID)
                }
            }
            h.mu.RUnlock()
        }
    }
}

// Send message to specific user
func (h *Hub) SendToUser(userID int, msgType string, payload interface{}) {
    h.mu.RLock()
    client, ok := h.clients[userID]
    h.mu.RUnlock()

    if ok {
        message := Message{
            Type:      msgType,
            Payload:   payload,
            Timestamp: time.Now(),
        }
        client.send <- encodeMessage(message)
    }
}

// WebSocket handler
var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool {
        return true // Configure properly in production
    },
}

func (h *Hub) ServeWS(w http.ResponseWriter, r *http.Request) {
    userID := getUserID(r.Context())

    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Println("Upgrade error:", err)
        return
    }

    client := &Client{
        hub:    h,
        conn:   conn,
        userID: userID,
        send:   make(chan []byte, 256),
    }

    h.register <- client

    // Start goroutines for reading and writing
    go client.writePump()
    go client.readPump()
}

func (c *Client) readPump() {
    defer func() {
        c.hub.unregister <- c
        c.conn.Close()
    }()

    for {
        _, message, err := c.conn.ReadMessage()
        if err != nil {
            break
        }

        // Handle incoming message
        handleClientMessage(c, message)
    }
}

func (c *Client) writePump() {
    ticker := time.NewTicker(54 * time.Second) // Ping interval
    defer func() {
        ticker.Stop()
        c.conn.Close()
    }()

    for {
        select {
        case message, ok := <-c.send:
            if !ok {
                c.conn.WriteMessage(websocket.CloseMessage, []byte{})
                return
            }

            c.conn.WriteMessage(websocket.TextMessage, message)

        case <-ticker.C:
            if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
                return
            }
        }
    }
}
```

---

## 🔴 Feature Flags

```go
// Feature flags for gradual rollouts
type FeatureFlags struct {
    EnableComments bool
    EnableLikes    bool
    EnableFriends  bool
}

// Load from environment or config
func LoadFeatureFlags() *FeatureFlags {
    return &FeatureFlags{
        EnableComments: os.Getenv("FEATURE_COMMENTS") == "enabled",
        EnableLikes:    os.Getenv("FEATURE_LIKES") == "enabled",
        EnableFriends:  os.Getenv("FEATURE_FRIENDS") == "enabled",
    }
}

// Check if feature is enabled for user
func (f *FeatureFlags) IsEnabled(userID int, feature string) bool {
    switch feature {
    case "comments":
        return f.EnableComments
    case "likes":
        return f.EnableLikes
    case "friends":
        return f.EnableFriends
    default:
        return false
    }
}

// Middleware to check feature flags
func (m *FeatureFlagMiddleware) CheckFeature(feature string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            userID := getUserID(r.Context())

            if !m.flags.IsEnabled(userID, feature) {
                response.Error(w, http.StatusForbidden, "Feature not available")
                return
            }

            next.ServeHTTP(w, r)
        })
    }
}

// Usage in router
router.With(featureFlag.CheckFeature("comments")).Post("/activities/{id}/comment", handler.AddComment)
```

**Use Cases:**
- **Gradual feature rollouts** (beta testing with select users)
- **A/B testing new features** (compare variants)
- **Quick feature disabling in production** (kill switch)
- **Per-user feature access** (premium features)

---

## Likes & Comments

```go
// Like an activity
func (s *LikeService) Like(ctx context.Context, activityID, userID int) error {
    like := &models.Like{
        ActivityID: activityID,
        UserID:     userID,
    }

    if err := s.repo.Create(ctx, like); err != nil {
        return err
    }

    // Get activity owner
    activity, _ := s.activityRepo.GetByID(ctx, activityID)

    // Send real-time notification
    s.hub.SendToUser(activity.UserID, "activity_liked", map[string]interface{}{
        "activity_id": activityID,
        "user_id":     userID,
    })

    // Create notification
    notification := &models.Notification{
        UserID:    activity.UserID,
        Type:      "like",
        Title:     "Someone liked your activity",
        RelatedID: activityID,
    }

    return s.notificationRepo.Create(ctx, notification)
}

// Add comment
func (s *CommentService) AddComment(ctx context.Context, activityID, userID int, content string) error {
    comment := &models.Comment{
        ActivityID: activityID,
        UserID:     userID,
        Content:    content,
    }

    if err := s.repo.Create(ctx, comment); err != nil {
        return err
    }

    // Get activity owner
    activity, _ := s.activityRepo.GetByID(ctx, activityID)

    // Send real-time notification
    s.hub.SendToUser(activity.UserID, "new_comment", map[string]interface{}{
        "activity_id": activityID,
        "comment_id":  comment.ID,
        "user_id":     userID,
        "content":     content,
    })

    return nil
}
```

---

## Common Pitfalls

1. **Not handling WebSocket disconnections**
   - ❌ Memory leaks from orphaned connections
   - ✅ Properly clean up in hub

2. **Sending too many notifications**
   - ❌ Spamming users
   - ✅ Batch or throttle notifications

3. **N+1 queries in feed**
   - ❌ Loading likes/comments separately
   - ✅ Use JOINs and aggregates

4. **No feature flags**
   - ❌ All-or-nothing deployments
   - ✅ Gradual rollouts with flags

---

## Resources

- [gorilla/websocket](https://github.com/gorilla/websocket)
- [WebSocket Protocol RFC](https://datatracker.ietf.org/doc/html/rfc6455)
- [Feature Flags Best Practices](https://launchdarkly.com/blog/feature-flag-best-practices/)

---

## Next Steps

After completing Month 8, you'll move to **Month 9: Analytics & Reporting**, where you'll learn:
- Advanced SQL queries
- Goal tracking system
- PDF report generation
- Chart data endpoints

**Your app now has engaging social features!** 🎉
