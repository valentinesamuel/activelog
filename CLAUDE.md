# ActiveLog - 12-Month Go Learning Journey

## Project Overview

**Name:** ActiveLog  
**Tagline:** Track any activity, see everything in one place  
**Purpose:** A personal multi-sport activity tracker that serves as a vehicle for mastering Go over 12 months

### What is ActiveLog?

ActiveLog is a REST API and web application where users can log any type of physical activity (running, basketball, jump rope, walking, gym workouts, etc.) and view their complete activity history and analytics in one unified platform.

### Why This Project?

- **Learning Goal:** Master Go well enough to build secure, scalable enterprise systems
- **Career Goal:** Demonstrate production-ready Go skills for interviews and job opportunities
- **Personal Goal:** Build something useful that solves a real problem (tracking diverse activities in one place)
- **Timeline:** 12 months with 5-7 hours per week (primarily weekends)

---

## Learning Philosophy

This is a **project-based learning journey**, not a tutorial series. Every concept in Go will be learned by building real features that add value to ActiveLog. By Month 12, you'll have:

1. **Deep Go expertise** - idioms, patterns, concurrency, testing, performance
2. **A working product** - deployed, monitored, potentially monetizable
3. **Portfolio piece** - production-quality code to show in interviews
4. **Cloud/DevOps skills** - Docker, AWS, monitoring, CI/CD

---

## 12-Month Roadmap

### Phase 1: Foundation (Months 1-3) - "It Works"

**Goal:** Build a functional REST API with core CRUD operations

**What You'll Learn:**
- Go project structure and organization
- Basic syntax, types, and control flow
- HTTP servers and routing
- Database integration (PostgreSQL)
- Error handling the Go way
- Testing fundamentals
- JSON handling

**What You'll Build:**
- User registration and management
- Activity logging (create, read, update, delete)
- Activity retrieval and filtering
- Basic API documentation

**Deliverables:**
- Working REST API with core endpoints
- PostgreSQL database schema
- Unit tests for critical paths
- API documentation (README)

---

### Phase 2: Real App (Months 4-6) - "It's Useful"

**Goal:** Transform the API into something you'd actually use daily

**What You'll Learn:**
- Authentication & authorization (JWT)
- Middleware patterns
- Input validation and sanitization
- Database migrations
- Structured logging
- Configuration management
- More advanced testing (integration tests)

**What You'll Build:**
- User authentication (login, logout, token refresh)
- Activity types and categories
- Weekly/monthly activity summaries
- Streak tracking
- Personal records (PRs) tracking
- Search and filtering

**Deliverables:**
- Secured API with JWT authentication
- Rich activity analytics
- Comprehensive test suite
- Migration system

---

### Phase 3: Scale & Features (Months 7-9) - "It's Impressive"

**Goal:** Add advanced features that demonstrate Go's strengths

**What You'll Learn:**
- Goroutines and channels (concurrency)
- File uploads and storage (S3)
- Background job processing
- Caching strategies (Redis)
- Rate limiting
- Database query optimization
- Profiling and performance tuning

**What You'll Build:**
- Activity photo/video uploads
- Background jobs (weekly email summaries)
- Social features (achievements, sharing)
- Advanced analytics dashboard data
- Export functionality (CSV, PDF reports)
- API rate limiting

**Deliverables:**
- Concurrent processing for heavy operations
- S3 integration for media files
- Redis caching layer
- Performance benchmarks

---

### Phase 4: Production (Months 10-12) - "It's Real"

**Goal:** Deploy a production-ready system with monitoring and monetization

**What You'll Learn:**
- Docker containerization
- AWS deployment (ECS or EC2)
- CI/CD pipelines
- Monitoring and observability (Prometheus, Grafana)
- Log aggregation
- Database backups and recovery
- Security hardening
- Payment integration (Stripe)

**What You'll Build:**
- Dockerized application
- Deployed on AWS with proper networking
- CI/CD pipeline (GitHub Actions)
- Monitoring dashboards
- Automated backups
- Basic paid tier (premium analytics)

**Deliverables:**
- Production deployment on AWS
- Monitoring and alerting system
- Complete documentation
- Monetization ready

---

## Weekly Structure

### Time Commitment
- **7-10 hours per week** distributed strategically:
  - **Weekday evenings (2-3 hours total):** Short focused sessions - reading, small exercises, setup work
  - **Saturday (3-4 hours):** Main feature building and implementation
  - **Sunday (2-3 hours):** Testing + documentation + polish

### Weekly Flow Example
- **Monday evening (30-45 min):** Read about the week's Go concept, try small examples
- **Wednesday evening (30-45 min):** More practice, maybe start scaffolding the feature
- **Saturday (3-4 hours):** Build the actual feature
- **Sunday (2-3 hours):** Write tests, update docs, commit clean code

