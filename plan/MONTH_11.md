# MONTH 11: AWS Deployment & Monitoring

**Weeks:** 41-44
**Phase:** Production Deployment
**Theme:** Deploy to the cloud with confidence

---

## Overview

This month brings your application to production on AWS. You'll deploy to ECS Fargate, configure HTTPS/TLS, set up comprehensive monitoring with Prometheus and Grafana, and implement distributed tracing. By the end, you'll have a production-ready, monitored application running in the cloud.

---

## API Endpoints Reference

**Note:** Month 11 focuses on AWS deployment and monitoring infrastructure. The API endpoints remain the same but are now:
- Deployed on AWS ECS Fargate (production environment)
- Accessible via HTTPS with SSL/TLS certificates
- Load balanced with Application Load Balancer
- Monitored with Prometheus and Grafana dashboards
- Traced with OpenTelemetry distributed tracing

### Production URLs:
- **Development**: `http://localhost:8080`
- **Production**: `https://api.activelog.com`

### Monitoring Dashboards:
- **Grafana**: `https://monitoring.activelog.com`
- **Prometheus**: `https://prometheus.activelog.com` (internal only)

### Enhanced Health Check (Production):
- **HTTP Method:** `GET`
- **URL:** `https://api.activelog.com/health`
- **Success Response (200 OK):**
  ```json
  {
    "status": "healthy",
    "environment": "production",
    "version": "1.0.0",
    "services": {
      "database": {
        "status": "up",
        "latency_ms": 2.5,
        "connection_pool": {
          "active": 5,
          "idle": 15,
          "max": 20
        }
      },
      "redis": {
        "status": "up",
        "latency_ms": 0.8
      },
      "s3": {
        "status": "up",
        "latency_ms": 45.2
      }
    },
    "instance_id": "i-0123456789abcdef0",
    "region": "us-east-1",
    "uptime_seconds": 864000
  }
  ```

---

## Learning Path

### Week 41: AWS ECS Deployment Setup
- ECS concepts (clusters, services, tasks)
- Task definitions
- Load balancer configuration
- RDS and ElastiCache setup

### Week 42: HTTPS/TLS Configuration with Certificate Manager (45 min)
- Request SSL/TLS certificates
- Configure ALB with HTTPS
- Redirect HTTP to HTTPS
- Certificate auto-renewal

### Week 43: Monitoring and Logging + Distributed Tracing Basics (60 min)
- CloudWatch Logs integration
- Prometheus metrics collection
- Grafana dashboard setup
- **NEW:** OpenTelemetry distributed tracing

### Week 44: Production Hardening and Security
- Security groups configuration
- Secrets Manager for credentials
- Auto-scaling policies
- Backup strategies

---

## AWS Architecture

```
Internet
   ↓
Route 53 (DNS: activelog.com)
   ↓
ALB (HTTPS) - Certificate Manager
   ↓
ECS Fargate (Auto-scaling)
   ├─→ RDS PostgreSQL (Multi-AZ)
   ├─→ ElastiCache Redis (Cluster Mode)
   ├─→ S3 (File Storage)
   └─→ CloudWatch (Logs & Metrics)
```

---

## Services Used

- **ECS Fargate** (container orchestration)
  - Serverless container runtime
  - No EC2 management
  - Pay per vCPU and memory

- **RDS PostgreSQL** (managed database)
  - Automated backups
  - Multi-AZ for high availability
  - Automatic failover

- **ElastiCache Redis** (managed cache)
  - Cluster mode for scalability
  - Automatic failover
  - Backup and restore

- **S3** (file storage)
  - Activity photos
  - Report exports
  - Static assets

- **Application Load Balancer**
  - HTTPS termination
  - Health checks
  - Traffic distribution

- **Route 53** (DNS)
  - Domain management
  - Health checks
  - Traffic routing

- 🔴 **Certificate Manager** (SSL/TLS certificates)
  - Free SSL certificates
  - Automatic renewal
  - Wildcard certificates

- **CloudWatch** (logs/metrics)
  - Application logs
  - Infrastructure metrics
  - Alarms and notifications

