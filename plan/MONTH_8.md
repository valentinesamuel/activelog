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

# WEEKLY TASK BREAKDOWNS

## Week 29: Friend System

### 📋 Implementation Tasks

**Task 1: Create Database Migrations** (30 min)
- [ ] Create migration `migrations/009_create_friendships.up.sql`
- [ ] Add friendships table with user_id, friend_id, status columns
- [ ] Add UNIQUE constraint on (user_id, friend_id)
- [ ] Create indexes on user_id and friend_id
- [ ] Add status CHECK constraint (pending, accepted, rejected)
- [ ] Run migration

**Task 2: Create Friend Models** (20 min)
- [ ] Create `internal/models/friendship.go`
- [ ] Define Friendship struct
- [ ] Define Friend struct (user info + friendship status)
- [ ] Add JSON tags for API responses

**Task 3: Create Friend Repository** (90 min)
- [ ] Create `internal/repository/friend_repository.go`
- [ ] Implement `Create(ctx, friendship) error` (send request)
- [ ] Implement `UpdateStatus(ctx, userID, friendID, status) error`
- [ ] Implement `FriendshipExists(ctx, userID, friendID) (bool, error)`
- [ ] Implement `GetFriends(ctx, userID) ([]*User, error)`
- [ ] Implement `GetPendingRequests(ctx, userID) ([]*User, error)`
- [ ] Implement `Delete(ctx, userID, friendID) error`

**Task 4: Create Friend Service** (120 min)
- [ ] Create `internal/services/friend_service.go`
- [ ] Implement `SendRequest(ctx, userID, friendID) error`
  - Check friendship doesn't exist
  - Create pending friendship
  - Create notification
- [ ] Implement `AcceptRequest(ctx, userID, requesterID) error`
  - Update status to accepted
  - Create reciprocal friendship
  - Create notification
- [ ] Implement `RejectRequest(ctx, userID, requesterID) error`
- [ ] Implement `RemoveFriend(ctx, userID, friendID) error`

**Task 5: Create Friend Handlers** (90 min)
- [ ] Create `internal/handlers/friend_handler.go`
- [ ] Implement `SendRequest(w, r)` - POST /friends/request/:id
- [ ] Implement `AcceptRequest(w, r)` - POST /friends/accept/:id
- [ ] Implement `RejectRequest(w, r)` - POST /friends/reject/:id
- [ ] Implement `RemoveFriend(w, r)` - DELETE /friends/:id
- [ ] Implement `ListFriends(w, r)` - GET /friends
- [ ] Implement `ListRequests(w, r)` - GET /friends/requests

**Task 6: Add Authorization Checks** (45 min)
- [ ] Verify user can't send request to themselves
- [ ] Verify user can only accept requests sent to them
- [ ] Verify user owns friendship before deleting
- [ ] Add rate limiting for friend requests (max 20/day)
- [ ] Test all authorization scenarios

**Task 7: Add Notifications** (60 min)
- [ ] Create notification on friend request sent
- [ ] Create notification on request accepted
- [ ] Include user info in notification
- [ ] Test notifications created correctly

### 📦 Files You'll Create/Modify

```
migrations/
├── 009_create_friendships.up.sql  [CREATE]
└── 009_create_friendships.down.sql [CREATE]

internal/
├── models/
│   └── friendship.go              [CREATE]
├── repository/
│   └── friend_repository.go       [CREATE]
├── services/
│   └── friend_service.go          [CREATE]
└── handlers/
    └── friend_handler.go          [CREATE]
```

### 🔄 Implementation Order

1. **Database**: Migration → Models
2. **Repository**: Friend repository methods
3. **Service**: Friend service with business logic
4. **Handlers**: HTTP endpoints
5. **Authorization**: Access control checks
6. **Notifications**: Friend request notifications

### ⚠️ Blockers to Watch For