This rhythm ensures:
- Consistent exposure to Go (not 5 days without touching code)
- Momentum throughout the week
- Focused deep work on weekends
- Better retention through spaced repetition

### Weekly Deliverables
Each week will have:
1. **Learning Objective** - Specific Go concept or pattern to master
2. **Feature to Build** - Concrete functionality to add
3. **Testing Requirements** - What needs test coverage
4. **Documentation Updates** - What to document

### Accountability
- **Strict weekly deadlines** - Each week builds on the previous
- **No skipping ahead** - Master fundamentals before advanced topics
- **Working code only** - Every feature must work and be tested
- **Document as you go** - Future you will thank present you

---

## Technical Stack

### Core Technologies
- **Language:** Go 1.21+
- **Database:** PostgreSQL 15+
- **Cache:** Redis (Phase 3+)
- **Storage:** AWS S3 (Phase 3+)
- **Web Framework:** Standard library `net/http` + `gorilla/mux` or `chi`

### Development Tools
- **Version Control:** Git + GitHub
- **Database Migrations:** golang-migrate
- **Testing:** Go standard testing + testify
- **API Documentation:** OpenAPI/Swagger
- **Linting:** golangci-lint

### Production Infrastructure (Phase 4)
- **Containerization:** Docker
- **Orchestration:** AWS ECS (K8s later if desired)
- **CI/CD:** GitHub Actions
- **Monitoring:** Prometheus + Grafana
- **Logging:** structured logging with zap/zerolog

---

## Project Structure

```
activelog/
├── cmd/
│   └── api/              # Application entry point
├── internal/
│   ├── models/           # Data models
│   ├── handlers/         # HTTP handlers
│   ├── repository/       # Database layer
│   ├── service/          # Business logic
│   ├── middleware/       # HTTP middleware
│   └── config/           # Configuration
├── migrations/           # Database migrations
├── pkg/                  # Reusable packages
├── tests/                # Integration tests
├── docs/                 # Documentation
├── scripts/              # Utility scripts
├── .github/              # CI/CD workflows
├── Dockerfile
├── docker-compose.yml
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

---

## Success Metrics

### Technical Milestones
- [ ] Month 3: Working REST API with database
- [ ] Month 6: Authenticated API with analytics
- [ ] Month 9: Concurrent processing and caching
- [ ] Month 12: Production deployment with monitoring

### Learning Milestones
- [ ] Can structure a Go project properly
- [ ] Understand Go idioms and best practices
- [ ] Write idiomatic error handling
- [ ] Implement concurrent operations safely
- [ ] Deploy and monitor Go applications
- [ ] Debug and optimize Go code

### Career Milestones
- [ ] Portfolio project ready for interviews
- [ ] Can discuss Go architecture decisions confidently
- [ ] Understand trade-offs in Go system design
- [ ] Ready to contribute to Go codebases professionally

---

## Resources

### Official Documentation
- [Go Official Documentation](https://go.dev/doc/)
- [Go Tour](https://go.dev/tour/)
- [Effective Go](https://go.dev/doc/effective_go)

### Books (as needed)
- "Learning Go" by Jon Bodner
- "Let's Go" by Alex Edwards (web development)
- "Concurrency in Go" by Katherine Cox-Buday

### Community
- [r/golang](https://reddit.com/r/golang)
- [Gophers Slack](https://gophers.slack.com)
- Go Time Podcast

---

## Notes and Reflections

### Current Status
- **Start Date:** December 24, 2024
- **Current Phase:** Planning
- **Weeks Completed:** 0/52

### Key Learnings
(To be updated weekly with insights, challenges, breakthroughs)

### Challenges Overcome
(Document difficulties and how they were solved)

### What's Working
(Track what study/practice methods are most effective)

### What Needs Adjustment
(Honest assessment of what's not working)

---

## Motivation

### Why Go Matters
- Financial systems are built with Go
- Cloud infrastructure requires Go expertise
- High demand in job market
- Performance + simplicity for enterprise systems

### Why This Approach Works
- Project-based learning > passive tutorials
- Building something useful = sustained motivation
- Strict deadlines = consistent progress
- Real deployments = interview-ready experience

### Remember When Struggling
- You're a backend team lead - you understand complex systems
- JavaScript mastery transfers - you know the concepts
- One year from now, you'll look back at this moment
- Struggling doesn't mean dumb - it means learning

---

## Contact & Support

**Mentor:** Claude (AI Assistant)  
**Your Commitment:** 7-10 hours/week for 52 weeks  
**Your Goal:** Build secure enterprise Go systems with confidence

---

*"The expert in anything was once a beginner."*

Let's build something remarkable.