- **Secrets Manager** (credentials)
  - Database passwords
  - API keys
  - Automatic rotation

- 🔴 **X-Ray** (distributed tracing - optional)
  - Request tracing
  - Service maps
  - Performance insights

---

## ECS Deployment

### Task Definition (task-definition.json)
```json
{
  "family": "activelog-api",
  "networkMode": "awsvpc",
  "requiresCompatibilities": ["FARGATE"],
  "cpu": "256",
  "memory": "512",
  "executionRoleArn": "arn:aws:iam::ACCOUNT:role/ecsTaskExecutionRole",
  "taskRoleArn": "arn:aws:iam::ACCOUNT:role/ecsTaskRole",
  "containerDefinitions": [
    {
      "name": "activelog-api",
      "image": "username/activelog:latest",
      "portMappings": [
        {
          "containerPort": 8080,
          "protocol": "tcp"
        }
      ],
      "environment": [
        {
          "name": "PORT",
          "value": "8080"
        }
      ],
      "secrets": [
        {
          "name": "DATABASE_URL",
          "valueFrom": "arn:aws:secretsmanager:region:account:secret:activelog/db-url"
        },
        {
          "name": "REDIS_URL",
          "valueFrom": "arn:aws:secretsmanager:region:account:secret:activelog/redis-url"
        },
        {
          "name": "JWT_SECRET",
          "valueFrom": "arn:aws:secretsmanager:region:account:secret:activelog/jwt-secret"
        }
      ],
      "logConfiguration": {
        "logDriver": "awslogs",
        "options": {
          "awslogs-group": "/ecs/activelog-api",
          "awslogs-region": "us-east-1",
          "awslogs-stream-prefix": "ecs"
        }
      },
      "healthCheck": {
        "command": ["CMD-SHELL", "wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1"],
        "interval": 30,
        "timeout": 5,
        "retries": 3,
        "startPeriod": 60
      }
    }
  ]
}
```

### Deploy Script
```bash
#!/bin/bash
set -e

# Variables
REGION="us-east-1"
CLUSTER="activelog-cluster"
SERVICE="activelog-service"
IMAGE_TAG="${1:-latest}"

echo "Deploying version $IMAGE_TAG to ECS..."

# Build and push Docker image
docker build -t activelog:$IMAGE_TAG .
docker tag activelog:$IMAGE_TAG username/activelog:$IMAGE_TAG
docker push username/activelog:$IMAGE_TAG

# Update task definition with new image
sed "s/:latest/:$IMAGE_TAG/g" task-definition.json > task-definition-new.json

# Register new task definition
NEW_TASK_DEF=$(aws ecs register-task-definition \
  --cli-input-json file://task-definition-new.json \
  --region $REGION \
  --query 'taskDefinition.taskDefinitionArn' \
  --output text)

echo "Registered new task definition: $NEW_TASK_DEF"

# Update service with new task definition
aws ecs update-service \
  --cluster $CLUSTER \
  --service $SERVICE \
  --task-definition $NEW_TASK_DEF \
  --force-new-deployment \
  --region $REGION

echo "Service updated. Waiting for deployment to complete..."

# Wait for service to stabilize
aws ecs wait services-stable \
  --cluster $CLUSTER \
  --services $SERVICE \
  --region $REGION

echo "Deployment complete!"
```

---

## 🔴 HTTPS/TLS Configuration

### Request Certificate in ACM
```bash
# Request certificate
aws acm request-certificate \
  --domain-name activelog.com \
  --subject-alternative-names www.activelog.com \
  --validation-method DNS \
  --region us-east-1

# Get certificate details
aws acm describe-certificate \
  --certificate-arn arn:aws:acm:region:account:certificate/id
```

