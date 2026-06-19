package audit

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/notfixingit3/echostate/internal/config"
	"github.com/notfixingit3/echostate/internal/db"
)

const purgeInterval = time.Hour

// Purger periodically deletes audit events past the configured retention window.
type Purger struct {
	db *db.DB

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

func NewPurger(database *db.DB) *Purger {
	return &Purger{db: database}
}

func (p *Purger) Start(ctx context.Context) {
	p.ctx, p.cancel = context.WithCancel(ctx)
	p.wg.Add(1)
	go p.loop()
}

func (p *Purger) Stop() {
	if p.cancel != nil {
		p.cancel()
	}
	p.wg.Wait()
}

// RunOnce purges expired audit rows using current settings.
func (p *Purger) RunOnce(ctx context.Context) {
	days := config.GetSettings().AuditRetentionDays
	if days <= 0 {
		return
	}

	runCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	deleted, err := PurgeExpired(runCtx, p.db, days)
	if err != nil {
		log.Printf("audit purge: %v", err)
		return
	}
	if deleted > 0 {
		log.Printf("audit purge: removed %d events older than %d days", deleted, days)
	}
}

func (p *Purger) loop() {
	defer p.wg.Done()

	p.RunOnce(p.ctx)
	timer := time.NewTimer(purgeInterval)
	defer timer.Stop()

	for {
		select {
		case <-p.ctx.Done():
			return
		case <-timer.C:
			p.RunOnce(p.ctx)
			timer.Reset(purgeInterval)
		}
	}
}