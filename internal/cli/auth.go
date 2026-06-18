package cli

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/notfixingit3/echostate/internal/auth"
	"github.com/notfixingit3/echostate/internal/config"
	"github.com/notfixingit3/echostate/internal/db"
)

func RunAuth(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: echostate auth <bootstrap-admin|issue-admin-code>")
	}

	switch args[0] {
	case "bootstrap-admin":
		return runBootstrapAdmin(args[1:])
	case "issue-admin-code":
		return runIssueAdminCode(args[1:])
	default:
		return fmt.Errorf("unknown auth command: %s", args[0])
	}
}

func runBootstrapAdmin(args []string) error {
	fs := flag.NewFlagSet("bootstrap-admin", flag.ExitOnError)
	name := fs.String("name", "Admin", "display name for the first admin user")
	if err := fs.Parse(args); err != nil {
		return err
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	database, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer database.Close()

	if err := db.Migrate(database); err != nil {
		return err
	}

	svc := auth.NewService(database, cfg.FrontendURL)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	user, code, err := svc.BootstrapAdmin(ctx, *name)
	if err != nil {
		return err
	}

	fmt.Printf("Admin user created: %s (%s)\n", user.DisplayName, user.ID)
	fmt.Printf("Enrollment code (single-use, expires %s): %s\n", code.ExpiresAt.Format(time.RFC3339), code.Code)
	fmt.Println("Open the EchoState login page, enter this code, and register a passkey.")
	return nil
}

func runIssueAdminCode(args []string) error {
	if err := flag.NewFlagSet("issue-admin-code", flag.ExitOnError).Parse(args); err != nil {
		return err
	}

	secret := strings.TrimSpace(os.Getenv("ECHOSTATE_BREAK_GLASS_SECRET"))
	if secret == "" {
		return auth.ErrBreakGlassSecret
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	database, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer database.Close()

	svc := auth.NewService(database, cfg.FrontendURL)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	admin, err := svc.FindAdminUser(ctx)
	if err != nil {
		return fmt.Errorf("admin user not found: %w", err)
	}

	code, err := svc.IssueEnrollmentCode(ctx, admin.ID, nil, auth.PurposeRecovery, true)
	if err != nil {
		return err
	}

	fmt.Printf("Recovery code for %s (expires %s): %s\n", admin.DisplayName, code.ExpiresAt.Format(time.RFC3339), code.Code)
	return nil
}