package cron

import (
	"context"
	"sync"
	"time"

	"github.com/questx-lab/backend/pkg/xcontext"
)

type CronJob interface {
	Do(context.Context)
	RunNow() bool
	Next() time.Time
}

type CronJobManager struct {
	mutex sync.Mutex
	wait  sync.WaitGroup
	jobs  map[CronJob]*time.Timer
}

func NewCronJobManager() *CronJobManager {
	return &CronJobManager{jobs: make(map[CronJob]*time.Timer)}
}

func (m *CronJobManager) Register(job CronJob) {
	m.jobs[job] = nil
}

func (m *CronJobManager) Start(ctx context.Context) {
	xcontext.Logger(ctx).Infof("Cron job manager started")

	for job := range m.jobs {
		if job.RunNow() {
			go m.run(ctx, job)
		} else {
			m.schedule(ctx, job)
		}

		m.wait.Add(1)
	}

	m.wait.Wait()
	xcontext.Logger(ctx).Infof("Cron job manager stopped")
}

func (m *CronJobManager) Cancel(ctx context.Context) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	for job, timer := range m.jobs {
		if timer == nil {
			xcontext.Logger(ctx).Warnf("Stop a job that hasn't started: %T", job)
			continue
		}

		timer.Stop()
		m.wait.Done()
	}

	// Clear all jobs to not schedule them again.
	m.jobs = make(map[CronJob]*time.Timer)
}

func (m *CronJobManager) run(ctx context.Context, job CronJob) {
	xcontext.Logger(ctx).Infof("%T is running...", job)
	job.Do(ctx)
	xcontext.Logger(ctx).Infof("%T ok", job)

	m.schedule(ctx, job)
}

func (m *CronJobManager) schedule(ctx context.Context, job CronJob) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Only scheudule jobs which existed in job list.
	if _, ok := m.jobs[job]; ok {
		m.jobs[job] = time.AfterFunc(job.Next().Sub(time.Now()), func() { m.run(ctx, job) })
	}
}