- **Duplicate requests**: UNIQUE constraint prevents duplicates
- **Bidirectional**: Need TWO rows for accepted friendship (A→B and B→A)
- **Status transitions**: Validate state transitions (pending → accepted only)
- **Self-requests**: Prevent users from friending themselves
- **Notification spam**: Rate limit friend requests

### ✅ Definition of Done

- [ ] Can send friend requests
- [ ] Can accept/reject requests
- [ ] Can remove friends
- [ ] Can list friends and pending requests
- [ ] Authorization checks working
- [ ] Notifications created on actions
- [ ] All tests passing

---

## Week 30: Activity Feed + Feature Flags

### 📋 Implementation Tasks

**Task 1: Create Feed Repository** (90 min)
- [ ] Create `internal/repository/feed_repository.go`
- [ ] Implement `GetFeed(ctx, userID, limit, offset) ([]*Activity, error)`
- [ ] Join activities, friendships, users tables
- [ ] Include like_count and comment_count (LEFT JOIN + COUNT)
- [ ] Order by created_at DESC
- [ ] Support pagination

**Task 2: Create Feed Service** (60 min)
- [ ] Create `internal/services/feed_service.go`
- [ ] Implement `GetFeed(ctx, userID, limit, offset) ([]*Activity, error)`
- [ ] Fetch from repository
- [ ] Enrich with user data
- [ ] Cache feed results (TTL: 1 minute)

**Task 3: Create Feed Handler** (45 min)
- [ ] Create `internal/handlers/feed_handler.go`
- [ ] Implement `GetFeed(w, r)` - GET /feed
- [ ] Parse limit/offset from query params
- [ ] Default limit: 20, max: 100
- [ ] Return paginated feed

**Task 4: Optimize Feed Query** (75 min)
- [ ] Add composite index on (user_id, created_at)
- [ ] Add index on (activity_id) for likes/comments tables
- [ ] Test query with EXPLAIN ANALYZE
- [ ] Verify index usage
- [ ] Benchmark feed performance

**Task 5: Implement Feature Flags System** (90 min)
- [ ] Create `internal/featureflags/flags.go`
- [ ] Define FeatureFlags struct
- [ ] Load from environment variables
- [ ] Implement `IsEnabled(userID, feature) bool`
- [ ] Support percentage rollouts (e.g., 10% of users)
- [ ] Create middleware `CheckFeature(feature)`

**Task 6: Add Feature Flags to Endpoints** (45 min)
- [ ] Protect friend endpoints with "friends" flag
- [ ] Protect comments with "comments" flag
- [ ] Protect likes with "likes" flag
- [ ] Test feature toggles work
- [ ] Document available flags

**Task 7: Test Feed Performance** (30 min)
- [ ] Load test feed with 100 friends
- [ ] Load test feed with 1000 activities
- [ ] Verify response time < 200ms
- [ ] Check cache hit ratio

### 📦 Files You'll Create/Modify

```
internal/
├── repository/
│   └── feed_repository.go         [CREATE]
├── services/
│   └── feed_service.go            [CREATE]
├── handlers/
│   └── feed_handler.go            [CREATE]
├── featureflags/
│   ├── flags.go                   [CREATE]
│   ├── middleware.go              [CREATE]
│   └── flags_test.go              [CREATE]

cmd/api/
└── main.go                        [MODIFY - add feature flags]
```

### 🔄 Implementation Order

1. **Feed**: Repository → Service → Handler
2. **Optimization**: Indexes → Query tuning
3. **Feature flags**: System → Middleware
4. **Integration**: Apply flags to endpoints
5. **Testing**: Performance and load tests

### ⚠️ Blockers to Watch For

- **N+1 queries**: Use JOINs, not separate queries
- **Large feeds**: Pagination required
- **Stale data**: Cache invalidation on new activities
- **Feature flag defaults**: Default to disabled for safety
- **Percentage rollouts**: Consistent hashing for same user

### ✅ Definition of Done

