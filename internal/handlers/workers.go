package handlers

import (
	"github.com/notfixingit3/echostate/internal/audit"
	"github.com/notfixingit3/echostate/internal/reports"
	"github.com/notfixingit3/echostate/internal/scans"
	"github.com/notfixingit3/echostate/internal/scheduler"
)

// Workers groups long-running background workers started by Register.
type Workers struct {
	Reports     *reports.Worker
	Scans       *scans.Worker
	Scheduler   *scheduler.Scheduler
	AuditPurger *audit.Purger
}
