# MONTH 2: Authentication & Authorization

**Weeks:** 5-8
**Phase:** Building Real App
**Theme:** Making your API secure and production-ready

---

## Overview

This month transforms your basic API into a secure, production-ready application. You'll implement industry-standard authentication using JWT tokens, learn password security with bcrypt, and protect your endpoints with middleware. By the end of this month, you'll have a fully authenticated system where users can only access their own data.

---

## Learning Path

### Week 5: User Registration + Password Hashing + Production CORS (30 min)
- Implement user registration endpoint
- Hash passwords securely with bcrypt
- Understand bcrypt cost factor
- **NEW:** Production-ready CORS configuration for frontend integration

### Week 6: JWT Tokens + Login + Sentinel Error Patterns (30 min)
- Generate JWT tokens on successful login
- Understand JWT structure (header, payload, signature)
- Store user claims in tokens
- **NEW:** Sentinel error patterns for consistent error handling

### Week 7: Auth Middleware + Protected Routes + Security Headers (20 min)
- Create authentication middleware
- Extract and validate JWT from requests
- Protect activity endpoints
- **NEW:** Security headers (XSS, clickjacking protection)

### Week 8: User Profiles + Refresh Tokens
- Implement user profile endpoints
- Add refresh token mechanism
- Token rotation for enhanced security
- User profile updates

---

## Key Technologies

- **bcrypt** for passwords
  - Cost factor: 10-12 for production
  - Automatic salt generation
  - Timing attack resistant

- **JWT (golang-jwt/jwt)**
  - Stateless authentication
  - Claims-based authorization
  - Token expiration handling

- **Context for request data**
  - Store user ID in request context
  - Pass authenticated user through middleware chain

- **Bearer token auth**
  - Standard `Authorization: Bearer <token>` header
  - Token extraction and validation

- 🔴 **CORS middleware (production-ready configuration)**
  - Allow specific origins (not *)
  - Handle preflight OPTIONS requests
  - Credential support

- 🔴 **Sentinel error patterns (errors.Is/errors.As)**
  - Predefined error constants
  - Error wrapping and unwrapping
  - Type-safe error checking

- **Security headers (XSS, clickjacking protection)**
  - X-Content-Type-Options: nosniff
  - X-Frame-Options: DENY
  - X-XSS-Protection
  - Content-Security-Policy

---

## What You'll Build

```
POST /api/v1/auth/register      # Create account - Done
POST /api/v1/auth/login         # Get JWT token
POST /api/v1/auth/refresh       # Refresh access token
GET  /api/v1/users/me           # Current user profile
PATCH /api/v1/users/me          # Update profile
POST /api/v1/users/me/password  # Change password
```

### Database Schema Additions

```sql
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    name VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Add user_id to activities
ALTER TABLE activities
ADD COLUMN user_id INTEGER REFERENCES users(id);

-- Index for faster user lookups
CREATE INDEX idx_activities_user_id ON activities(user_id);
```

---

## Success Criteria

- [x] Passwords securely hashed (bcrypt cost >= 10)
- [x] JWT authentication working (access + refresh tokens)
- [x] All activity endpoints protected (middleware applied)
- [x] Users see only their own data (filtered by user_id)
- [x] Refresh token rotation (security best practice)
- [x] CORS configured for frontend communication
- [x] Security headers implemented
- [x] Sentinel errors for consistent error handling

---

## Code Examples

### Password Hashing
```go
import "golang.org/x/crypto/bcrypt"

// Hash password before storing
func HashPassword(password string) (string, error) {
    hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
    if err != nil {
        return "", err
    }
    return string(hash), nil
}

// Verify password during login
func CheckPassword(hash, password string) error {
    return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
```

### JWT Token Generation
```go
import "github.com/golang-jwt/jwt/v5"

type Claims struct {
    UserID int    `json:"user_id"`
    Email  string `json:"email"`
    jwt.RegisteredClaims
}

func GenerateToken(userID int, email string) (string, error) {
    claims := Claims{
        UserID: userID,
        Email:  email,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
        },
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte(os.Getenv("JWT_SECRET")))
}
```