- [ ] Feed shows friends' activities chronologically
- [ ] Pagination working (limit/offset)
- [ ] Feed query optimized with indexes
- [ ] Feature flags system implemented
- [ ] Can enable/disable features via environment
- [ ] Feed response time < 200ms
- [ ] All tests passing

---

## Week 31: WebSocket Integration

### 📋 Implementation Tasks

**Task 1: Install WebSocket Library** (15 min)
- [ ] Install: `go get github.com/gorilla/websocket`
- [ ] Review WebSocket documentation
- [ ] Test basic WebSocket connection

**Task 2: Create WebSocket Hub** (120 min)
- [ ] Create `internal/websocket/hub.go`
- [ ] Implement Hub struct with clients map
- [ ] Add register/unregister channels
- [ ] Add broadcast channel
- [ ] Implement `Run()` method (event loop)
- [ ] Use RWMutex for thread-safe client map
- [ ] Test hub manages clients correctly

**Task 3: Create WebSocket Client** (90 min)
- [ ] Create `internal/websocket/client.go`
- [ ] Implement Client struct with connection and send channel
- [ ] Implement `readPump()` (read from WebSocket)
- [ ] Implement `writePump()` (write to WebSocket)
- [ ] Add ping/pong for connection health
- [ ] Handle disconnections gracefully

**Task 4: Create WebSocket Handler** (60 min)
- [ ] Add `ServeWS(w, r)` handler
- [ ] Upgrade HTTP to WebSocket
- [ ] Extract user ID from auth
- [ ] Create client and register with hub
- [ ] Start read/write pumps
- [ ] Add route: GET /ws

**Task 5: Implement Real-Time Notifications** (90 min)
- [ ] Add `SendToUser(userID, msgType, payload)` to hub
- [ ] Send notification when friend request received
- [ ] Send notification when activity liked
- [ ] Send notification when comment added
- [ ] Test notifications received in real-time

**Task 6: Add WebSocket Message Types** (45 min)
- [ ] Define message types: friend_request, activity_liked, new_comment
- [ ] Create message structs with JSON tags
- [ ] Handle incoming messages from clients (if needed)
- [ ] Test message serialization

**Task 7: Monitor WebSocket Connections** (30 min)
- [ ] Track active connections gauge in Prometheus
- [ ] Track messages sent counter
- [ ] Track connection errors
- [ ] Add to Grafana dashboard

### 📦 Files You'll Create/Modify

```
internal/
├── websocket/
│   ├── hub.go                     [CREATE]
│   ├── client.go                  [CREATE]
│   ├── handler.go                 [CREATE]
│   ├── types.go                   [CREATE]
│   └── hub_test.go                [CREATE]

cmd/api/
└── main.go                        [MODIFY - initialize hub]
```

### 🔄 Implementation Order

1. **Hub**: WebSocket hub with event loop
2. **Client**: Client read/write pumps
3. **Handler**: WebSocket upgrade handler
4. **Notifications**: Send real-time notifications
5. **Messages**: Define message types
6. **Monitoring**: WebSocket metrics

### ⚠️ Blockers to Watch For

- **Goroutine leaks**: Clean up on disconnect
- **Ping/pong**: Detect dead connections
- **Buffer size**: Client send buffer can fill up
- **Origin checking**: Validate WebSocket origin
- **Concurrent writes**: Only one goroutine writes to connection
- **Memory**: Limit max connections per user

### ✅ Definition of Done

- [ ] WebSocket connections working
- [ ] Hub manages clients correctly
- [ ] Real-time notifications sent on events
- [ ] Connection health monitored with ping/pong
- [ ] Disconnections handled gracefully
- [ ] WebSocket metrics tracked
- [ ] All tests passing

---

## Week 32: Comments + Likes

### 📋 Implementation Tasks

**Task 1: Create Database Migrations** (25 min)
- [ ] Create migration `migrations/010_create_comments_likes.up.sql`
- [ ] Add comments table
- [ ] Add likes table with UNIQUE(activity_id, user_id)
- [ ] Create indexes for performance
- [ ] Run migration

