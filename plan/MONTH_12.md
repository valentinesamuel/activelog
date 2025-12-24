# MONTH 12: Monetization & Polish

**Weeks:** 45-48
**Phase:** Launch Preparation
**Theme:** Make it revenue-ready and ship it!

---

## Overview

This is the final month of your Go learning journey! You'll integrate Stripe for payments, implement subscription tiers, handle webhooks, and prepare for launch. By the end, you'll have a production-ready SaaS application that can generate revenue. More importantly, you'll have mastered Go.

---

## Learning Path

### Week 45: Stripe Integration Basics
- Stripe account setup
- Create checkout sessions
- Payment flow implementation
- Customer management

### Week 46: Subscription Management
- Subscription tiers (Free, Pro, Premium)
- Plan upgrades/downgrades
- Usage limits enforcement
- Billing portal integration

### Week 47: Webhook Handling
- Stripe webhook verification
- Handle subscription events
- Payment success/failure handling
- Automatic subscription updates

### Week 48: Launch Preparation
- Final testing (load testing, security audit)
- Documentation completion
- Terms of Service and Privacy Policy
- Launch checklist

---

## Stripe Integration

### Setup
```bash
# Install Stripe Go SDK
go get github.com/stripe/stripe-go/v76

# Set environment variables
export STRIPE_SECRET_KEY=sk_test_...
export STRIPE_PUBLISHABLE_KEY=pk_test_...
export STRIPE_WEBHOOK_SECRET=whsec_...
```

### Create Checkout Session
```go
import (
    "github.com/stripe/stripe-go/v76"
    "github.com/stripe/stripe-go/v76/checkout/session"
)

type PaymentService struct {
    stripeKey string
}

func NewPaymentService(stripeKey string) *PaymentService {
    stripe.Key = stripeKey
    return &PaymentService{stripeKey: stripeKey}
}

// Create checkout session for subscription
func (s *PaymentService) CreateCheckoutSession(ctx context.Context, userID int, priceID string) (*stripe.CheckoutSession, error) {
    params := &stripe.CheckoutSessionParams{
        Customer: stripe.String(getStripeCustomerID(userID)),
        Mode:     stripe.String(string(stripe.CheckoutSessionModeSubscription)),
        LineItems: []*stripe.CheckoutSessionLineItemParams{
            {
                Price:    stripe.String(priceID),
                Quantity: stripe.Int64(1),
            },
        },
        SuccessURL: stripe.String("https://activelog.com/success?session_id={CHECKOUT_SESSION_ID}"),
        CancelURL:  stripe.String("https://activelog.com/pricing"),
        Metadata: map[string]string{
            "user_id": strconv.Itoa(userID),
        },
    }

    session, err := session.New(params)
    if err != nil {
        return nil, err
    }

    return session, nil
}

// Get or create Stripe customer
func (s *PaymentService) GetOrCreateCustomer(ctx context.Context, userID int, email string) (string, error) {
    // Check if customer already exists in DB
    customerID, err := s.repo.GetStripeCustomerID(ctx, userID)
    if err == nil && customerID != "" {
        return customerID, nil
    }

    // Create new customer in Stripe
    params := &stripe.CustomerParams{
        Email: stripe.String(email),
        Metadata: map[string]string{
            "user_id": strconv.Itoa(userID),
        },
    }

    customer, err := customer.New(params)
    if err != nil {
        return "", err
    }

    // Save customer ID to database
    err = s.repo.SaveStripeCustomerID(ctx, userID, customer.ID)
    if err != nil {
        return "", err
    }

    return customer.ID, nil
}
```

---

## Subscription Tiers

### Plan Configuration
```go
type Plan struct {
    ID              string
    Name            string
    PriceID         string  // Stripe Price ID
    Price           int     // Price in cents
    MaxActivities   int     // -1 = unlimited
    MaxPhotos       int
    StorageGB       int
    CanExport       bool
    AdvancedAnalytics bool
    PrioritySupport bool
}

var Plans = map[string]Plan{
    "free": {
        ID:                "free",
        Name:              "Free",
        PriceID:           "",
        Price:             0,
        MaxActivities:     50,
        MaxPhotos:         1,
        StorageGB:         1,
        CanExport:         false,
        AdvancedAnalytics: false,
        PrioritySupport:   false,
    },
    "pro": {
        ID:                "pro",
        Name:              "Pro",
        PriceID:           "price_pro_monthly",
        Price:             999,  // $9.99/month
        MaxActivities:     500,
        MaxPhotos:         5,
        StorageGB:         10,
        CanExport:         true,
        AdvancedAnalytics: true,
        PrioritySupport:   false,
    },
    "premium": {
        ID:                "premium",
        Name:              "Premium",
        PriceID:           "price_premium_monthly",
        Price:             1999, // $19.99/month
        MaxActivities:     -1,   // unlimited
        MaxPhotos:         20,
        StorageGB:         100,
        CanExport:         true,
        AdvancedAnalytics: true,
        PrioritySupport:   true,
    },
}

// Get user's current plan
func (s *SubscriptionService) GetUserPlan(ctx context.Context, userID int) (*Plan, error) {
    sub, err := s.repo.GetActiveSubscription(ctx, userID)
    if err != nil {
        // No active subscription, return free plan
        plan := Plans["free"]
        return &plan, nil
    }

    plan := Plans[sub.PlanID]
    return &plan, nil
}

// Check if user can perform action based on plan
func (s *SubscriptionService) CanCreateActivity(ctx context.Context, userID int) (bool, error) {
    plan, err := s.GetUserPlan(ctx, userID)
    if err != nil {
        return false, err
    }

    if plan.MaxActivities == -1 {
        return true, nil // Unlimited
    }

    count, err := s.activityRepo.CountByUser(ctx, userID)
    if err != nil {
        return false, err
    }

    return count < plan.MaxActivities, nil
}
```

