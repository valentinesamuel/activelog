package featureflags

import "os"

// FeatureFlags holds all feature toggle states loaded from environment variables
type FeatureFlags struct {
	EnableComments bool
	EnableLikes    bool
	EnableFriends  bool
	EnableWebhooks bool
	EnableFeed     bool
}

// Load reads feature flags from environment variables.
// A flag is enabled when its env var is set to "enabled".
func Load() *FeatureFlags {
	return &FeatureFlags{
		EnableComments: isEnabled("FEATURE_COMMENTS"),
		EnableLikes:    isEnabled("FEATURE_LIKES"),
		EnableFriends:  isEnabled("FEATURE_FRIENDS"),
		EnableWebhooks: isEnabled("FEATURE_WEBHOOKS"),
		EnableFeed:     isEnabled("FEATURE_FEED"),
	}
}

// IsEnabled checks whether a named feature is on.
// Supported feature names: "comments", "likes", "friends", "webhooks", "feed".
func (f *FeatureFlags) IsEnabled(feature string) bool {
	switch feature {
	case "comments":
		return f.EnableComments
	case "likes":
		return f.EnableLikes
	case "friends":
		return f.EnableFriends
	case "webhooks":
		return f.EnableWebhooks
	case "feed":
		return f.EnableFeed
	default:
		return false
	}
}

func isEnabled(env string) bool {
	return os.Getenv(env) == "enabled"
}