**Task 2: Create Models** (20 min)
- [ ] Create `internal/models/comment.go`
- [ ] Create `internal/models/like.go`
- [ ] Add JSON tags for API responses

**Task 3: Create Comment Repository** (60 min)
- [ ] Create `internal/repository/comment_repository.go`
- [ ] Implement `Create(ctx, comment) error`
- [ ] Implement `GetByActivityID(ctx, activityID) ([]*Comment, error)`
- [ ] Implement `GetByID(ctx, id) (*Comment, error)`
- [ ] Implement `Delete(ctx, id) error`
- [ ] Join with users table for author info

**Task 4: Create Like Repository** (60 min)
- [ ] Create `internal/repository/like_repository.go`
- [ ] Implement `Create(ctx, like) error`
- [ ] Implement `Delete(ctx, activityID, userID) error`
- [ ] Implement `GetByActivityID(ctx, activityID) ([]*Like, error)`
- [ ] Implement `HasLiked(ctx, activityID, userID) (bool, error)`
- [ ] Implement `GetCount(ctx, activityID) (int, error)`

**Task 5: Create Interaction Services** (120 min)
- [ ] Create `internal/services/comment_service.go`
- [ ] Implement `AddComment(ctx, activityID, userID, content) error`
  - Create comment
  - Send WebSocket notification
  - Create notification record
- [ ] Create `internal/services/like_service.go`
- [ ] Implement `Like(ctx, activityID, userID) error`
- [ ] Implement `Unlike(ctx, activityID, userID) error`
- [ ] Send real-time notifications via WebSocket hub

**Task 6: Create Interaction Handlers** (90 min)
- [ ] Create handlers for comments:
  - POST /activities/:id/comment
  - GET /activities/:id/comments
  - DELETE /activities/:id/comment/:commentId
- [ ] Create handlers for likes:
  - POST /activities/:id/like
  - DELETE /activities/:id/like
  - GET /activities/:id/likes
- [ ] Add authorization (only owner can delete comment)

**Task 7: Update Activity Feed** (45 min)
- [ ] Include like_count in feed query
- [ ] Include comment_count in feed query
- [ ] Add has_liked field (true if current user liked)
- [ ] Test feed shows engagement data

### 📦 Files You'll Create/Modify

```
migrations/
├── 010_create_comments_likes.up.sql [CREATE]
└── 010_create_comments_likes.down.sql [CREATE]

internal/
├── models/
│   ├── comment.go                 [CREATE]
│   └── like.go                    [CREATE]
├── repository/
│   ├── comment_repository.go      [CREATE]
│   └── like_repository.go         [CREATE]
├── services/
│   ├── comment_service.go         [CREATE]
│   └── like_service.go            [CREATE]
└── handlers/
    ├── comment_handler.go         [CREATE]
    └── like_handler.go            [CREATE]
```

### 🔄 Implementation Order

1. **Database**: Migrations → Models
2. **Repository**: Comment and like repositories
3. **Service**: Interaction services with notifications
4. **Handlers**: HTTP endpoints
5. **Integration**: WebSocket notifications
6. **Feed**: Update feed with engagement data

### ⚠️ Blockers to Watch For

- **Duplicate likes**: UNIQUE constraint prevents double-liking
- **Notification spam**: Don't notify on self-like
- **Authorization**: Only comment author can delete
- **WebSocket**: Ensure hub is passed to services
- **N+1 queries**: Join comments/likes efficiently in feed

### ✅ Definition of Done

- [ ] Can add/delete comments
- [ ] Can like/unlike activities
- [ ] Real-time notifications sent via WebSocket
- [ ] Feed shows like/comment counts
- [ ] Authorization working (only owner can delete)
- [ ] No duplicate likes (unique constraint)
- [ ] All tests passing

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