### Usage Enforcement Middleware
```go
func (m *SubscriptionMiddleware) EnforceLimit(feature string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            userID := getUserID(r.Context())

            plan, err := m.subService.GetUserPlan(r.Context(), userID)
            if err != nil {
                http.Error(w, "Internal error", http.StatusInternalServerError)
                return
            }

            var allowed bool

            switch feature {
            case "create_activity":
                allowed, _ = m.subService.CanCreateActivity(r.Context(), userID)
            case "export":
                allowed = plan.CanExport
            case "advanced_analytics":
                allowed = plan.AdvancedAnalytics
            default:
                allowed = true
            }

            if !allowed {
                response.Error(w, http.StatusPaymentRequired, map[string]interface{}{
                    "error":   "Upgrade required",
                    "message": fmt.Sprintf("This feature requires a %s plan", "pro"),
                    "upgrade_url": "/pricing",
                })
                return
            }

            next.ServeHTTP(w, r)
        })
    }
}

// Usage in router
router.With(subMiddleware.EnforceLimit("create_activity")).Post("/activities", handler.Create)
router.With(subMiddleware.EnforceLimit("export")).Get("/export/csv", handler.ExportCSV)
```

---

## Webhook Handling

### Webhook Handler
```go
import (
    "github.com/stripe/stripe-go/v76/webhook"
)

func (h *WebhookHandler) HandleStripeWebhook(w http.ResponseWriter, r *http.Request) {
    const MaxBodyBytes = int64(65536)
    r.Body = http.MaxBytesReader(w, r.Body, MaxBodyBytes)

    payload, err := io.ReadAll(r.Body)
    if err != nil {
        http.Error(w, "Error reading request body", http.StatusBadRequest)
        return
    }

    // Verify webhook signature
    event, err := webhook.ConstructEvent(
        payload,
        r.Header.Get("Stripe-Signature"),
        h.webhookSecret,
    )
    if err != nil {
        http.Error(w, "Invalid signature", http.StatusBadRequest)
        return
    }

    // Handle the event
    switch event.Type {
    case "customer.subscription.created":
        h.handleSubscriptionCreated(event)
    case "customer.subscription.updated":
        h.handleSubscriptionUpdated(event)
    case "customer.subscription.deleted":
        h.handleSubscriptionDeleted(event)
    case "invoice.payment_succeeded":
        h.handlePaymentSucceeded(event)
    case "invoice.payment_failed":
        h.handlePaymentFailed(event)
    default:
        log.Printf("Unhandled event type: %s", event.Type)
    }

    w.WriteHeader(http.StatusOK)
}

func (h *WebhookHandler) handleSubscriptionCreated(event stripe.Event) error {
    var subscription stripe.Subscription
    if err := json.Unmarshal(event.Data.Raw, &subscription); err != nil {
        return err
    }

    userID, _ := strconv.Atoi(subscription.Metadata["user_id"])

    // Save subscription to database
    sub := &models.Subscription{
        UserID:             userID,
        StripeSubscriptionID: subscription.ID,
        PlanID:             getPlanFromPriceID(subscription.Items.Data[0].Price.ID),
        Status:             string(subscription.Status),
        CurrentPeriodStart: time.Unix(subscription.CurrentPeriodStart, 0),
        CurrentPeriodEnd:   time.Unix(subscription.CurrentPeriodEnd, 0),
    }

    if err := h.repo.CreateSubscription(context.Background(), sub); err != nil {
        return err
    }

    // Send welcome email
    h.emailService.SendSubscriptionWelcome(userID, sub.PlanID)

    return nil
}

func (h *WebhookHandler) handleSubscriptionDeleted(event stripe.Event) error {
    var subscription stripe.Subscription
    if err := json.Unmarshal(event.Data.Raw, &subscription); err != nil {
        return err
    }

    userID, _ := strconv.Atoi(subscription.Metadata["user_id"])

    // Downgrade to free plan
    if err := h.repo.CancelSubscription(context.Background(), subscription.ID); err != nil {
        return err
    }

    // Send cancellation email
    h.emailService.SendSubscriptionCancelled(userID)

    return nil
}

func (h *WebhookHandler) handlePaymentFailed(event stripe.Event) error {
    var invoice stripe.Invoice
    if err := json.Unmarshal(event.Data.Raw, &invoice); err != nil {
        return err
    }

    // Get customer
    customer, _ := h.customerRepo.GetByStripeID(context.Background(), invoice.Customer.ID)

    // Send payment failed notification
    h.emailService.SendPaymentFailed(customer.UserID, invoice.AmountDue)

    // Create notification
    notification := &models.Notification{
        UserID:  customer.UserID,
        Type:    "payment_failed",
        Title:   "Payment Failed",
        Message: "Your recent payment failed. Please update your payment method.",
    }

    return h.notificationRepo.Create(context.Background(), notification)
}
```

