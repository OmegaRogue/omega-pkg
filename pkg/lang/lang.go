package lang

import (
	"context"
	"fmt"
	"github.com/Wing924/shellwords"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/pkg/errors"
	"github.com/zclconf/go-cty/cty"
	"os"
	"os/exec"
	"strings"
)

type contextKey struct {
	name string
}

const (
	ActionClean   = "clean"
	ActionInstall = "install"
	ActionRemove  = "remove"
	ActionRefresh = "refresh"
	ActionUpdate  = "update"
)

var (
	EnvContextKey           = contextKey{"env"}
	DryrunContextKey        = contextKey{"dryrun"}
	CwdContextKey           = contextKey{"cwd"}
	ActionContextKey        = contextKey{"action"}
	CustomManagerContextKey = contextKey{"customManager"}
	HclContextKey           = contextKey{"hclctx"}
)

type Action struct {
	Type       string         `hcl:"type,label"`
	CmdExpr    hcl.Expression `hcl:"cmd,optional"`
	Cmd        string
	FlagExprs  hcl.Expression `hcl:"flags,optional"`
	Flags      string
	InlineExpr []hcl.Expression `hcl:"inline,optional"`
	Inline     []string
}

func (a *Action) Validate(
	ctx *hcl.EvalContext, globalCmdExpr hcl.Expression, globalFlagExpr hcl.Expression,
) hcl.Diagnostics {
	var diags hcl.Diagnostics
	globalCmd := ""
	moreDiags := gohcl.DecodeExpression(globalCmdExpr, ctx, &globalCmd)
	diags = append(diags, moreDiags...)
	globalFlags := ""
	moreDiags = gohcl.DecodeExpression(globalFlagExpr, ctx, &globalFlags)
	diags = append(diags, moreDiags...)

	if a.Cmd == "" {
		if globalCmd != "" && len(a.Inline) == 0 {
			a.Cmd = globalCmd
		}
		if globalCmd == "" && len(a.Inline) == 0 {
			ran := a.CmdExpr.Range()
			for _, expression := range a.InlineExpr {
				ran = hcl.RangeOver(ran, expression.Range())
			}
			diag := &hcl.Diagnostic{
				Severity: hcl.DiagWarning,
				Summary:  fmt.Sprintf("no global command and no command or inline defined on action %s", a.Type),
				Subject:  &ran,
			}
			diags = append(diags, diag)
		}
		if len(a.Inline) > 0 {
			a.Cmd = "/bin/sh -c"
		}
	}
	moreDiags = gohcl.DecodeExpression(a.FlagExprs, ctx, &a.Flags)
	diags = append(diags, moreDiags...)
	if globalFlags != "" {
		a.Flags = fmt.Sprintf("%s %s", globalFlags, a.Flags)
	}

	if a.Cmd == "" {
		moreDiags := gohcl.DecodeExpression(a.CmdExpr, ctx, &a.Cmd)
		diags = append(diags, moreDiags...)
	}
	for _, expression := range a.InlineExpr {
		var inline string
		moreDiags := gohcl.DecodeExpression(expression, ctx, &inline)
		diags = append(diags, moreDiags...)
		a.Inline = append(a.Inline, inline)
	}
	if len(a.Inline) > 0 {
		a.Cmd = "/bin/sh -c"
	}

	return diags
}

func (a *Action) Run(ctx context.Context, data []string, additionalArgs ...string) error {
	args, err := shellwords.Split(a.Flags + strings.Join(additionalArgs, " "))
	if err != nil {
		return errors.Wrap(err, "error on split args")
	}
	if len(a.Inline) == 0 {
		if data != nil {
			args = append(args, data...)
		}
		if err := runCommand(ctx, a.Cmd, args...); err != nil {
			return errors.Wrap(err, "error on run command")
		}
		return nil
	}
	if data != nil {
		ctx = context.WithValue(ctx, EnvContextKey, []string{fmt.Sprintf("data=%s", data)})
	}

	if err := runCommand(ctx, a.Cmd, append(args, strings.Join(a.Inline, "\n"))...); err != nil {
		return errors.Wrap(err, "error on run command")
	}
	return nil

}
func runCommand(ctx context.Context, command string, args ...string) error {

	cmd := exec.CommandContext(ctx, command, args...)
	if cwd := ctx.Value(CwdContextKey); cwd != nil {
		cmd.Dir = cwd.(string)
	}
	if env := ctx.Value(EnvContextKey); env != nil {
		cmd.Env = append(os.Environ(), env.([]string)...)
	}

	if ctx.Value(DryrunContextKey) == true {
		for _, s := range cmd.Env {
			fmt.Println(s)
		}
		fmt.Println(command + " " + strings.Join(args, " "))
		return nil
	}

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return errors.Wrap(err, "error on start command")
	}

	if err := cmd.Wait(); err != nil {
		return errors.Wrap(err, "error on wait for command completion")
	}
	return nil

}

type CustomManager struct {
	Name      string         `hcl:"name,label"`
	CmdExpr   hcl.Expression `hcl:"cmd,optional"`
	FlagExprs hcl.Expression `hcl:"flags,optional"`
	Actions   []*Action      `hcl:"action,block"`
	ActionMap map[string]*Action
}