### ALB HTTPS Configuration (Terraform)
```hcl
# Application Load Balancer
resource "aws_lb" "main" {
  name               = "activelog-alb"
  internal           = false
  load_balancer_type = "application"
  security_groups    = [aws_security_group.alb.id]
  subnets            = aws_subnet.public[*].id

  enable_deletion_protection = true
  enable_http2              = true
}

# HTTPS Listener
resource "aws_lb_listener" "https" {
  load_balancer_arn = aws_lb.main.arn
  port              = "443"
  protocol          = "HTTPS"
  ssl_policy        = "ELBSecurityPolicy-TLS-1-2-2017-01"
  certificate_arn   = aws_acm_certificate.main.arn

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.app.arn
  }
}

# HTTP to HTTPS Redirect
resource "aws_lb_listener" "http" {
  load_balancer_arn = aws_lb.main.arn
  port              = "80"
  protocol          = "HTTP"

  default_action {
    type = "redirect"

    redirect {
      port        = "443"
      protocol    = "HTTPS"
      status_code = "HTTP_301"
    }
  }
}

# Target Group
resource "aws_lb_target_group" "app" {
  name        = "activelog-tg"
  port        = 8080
  protocol    = "HTTP"
  vpc_id      = aws_vpc.main.id
  target_type = "ip"

  health_check {
    enabled             = true
    healthy_threshold   = 2
    unhealthy_threshold = 3
    timeout             = 5
    interval            = 30
    path                = "/health"
    matcher             = "200"
  }
}
```

---

## Monitoring Stack

### Prometheus Configuration (prometheus.yml)
```yaml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

scrape_configs:
  # Scrape from ECS tasks
  - job_name: 'activelog-api'
    ec2_sd_configs:
      - region: us-east-1
        port: 8080
        filters:
          - name: tag:Service
            values: [activelog-api]
    relabel_configs:
      - source_labels: [__meta_ec2_public_ip]
        target_label: __address__
        replacement: '$1:8080'

  # PostgreSQL Exporter
  - job_name: 'postgres'
    static_configs:
      - targets: ['postgres-exporter:9187']

  # Redis Exporter
  - job_name: 'redis'
    static_configs:
      - targets: ['redis-exporter:9121']

  # Node Exporter (if using EC2)
  - job_name: 'node'
    static_configs:
      - targets: ['node-exporter:9100']
```

### CloudWatch Logs Configuration
```go
import "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"

type CloudWatchLogger struct {
    client      *cloudwatchlogs.Client
    groupName   string
    streamName  string
}

func (l *CloudWatchLogger) Log(message string) error {
    _, err := l.client.PutLogEvents(context.Background(), &cloudwatchlogs.PutLogEventsInput{
        LogGroupName:  aws.String(l.groupName),
        LogStreamName: aws.String(l.streamName),
        LogEvents: []types.InputLogEvent{
            {
                Message:   aws.String(message),
                Timestamp: aws.Int64(time.Now().UnixMilli()),
            },
        },
    })
    return err
}
```

---

## 🔴 Distributed Tracing

### OpenTelemetry Setup
```go
import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/attribute"
    "go.opentelemetry.io/otel/exporters/otlp/otlptrace"
    "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
    "go.opentelemetry.io/otel/sdk/resource"
    sdktrace "go.opentelemetry.io/otel/sdk/trace"
    semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

// Initialize OpenTelemetry
func InitTracer() (*sdktrace.TracerProvider, error) {
    // Create OTLP exporter
    exporter, err := otlptrace.New(
        context.Background(),
        otlptracegrpc.NewClient(
            otlptracegrpc.WithEndpoint("localhost:4317"),
            otlptracegrpc.WithInsecure(),
        ),
    )
    if err != nil {
        return nil, err
    }

    // Create resource
    res, err := resource.New(
        context.Background(),
        resource.WithAttributes(
            semconv.ServiceNameKey.String("activelog-api"),
            semconv.ServiceVersionKey.String("1.0.0"),
        ),
    )
    if err != nil {
        return nil, err
    }

    // Create tracer provider
    tp := sdktrace.NewTracerProvider(
        sdktrace.WithBatcher(exporter),
        sdktrace.WithResource(res),
        sdktrace.WithSampler(sdktrace.AlwaysSample()),
    )

    otel.SetTracerProvider(tp)

    return tp, nil
}

// Use in handlers
func (h *ActivityHandler) GetActivity(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    // Create a span for this operation
    ctx, span := otel.Tracer("activelog").Start(ctx, "GetActivity")
    defer span.End()

    activityID := getActivityID(r)

    // Add attributes to span
    span.SetAttributes(
        attribute.Int("activity.id", activityID),
        attribute.String("user.id", getUserID(ctx)),
    )

    // Trace database call (pass ctx to propagate trace)
    activity, err := h.repo.GetByID(ctx, activityID)
    if err != nil {
        span.RecordError(err)
        response.Error(w, http.StatusNotFound, "Activity not found")
        return
    }

    response.JSON(w, http.StatusOK, activity)
}

// Trace database queries
func (r *ActivityRepository) GetByID(ctx context.Context, id int) (*models.Activity, error) {
    ctx, span := otel.Tracer("activelog").Start(ctx, "DB: GetActivityByID")
    defer span.End()

    span.SetAttributes(attribute.Int("activity.id", id))

    var activity models.Activity
    err := r.db.QueryRowContext(ctx, "SELECT * FROM activities WHERE id = $1", id).
        Scan(&activity.ID, &activity.UserID, /*...*/)

    if err != nil {
        span.RecordError(err)
        return nil, err
    }

    return &activity, nil
}
```