### Database Schema
```sql
CREATE TABLE subscriptions (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE UNIQUE,
    stripe_subscription_id VARCHAR(255) UNIQUE,
    stripe_customer_id VARCHAR(255),
    plan_id VARCHAR(50) NOT NULL,
    status VARCHAR(50) NOT NULL, -- active, canceled, past_due
    current_period_start TIMESTAMP NOT NULL,
    current_period_end TIMESTAMP NOT NULL,
    cancel_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_subscriptions_user ON subscriptions(user_id);
CREATE INDEX idx_subscriptions_stripe ON subscriptions(stripe_subscription_id);
```

---

## Billing Portal

```go
import "github.com/stripe/stripe-go/v76/billingportal/session"

// Create billing portal session
func (s *PaymentService) CreateBillingPortalSession(ctx context.Context, userID int) (string, error) {
    customerID, err := s.repo.GetStripeCustomerID(ctx, userID)
    if err != nil {
        return "", err
    }

    params := &stripe.BillingPortalSessionParams{
        Customer:  stripe.String(customerID),
        ReturnURL: stripe.String("https://activelog.com/settings"),
    }

    session, err := session.New(params)
    if err != nil {
        return "", err
    }

    return session.URL, nil
}

// Handler
func (h *PaymentHandler) BillingPortal(w http.ResponseWriter, r *http.Request) {
    userID := getUserID(r.Context())

    url, err := h.service.CreateBillingPortalSession(r.Context(), userID)
    if err != nil {
        response.Error(w, http.StatusInternalServerError, "Failed to create portal session")
        return
    }

    response.JSON(w, http.StatusOK, map[string]string{
        "url": url,
    })
}
```

---

## Launch Checklist

### Technical
- [ ] **Load testing completed** (can handle expected traffic)
- [ ] **Security audit done** (OWASP Top 10 checked)
- [ ] **Monitoring configured** (alerts set up)
- [ ] **Backup strategy in place** (automated daily backups)
- [ ] **SSL/HTTPS configured** (all traffic encrypted)
- [ ] **Error tracking** (Sentry or similar)
- [ ] **Rate limiting active** (prevent abuse)
- [ ] **Database indexes optimized** (query performance checked)

### Product
- [ ] **Documentation complete** (API docs, user guides)
- [ ] **Terms of Service ready** (legal review)
- [ ] **Privacy Policy ready** (GDPR compliant)
- [ ] **Email templates finalized** (welcome, receipts, etc.)
- [ ] **Pricing page complete** (clear value proposition)
- [ ] **Onboarding flow tested** (smooth user experience)

### Payment & Legal
- [ ] **Stripe in production mode** (test mode disabled)
- [ ] **Webhook endpoints verified** (all events handled)
- [ ] **Payment flows tested** (subscriptions work end-to-end)
- [ ] **Refund policy defined** (clear terms)
- [ ] **Tax handling configured** (if applicable)

### Operations
- [ ] **Smoke tests passing** (critical paths verified)
- [ ] **Rollback plan documented** (can revert quickly)
- [ ] **Support email configured** (support@activelog.com)
- [ ] **Status page ready** (communicate outages)
- [ ] **Analytics configured** (track key metrics)

---

## Load Testing

