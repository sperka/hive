package cli

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"

	"github.com/mdipanjan/hive/internal/update"
)

func Update(args []string, currentVersion string) error {
	fs := flag.NewFlagSet("update", flag.ContinueOnError)
	check := fs.Bool("check", false, "Check for a new release without installing it")
	fs.Usage = func() {
		fmt.Fprintln(fs.Output(), "Usage: hive update [--check]")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if fs.NArg() != 0 {
		return fmt.Errorf("unexpected argument %q; usage: hive update [--check]", fs.Arg(0))
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	return update.Run(ctx, currentVersion, *check, os.Stdout)
}