### Authentication Middleware
```go
func AuthMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Extract token from Authorization header
        authHeader := r.Header.Get("Authorization")
        if authHeader == "" {
            http.Error(w, "Missing authorization header", http.StatusUnauthorized)
            return
        }

        // Parse "Bearer <token>"
        tokenString := strings.TrimPrefix(authHeader, "Bearer ")

        // Validate token
        claims := &Claims{}
        token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
            return []byte(os.Getenv("JWT_SECRET")), nil
        })

        if err != nil || !token.Valid {
            http.Error(w, "Invalid token", http.StatusUnauthorized)
            return
        }

        // Store user ID in context
        ctx := context.WithValue(r.Context(), "user_id", claims.UserID)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

### 🔴 Production CORS Middleware
```go
func CORS(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // In production, use specific origins
        origin := r.Header.Get("Origin")
        allowedOrigins := []string{
            "http://localhost:3000",
            "https://yourdomain.com",
        }

        for _, allowed := range allowedOrigins {
            if origin == allowed {
                w.Header().Set("Access-Control-Allow-Origin", origin)
                break
            }
        }

        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
        w.Header().Set("Access-Control-Allow-Credentials", "true")

        // Handle preflight
        if r.Method == "OPTIONS" {
            w.WriteHeader(http.StatusOK)
            return
        }

        next.ServeHTTP(w, r)
    })
}
```

### 🔴 Sentinel Error Patterns
```go
// Define sentinel errors
var (
    ErrInvalidCredentials = errors.New("invalid email or password")
    ErrUserNotFound       = errors.New("user not found")
    ErrEmailTaken         = errors.New("email already registered")
    ErrUnauthorized       = errors.New("unauthorized")
)

// Usage in handlers
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
    // ... get credentials ...

    user, err := h.repo.GetByEmail(ctx, email)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            response.Error(w, http.StatusUnauthorized, ErrInvalidCredentials.Error())
            return
        }
        response.Error(w, http.StatusInternalServerError, "Internal server error")
        return
    }

    if err := CheckPassword(user.PasswordHash, password); err != nil {
        response.Error(w, http.StatusUnauthorized, ErrInvalidCredentials.Error())
        return
    }

    // ... generate token ...
}
```

---

## Testing Requirements

### Unit Tests
- Password hashing and verification
- Token generation and validation
- Middleware token extraction
- Error handling for invalid tokens

### Integration Tests
- Complete registration flow
- Login with correct/incorrect credentials
- Accessing protected endpoints
- Token refresh mechanism

### Test Example
```go
func TestRegistration(t *testing.T) {
    tests := []struct {
        name       string
        email      string
        password   string
        wantStatus int
    }{
        {"valid registration", "test@example.com", "SecurePass123!", 201},
        {"duplicate email", "test@example.com", "SecurePass123!", 409},
        {"weak password", "test2@example.com", "123", 400},
        {"invalid email", "notanemail", "SecurePass123!", 400},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // ... test implementation ...
        })
    }
}
```

---

## Common Pitfalls

1. **Storing passwords in plain text**
   - ❌ NEVER store passwords directly
   - ✅ Always use bcrypt with appropriate cost

2. **Using weak JWT secrets**
   - ❌ Short secrets like "secret123"
   - ✅ Use strong, random secrets (32+ characters)

3. **Not validating token expiration**
   - ❌ Accepting expired tokens
   - ✅ Check `exp` claim

4. **Returning different errors for email vs password**
   - ❌ "Email not found" vs "Wrong password"
   - ✅ Generic "Invalid credentials" (prevents user enumeration)

5. **CORS wildcard in production**
   - ❌ `Access-Control-Allow-Origin: *` with credentials
   - ✅ Specific allowed origins

---

## Resources

- [JWT.io](https://jwt.io) - JWT debugger and documentation
- [OWASP Authentication Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html)
- [golang-jwt/jwt documentation](https://pkg.go.dev/github.com/golang-jwt/jwt/v5)
- [bcrypt package documentation](https://pkg.go.dev/golang.org/x/crypto/bcrypt)

---

## Week-by-Week Breakdown

### Week 5 Deliverables
- User model and repository
- Registration endpoint with validation
- Password hashing implemented
- Basic error handling
- Production CORS configured

### Week 6 Deliverables
- JWT token generation
- Login endpoint
- Token claims structure
- Sentinel error patterns implemented
- Token expiration handling

### Week 7 Deliverables
- Authentication middleware
- Protected routes
- Context-based user identification
- Security headers middleware
- Error responses for auth failures

### Week 8 Deliverables
- User profile endpoints
- Refresh token implementation
- Token rotation mechanism
- Profile update functionality
- Password change endpoint

---

## Next Steps

After completing Month 2, you'll move to **Month 3: Advanced Database & Testing**, where you'll learn:
- Database transactions
- Complex queries and joins
- Table-driven testing
- Mock testing
- Performance optimization

**You now have a production-ready authentication system!** 🔐