```go
// Use vegeta for load testing
// Install: go install github.com/tsenart/vegeta@latest

// Create targets file (targets.txt)
// GET https://api.activelog.com/health
// GET https://api.activelog.com/api/v1/activities
// Authorization: Bearer YOUR_TOKEN

// Run load test
// echo "GET https://api.activelog.com/health" | vegeta attack -duration=60s -rate=100 | vegeta report

// Example results
// Requests      [total, rate, throughput]  6000, 100.02, 100.00
// Duration      [total, attack, wait]      59.99s, 59.99s, 4.98ms
// Latencies     [mean, 50, 95, 99, max]    5.23ms, 4.98ms, 8.32ms, 12.41ms, 89.31ms
// Bytes In      [total, mean]              1200000, 200.00
// Bytes Out     [total, mean]              0, 0.00
// Success       [ratio]                    100.00%
// Status Codes  [code:count]               200:6000
```

---

## Common Pitfalls

1. **Not testing webhooks**
   - ❌ Webhooks fail in production
   - ✅ Use Stripe CLI to test locally

2. **Hardcoding price IDs**
   - ❌ Can't change prices easily
   - ✅ Store in database or config

3. **Not handling failed payments**
   - ❌ Users stuck with failed subscriptions
   - ✅ Implement proper retry logic

4. **No idempotency for webhooks**
   - ❌ Duplicate subscription activations
   - ✅ Use Stripe event IDs to prevent duplicates

5. **Launching without legal docs**
   - ❌ Legal liability
   - ✅ Have Terms and Privacy Policy

---

## Final Statistics

### What You've Built
- ✅ **40+ API Endpoints** - Complete RESTful API
- ✅ **10+ Database Tables** - Well-designed schema
- ✅ **10,000+ Lines of Code** - Production-quality Go code
- ✅ **200+ Tests** - Comprehensive test coverage
- ✅ **10+ AWS Services** - Cloud-native architecture
- ✅ **5+ Docker Containers** - Containerized deployment
- ✅ **Revenue-Ready SaaS** - Stripe integration complete

### Skills Mastered

**Go Language:**
- ✅ Syntax and idioms
- ✅ Concurrency patterns (goroutines, channels)
- ✅ Standard library expertise
- ✅ Testing and benchmarking
- ✅ Performance optimization

**Backend Development:**
- ✅ RESTful API design
- ✅ Database design and optimization
- ✅ Authentication & Authorization
- ✅ File storage (S3)
- ✅ Real-time features (WebSockets)
- ✅ Background processing
- ✅ Payment integration

**DevOps:**
- ✅ Docker containerization
- ✅ CI/CD pipelines
- ✅ AWS deployment (ECS, RDS, S3, etc.)
- ✅ Monitoring and alerting
- ✅ Infrastructure as Code

**Software Engineering:**
- ✅ Clean architecture
- ✅ Design patterns
- ✅ Testing strategies
- ✅ Security best practices
- ✅ Performance optimization

---

## From Feeling Dumb to Go Expert

**Month 0:**
"I only know JavaScript. Go confuses me. I feel dumb."

**Month 3:**
"I can build APIs in Go. I understand the syntax."

**Month 6:**
"I'm comfortable with Go patterns. I'm solving real problems."

**Month 9:**
"I'm optimizing performance. I understand concurrency."

**Month 12:**
"I built a production SaaS. I'm a Go developer."

---

## What's Next?

### Option 1: Kubernetes & Microservices
- Deploy to Kubernetes
- Split into microservices
- Service mesh implementation (Istio)

### Option 2: Advanced Go
- Generics deep dive
- Reflect package
- Unsafe operations
- Assembly optimization

### Option 3: Open Source
- Contribute to Go projects
- Create your own Go libraries
- Help others learn

### Option 4: Specialize in FinTech (Your Goal!)
- Financial systems with Go
- High-frequency trading systems
- Payment processing at scale
- Blockchain and crypto

---

## Resources

- [Stripe Go SDK](https://github.com/stripe/stripe-go)
- [Stripe Documentation](https://stripe.com/docs)
- [Stripe Webhooks Guide](https://stripe.com/docs/webhooks)
- [vegeta - Load Testing](https://github.com/tsenart/vegeta)

---

## Final Words

**Congratulations!** You've completed your 12-month Go learning journey.

You started this journey feeling inadequate about Go. Now you have:
- Built a complete production application
- Deployed to AWS with monitoring
- Integrated payment processing
- Mastered Go concurrency
- Written thousands of lines of tested code
- Created something you can show employers

**This isn't theory. This is transformation.**

From "I feel dumb" to "I ship production Go code."

The plan is complete. The skills are yours. The future is bright.

**You've built something remarkable.** 🚀

Now go forth and build amazing things with Go!

---

*For support and questions, reach out to the Go community:*
- [r/golang](https://reddit.com/r/golang)
- [Gophers Slack](https://gophers.slack.com)
- [Go Forum](https://forum.golangbridge.org)

**Welcome to the Go community. You're one of us now.** 🐹