func (m *CustomManager) Validate(ctx *hcl.EvalContext) hcl.Diagnostics {
	var diags hcl.Diagnostics
	if m.ActionMap == nil {
		m.ActionMap = make(map[string]*Action)
	}
	if len(m.Actions) == 0 {
		diag := &hcl.Diagnostic{
			Severity: hcl.DiagWarning,
			Summary:  fmt.Sprintf("no actions defined on CustomManager %s", m.Name),
		}
		diags = append(diags, diag)
	} else {
		for _, action := range m.Actions {
			moreDiags := action.Validate(ctx.NewChild(), m.CmdExpr, m.FlagExprs)
			diags = append(diags, moreDiags...)
			m.ActionMap[action.Type] = action
		}
	}
	return diags
}

type Constraints struct {
	Value bool `hcl:"value"`
}

func (c *Constraints) Match() bool {
	if c == nil {
		return true
	}
	return c.Value
}

type Set struct {
	Action      string       `hcl:"action,label"`
	Packages    []string     `hcl:"packages"`
	Flags       string       `hcl:"flags,optional"`
	Constraints *Constraints `hcl:"constraints,block"`
}

func (s *Set) Run(ctx context.Context) error {
	run := s.Constraints.Match()
	action, ok := ctx.Value(ActionContextKey).(*Action)
	if !ok {
		return errors.New("action is nil")
	}
	if !run {
		return nil
	}
	if err := action.Run(ctx, s.Packages, s.Flags); err != nil {
		return errors.Wrap(err, "error on run action")
	}
	return nil
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
	DryRun       bool         `hcl:"dry,optional"`
	Sets         []Set        `hcl:"set,block"`
	Repositories []Repository `hcl:"repo,block"`
}

func (m *Manager) Run(ctx context.Context) error {
	if m.DryRun {
		ctx = context.WithValue(ctx, DryrunContextKey, true)
	}
	customManager, ok := ctx.Value(CustomManagerContextKey).(*CustomManager)
	if !ok {
		return errors.New("customManager is nil")
	}
	if err := customManager.ActionMap["refresh"].Run(ctx, nil); err != nil {
		return errors.Wrap(err, "error on refresh packages")
	}
	if m.Update {
		if err := customManager.ActionMap["update"].Run(ctx, nil); err != nil {
			return errors.Wrap(err, "error on update packages")
		}
	}
	for _, set := range m.Sets {
		action := customManager.ActionMap[set.Action]
		ctx = context.WithValue(ctx, ActionContextKey, action)
		if err := set.Run(ctx); err != nil {
			return errors.Wrapf(err, "error on %s packages", action.Type)
		}
	}
	if m.Cleanup {
		if err := customManager.ActionMap["clean"].Run(ctx, nil); err != nil {
			return errors.Wrap(err, "error on clean packages")
		}
	}
	return nil
}

type Config struct {
	Managers         []Manager        `hcl:"manager,block"`
	CustomManagers   []*CustomManager `hcl:"custom_manager,block"`
	CustomManagerMap map[string]*CustomManager
}

func (c *Config) Validate(ctx *hcl.EvalContext) hcl.Diagnostics {
	var diags hcl.Diagnostics
	if c.CustomManagerMap == nil {
		c.CustomManagerMap = make(map[string]*CustomManager)
	}
	for _, manager := range c.CustomManagers {
		moreDiags := manager.Validate(ctx.NewChild())
		diags = append(diags, moreDiags...)
		c.CustomManagerMap[manager.Name] = manager
	}
	for _, manager := range c.Managers {
		m, ok := c.CustomManagerMap[manager.Name]
		if !ok {
			diag := &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  fmt.Sprintf("manager %s does not exist", manager.Name),
				Detail:   fmt.Sprintf("manager declared for %s but is no known CustomManager", manager.Name),
			}
			diags = append(diags, diag)
		}
		for i, set := range manager.Sets {
			ctx := ctx.NewChild()
			pkgs := []cty.Value{}
			ctx.Variables["pkgs"] = cty.ListVal()
		}

	}

	return diags
}

func (c *Config) Run(ctx context.Context) error {
	for _, manager := range c.Managers {
		customManager := c.CustomManagerMap[manager.Name]
		ctx = context.WithValue(ctx, CustomManagerContextKey, customManager)
		if err := manager.Run(ctx); err != nil {
			return errors.Wrapf(err, "error on run manager %s", manager.Name)
		}
	}
	return nil
}

var spec = hcldec.ObjectSpec{
	"manager": &hcldec.BlockMapSpec{
		TypeName:   "manager",
		LabelNames: []string{"name"},
		Nested: hcldec.ObjectSpec{

			"update": &hcldec.AttrSpec{
				Name:     "update",
				Type:     cty.Bool,
				Required: false,
			},
			"clean": &hcldec.AttrSpec{
				Name:     "clean",
				Type:     cty.Bool,
				Required: false,
			},
			"dry": &hcldec.AttrSpec{
				Name:     "dry",
				Type:     cty.Bool,
				Required: false,
			},
			"set": &hcldec.BlockListSpec{
				TypeName: "set",
				Nested: &hcldec.BlockMapSpec{
					TypeName:   "set",
					LabelNames: []string{"action"},
					Nested:     hcldec.ObjectSpec{},
				},
				MinItems: 0,
				MaxItems: 0,
			},
			"repo": &hcldec.BlockListSpec{
				TypeName: "repo",
				Nested: &hcldec.BlockMapSpec{
					TypeName:   "repo",
					LabelNames: []string{"name"},
					Nested:     hcldec.ObjectSpec{},
				},
				MinItems: 0,
				MaxItems: 0,
			},
		},
	},
	"custom_manager": &hcldec.BlockMapSpec{
		TypeName:   "custom_manager",
		LabelNames: []string{"name"},
		Nested:     hcldec.ObjectSpec{},
	},
}
