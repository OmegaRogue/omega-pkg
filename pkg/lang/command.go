package lang

import (
	"context"
	"github.com/pkg/errors"
	"strings"
)

type Command struct {
	Cmd    string   `hcl:"cmd,optional"`
	Flags  []string `hcl:"flags,optional"`
	Inline []string `hcl:"inline"`
}

func (c *Command) Run(ctx context.Context) error {
	if c.Cmd == "" {
		c.Cmd = "/bin/sh"
	}
	c.Flags = append(c.Flags, "-c")
	flags := append(c.Flags, strings.Join(c.Inline, "\n"))
	if err := runCommand(ctx, c.Cmd, flags...); err != nil {
		return errors.Wrap(err, "run command")
	}
	return nil
}
