package di

import (
	"database/sql"

	brokerDI "github.com/valentinesamuel/activelog/internal/application/broker/di"
	"github.com/valentinesamuel/activelog/internal/container"
	queueDI "github.com/valentinesamuel/activelog/internal/queue/di"
	"github.com/valentinesamuel/activelog/internal/queue/types"
	"github.com/valentinesamuel/activelog/internal/scheduler"
	"github.com/valentinesamuel/activelog/internal/service"
)

// RegisterScheduler registers the Scheduler in the DI container.
// Depends on: rawDB (broker/di.CoreRawDBKey) and QueueProvider.
func RegisterScheduler(c *container.Container) {
	c.Register(SchedulerKey, func(c *container.Container) (interface{}, error) {
		rawDB := c.MustResolve(brokerDI.CoreRawDBKey).(*sql.DB)
		queue := c.MustResolve(queueDI.QueueProviderKey).(types.QueueProvider)

		statsCalc := service.NewStatsCalculator(rawDB)
		cleanup := service.NewCleanupService(rawDB)

		return scheduler.New(statsCalc, cleanup, queue), nil
	})
}
