package scheduler

import (
	"context"
	"time"

	"dingtalk-dashboard/internal/domain/approval"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

// Scheduler handles scheduled sync jobs
type Scheduler struct {
	cron        *cron.Cron
	service     *approval.Service
	processCode string
	logger      *zap.Logger
}

// NewScheduler creates a new scheduler
func NewScheduler(service *approval.Service, processCode string, loc *time.Location, logger *zap.Logger) *Scheduler {
	return &Scheduler{
		cron:        cron.New(cron.WithLocation(loc)),
		service:     service,
		processCode: processCode,
		logger:      logger,
	}
}

// Start starts the scheduler
func (s *Scheduler) Start() error {
	// Schedule at 8AM, 11AM, 1PM, 4PM, 6PM daily (UTC+7)
	_, err := s.cron.AddFunc("0 8,11,13,16,18 * * *", s.runSync)
	if err != nil {
		return err
	}

	s.cron.Start()
	s.logger.Info("Scheduler started",
		zap.String("schedule", "8:00, 11:00, 13:00, 16:00, 18:00 daily"))

	return nil
}

// Stop stops the scheduler
func (s *Scheduler) Stop() {
	s.cron.Stop()
	s.logger.Info("Scheduler stopped")
}

// RunManualSync runs a manual sync
func (s *Scheduler) RunManualSync(ctx context.Context) (*approval.SyncLog, error) {
	s.logger.Info("Running manual sync")
	return s.service.SyncApprovals(ctx, s.processCode, "manual")
}

// runSync is the scheduled sync job
func (s *Scheduler) runSync() {
	s.logger.Info("Running scheduled sync")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	if _, err := s.service.SyncApprovals(ctx, s.processCode, "scheduled"); err != nil {
		s.logger.Error("Scheduled sync failed", zap.Error(err))
	}
}
