package queue

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/misalima/edunex-backend/internal/infra/logger"
	"go.uber.org/zap"
)

const (
	AnalysisJobNotifyChannel = "analysis_jobs_channel"
)

// NotificationListener listens to Postgres NOTIFY events and pushes job IDs to a channel
type NotificationListener struct {
	dbConnString string
	jobChan      chan uuid.UUID
	ctx          context.Context
	cancel       context.CancelFunc
	isActive     bool
	mu           sync.RWMutex
}

// NewNotificationListener creates a new notification listener
// connString should be a valid Postgres connection string
func NewNotificationListener(connString string, jobChan chan uuid.UUID) (*NotificationListener, error) {
	if connString == "" {
		return nil, fmt.Errorf("connection string is required")
	}

	ctx, cancel := context.WithCancel(context.Background())

	listener := &NotificationListener{
		dbConnString: connString,
		jobChan:      jobChan,
		ctx:          ctx,
		cancel:       cancel,
	}

	return listener, nil
}

// Start begins listening for notifications
func (nl *NotificationListener) Start() error {
	// Create a listener connection
	pgListener := pq.NewListener(
		nl.dbConnString,
		10*time.Second,
		time.Minute,
		func(ev pq.ListenerEventType, err error) {
			switch ev {
			case pq.ListenerEventConnected:
				logger.Log.Debug("Listener connected")
			case pq.ListenerEventDisconnected:
				logger.Log.Debug("Listener disconnected")
			case pq.ListenerEventReconnected:
				logger.Log.Debug("Listener reconnected")
			case pq.ListenerEventConnectionAttemptFailed:
				logger.Log.Warn("Listener connection attempt failed", zap.Error(err))
			}
		},
	)

	if err := pgListener.Listen(AnalysisJobNotifyChannel); err != nil {
		logger.Log.Error("failed to listen on channel", zap.Error(err))
		return fmt.Errorf("failed to listen on channel: %w", err)
	}

	nl.mu.Lock()
	nl.isActive = true
	nl.mu.Unlock()

	logger.Log.Info("Notification listener started")

	go nl.listen(pgListener)

	return nil
}

// listen polls for notifications
func (nl *NotificationListener) listen(pgListener *pq.Listener) {
	defer func() {
		nl.mu.Lock()
		nl.isActive = false
		nl.mu.Unlock()
		err := pgListener.Unlisten(AnalysisJobNotifyChannel)
		if err != nil {
			logger.Log.Error("failed to unlisten on channel", zap.Error(err))
		}
		err = pgListener.Close()
		if err != nil {
			logger.Log.Error("failed to close listener", zap.Error(err))
		}
		logger.Log.Info("Notification listener stopped")
	}()

	for {
		select {
		case <-nl.ctx.Done():
			return
		case notification := <-pgListener.Notify:
			if notification != nil {
				// Parse job ID from notification payload
				jobID, err := uuid.Parse(notification.Extra)
				if err != nil {
					logger.Log.Error("failed to parse job id from notification", zap.Error(err), zap.String("payload", notification.Extra))
					continue
				}

				select {
				case nl.jobChan <- jobID:
					logger.Log.Debug("job pushed to queue", zap.String("job_id", jobID.String()))
				case <-nl.ctx.Done():
					return
				}
			}
		}
	}
}

// Stop stops the listener
func (nl *NotificationListener) Stop() error {
	nl.cancel()

	// Wait for listener to be inactive
	timeout := time.After(5 * time.Second)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			return fmt.Errorf("listener stop timeout")
		case <-ticker.C:
			nl.mu.RLock()
			isActive := nl.isActive
			nl.mu.RUnlock()
			if !isActive {
				return nil
			}
		}
	}
}

// IsActive returns whether the listener is running
func (nl *NotificationListener) IsActive() bool {
	nl.mu.RLock()
	defer nl.mu.RUnlock()
	return nl.isActive
}
