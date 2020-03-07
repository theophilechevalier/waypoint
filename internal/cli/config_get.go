package cli

import (
	"fmt"
	"os"
)

type ConfigGetCommand struct {
	*baseCommand
}

func (c *ConfigGetCommand) Run(args []string) int {
	if len(args) != 1 {
		fmt.Fprintf(os.Stderr, "config-set requires 1 arguments: a variable name")
		return 1
	}

	ctx := c.Ctx
	log := c.Log.Named("exec")

	// Initialize. If we fail, we just exit since Init handles the UI.
	if err := c.Init(); err != nil {
		return 1
	}

	cfg := c.cfg
	proj := c.project

	// NOTE(mitchellh): temporary restriction
	if len(cfg.Apps) != 1 {
		c.ui.Error("only one app is supported at this time")
		return 1
	}

	// Get our app
	app, err := proj.App(cfg.Apps[0].Name)
	if err != nil {
		c.logError(c.Log, "failed to initialize app", err)
		return 1
	}

	cv, err := app.ConfigGet(ctx, args[0])
	if err != nil {
		log.Error("error exec", "error", err)
		return 1
	}

	c.ui.Output(cv.Value)

	return 0
}

func (c *ConfigGetCommand) Synopsis() string {
	return ""
}

func (c *ConfigGetCommand) Help() string {
	return ""
}