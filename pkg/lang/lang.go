package lang

import (
	"context"
	"fmt"
	"github.com/Wing924/shellwords"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"os"
	"os/exec"
	"strings"
)

const (
	ActionClean   = "clean"
	ActionInstall = "install"
	ActionRemove  = "remove"
	ActionRefresh = "refresh"
	ActionUpdate  = "update"
)

type CustomManagerAction struct {
	Type   string   `hcl:"type,label"`
	Cmd    string   `hcl:"cmd,optional"`
	Flags  string   `hcl:"flags,optional"`
	Inline []string `hcl:"inline,optional"`
}

func (a *CustomManagerAction) Validate(globalCmd string, hasCmd bool) error {
	if hasCmd && len(a.Inline) == 0 {
		a.Cmd = globalCmd
	}
	if a.Cmd == "" && len(a.Inline) > 0 {
		a.Cmd = "/bin/sh"
		a.Flags = "-c"
	}
	if !hasCmd && a.Cmd == "" && len(a.Inline) == 0 {
		return errors.Errorf(
			"no global command and no command or inline defined on action %s", a.Type,
		)
	}
	return nil
}

func (a *CustomManagerAction) Run(ctx context.Context, logger zerolog.Logger, data []string, additionalArgs ...string) {
	args, err := shellwords.Split(a.Flags)
	if err != nil {
		logger.Err(err).Msg("error on split args")
	}
	args = append(args, additionalArgs...)

	if len(a.Inline) == 0 {
		if data != nil {
			args = append(args, data...)
		}
		runCommand(ctx, logger, a.Cmd, args...)
		return
	}
	if data != nil {
		ctx = context.WithValue(ctx, "env", []string{fmt.Sprintf("data=%s", data)})
	}

	runCommand(ctx, logger, a.Cmd, append(args, strings.Join(a.Inline, "\n"))...)

}
func runCommand(ctx context.Context, logger zerolog.Logger, command string, args ...string) {

	cmd := exec.CommandContext(ctx, command, args...)
	if cwd := ctx.Value("cwd"); cwd != nil {
		cmd.Dir = cwd.(string)
	}
	if env := ctx.Value("env"); env != nil {
		cmd.Env = append(os.Environ(), env.([]string)...)
	}

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		logger.Err(err).Msg("error on run command")
	}

	if err := cmd.Wait(); err != nil {
		logger.Err(err).Msg("error on run command")
	}
}

type CustomManager struct {
	Name      string                 `hcl:"name,label"`
	Cmd       string                 `hcl:"cmd,optional"`
	Flags     []string               `hcl:"flags,optional"`
	Actions   []*CustomManagerAction `hcl:"action,block"`
	ActionMap map[string]*CustomManagerAction
}

func (m *CustomManager) Validate() error {
	if m.ActionMap == nil {
		m.ActionMap = make(map[string]*CustomManagerAction)
	}
	hasCmd := m.Cmd != ""
	if len(m.Actions) == 0 {
		return errors.Errorf("no actions defined on m %s", m.Name)
	}
	for _, action := range m.Actions {
		if err := action.Validate(m.Cmd, hasCmd); err != nil {
			return errors.Wrapf(err, "error on manager %s", m.Name)
		}
		m.ActionMap[action.Type] = action
	}
	return nil
}

type Constraints struct {
	OS       string   `hcl:"os,optional"`
	Arch     string   `hcl:"arch,optional"`
	Variants []string `hcl:"variants,optional"`
}

type Set struct {
	Action      string       `hcl:"action,label"`
	Packages    []string     `hcl:"packages"`
	Flags       []string     `hcl:"flags,optional"`
	Constraints *Constraints `hcl:"constraints,block"`
}

type Repository struct {
	Name        string       `hcl:"name,label"`
	Url         string       `hcl:"url"`
	Type        string       `hcl:"type,optional"`
	Key         string       `hcl:"key,optional"`
	Constraints *Constraints `hcl:"constraints,block"`
}

type Manager struct {
	Name         string       `hcl:"name,label"`
	Update       bool         `hcl:"update,optional"`
	Cleanup      bool         `hcl:"clean,optional"`
	Sets         []Set        `hcl:"set,block"`
	Repositories []Repository `hcl:"repo,block"`
}
type Config struct {
	Managers         []Manager        `hcl:"manager,block"`
	CustomManagers   []*CustomManager `hcl:"custom_manager,block"`
	CustomManagerMap map[string]*CustomManager
}

func (c *Config) Validate() error {
	if c.CustomManagerMap == nil {
		c.CustomManagerMap = make(map[string]*CustomManager)
	}
	for _, manager := range c.CustomManagers {
		if err := manager.Validate(); err != nil {
			return errors.Wrap(err, "error parsing custom managers")
		}
		c.CustomManagerMap[manager.Name] = manager
	}
	for _, manager := range c.Managers {
		if _, ok := c.CustomManagerMap[manager.Name]; !ok {
			return errors.Errorf("manager %s does not exist", manager.Name)
		}
	}

	return nil
}