**Benefits:**
- Track requests across microservices
- Identify performance bottlenecks
- Visualize call chains
- Debug distributed systems

**Tools:**
- **OpenTelemetry** (standard instrumentation)
- **AWS X-Ray** (AWS-native tracing)
- **Jaeger** (open-source alternative)

---

## Dashboards

### Key Metrics to Monitor

1. **Request rate and latency**
   - Requests per second
   - p50, p95, p99 latencies
   - Error rates

2. **Error rates**
   - 4xx errors (client errors)
   - 5xx errors (server errors)
   - Error rate trends

3. **Database performance**
   - Query duration
   - Connection pool usage
   - Slow queries

4. **Cache hit rates**
   - Redis hit/miss ratio
   - Cache memory usage
   - Eviction rate

5. **System resources**
   - CPU utilization
   - Memory usage
   - Network I/O

---

## Auto-scaling Configuration

```hcl
# ECS Service Auto Scaling
resource "aws_appautoscaling_target" "ecs" {
  max_capacity       = 10
  min_capacity       = 2
  resource_id        = "service/${aws_ecs_cluster.main.name}/${aws_ecs_service.app.name}"
  scalable_dimension = "ecs:service:DesiredCount"
  service_namespace  = "ecs"
}

# Scale up based on CPU
resource "aws_appautoscaling_policy" "cpu" {
  name               = "cpu-autoscaling"
  policy_type        = "TargetTrackingScaling"
  resource_id        = aws_appautoscaling_target.ecs.resource_id
  scalable_dimension = aws_appautoscaling_target.ecs.scalable_dimension
  service_namespace  = aws_appautoscaling_target.ecs.service_namespace

  target_tracking_scaling_policy_configuration {
    predefined_metric_specification {
      predefined_metric_type = "ECSServiceAverageCPUUtilization"
    }
    target_value = 70.0
  }
}
```

---

## Common Pitfalls

1. **No HTTPS**
   - ❌ Insecure data transmission
   - ✅ Always use HTTPS in production

2. **Hardcoded secrets**
   - ❌ Secrets in environment variables
   - ✅ Use AWS Secrets Manager

3. **No monitoring**
   - ❌ Blind to production issues
   - ✅ Set up comprehensive monitoring

4. **Single AZ deployment**
   - ❌ No high availability
   - ✅ Multi-AZ RDS and ECS

---

## Resources

- [AWS ECS Documentation](https://docs.aws.amazon.com/ecs/)
- [AWS Certificate Manager](https://docs.aws.amazon.com/acm/)
- [OpenTelemetry Go](https://opentelemetry.io/docs/instrumentation/go/)
- [Grafana Dashboards](https://grafana.com/grafana/dashboards/)

---

## Next Steps

After completing Month 11, you'll move to **Month 12: Monetization & Polish**, where you'll learn:
- Stripe payment integration
- Subscription tiers
- Webhook handling
- Launch preparation

**Your app is now live in production!** 🎉
