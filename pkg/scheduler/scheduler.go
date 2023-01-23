package scheduler

import (
	"fmt"
	"sync"
	"time"

	"github.com/go-logr/logr"
)

var ClusterStatusInterval time.Duration
var ClusterBackupsInterval time.Duration

const StatusChecker = "statusChecker"
const BackupsChecker = "backupsChecker"

type Job func() error

type Interface interface {
	ScheduleJob(jobID string, interval time.Duration, job Job) error
	RemoveJob(jobID string)
	Stop()
}

type workItem struct {
	ID     string
	ticker *time.Ticker
	done   chan struct{}
}

type scheduler struct {
	logger    logr.Logger
	close     chan struct{}
	m         sync.Mutex
	workItems map[string]*workItem
	isClosed  bool
}

func NewScheduler(logger logr.Logger) Interface {
	s := &scheduler{
		logger:    logger,
		close:     make(chan struct{}),
		workItems: make(map[string]*workItem),
	}

	go func() {
		<-s.close
		s.m.Lock()
		defer s.m.Unlock()
		for _, wi := range s.workItems {
			go func(w *workItem) {
				w.done <- struct{}{}
			}(wi)
		}
		s.isClosed = true
	}()

	return s
}

func (s *scheduler) Stop() {
	s.close <- struct{}{}
}

func (s *scheduler) ScheduleJob(jobID string, interval time.Duration, job Job) error {
	s.m.Lock()
	defer s.m.Unlock()

	if s.isClosed {
		return fmt.Errorf("scheduler is closed")
	}

	if interval <= 0 {
		return fmt.Errorf("interval need to be more than 0")
	}

	_, exists := s.workItems[jobID]
	if exists {
		s.logger.Info("Scheduled task already exists", "job id", jobID)
		return nil
	}

	wi := &workItem{
		ID:     jobID,
		ticker: time.NewTicker(interval),
		done:   make(chan struct{}),
	}

	go func() {
		for {
			select {
			case <-wi.ticker.C:
				err := job()
				if err != nil {
					s.logger.Error(err, "Failed job ID", "Job ID", wi.ID)
				}
			case <-wi.done:
				wi.ticker.Stop()
				return
			}
		}
	}()

	s.workItems[jobID] = wi

	s.logger.Info("Scheduled job has been started", "job id", jobID)

	return nil
}

func (s *scheduler) RemoveJob(jobID string) {
	s.m.Lock()
	defer s.m.Unlock()

	v, exists := s.workItems[jobID]
	if !exists {
		s.logger.Info("There is no scheduled job", "job id", jobID)
		return
	}

	delete(s.workItems, jobID)
	s.logger.Info("Job was removed", "job id", jobID)

	v.done <- struct{}{}
}
