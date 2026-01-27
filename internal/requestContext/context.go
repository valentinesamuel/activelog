package requestcontext

import "context"

type key int

var userKey key

type User struct {
	Id    int    `json:"user_id"`
	Email string `json:"email"`
}

func NewContext(ctx context.Context, u *User) context.Context {
	return context.WithValue(ctx, userKey, u)
}

// FromContext returns the User value stored in ctx, if any.
func FromContext(ctx context.Context) (*User, bool) {
	u, ok := ctx.Value(userKey).(*User)
	return u, ok
}
