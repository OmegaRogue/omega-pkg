package lang

import (
	"context"
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/pkg/errors"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
	"omega-pkg/pkg/utils"
	"strings"
)

type Set struct {
	Action      string   `hcl:"action,label"`
	Packages    []string `hcl:"packages"`
	command     []string
	Constraints *Constraints `hcl:"constraints,block"`
	Remain      hcl.Body     `hcl:",remain"`
}
type SetRemain struct {
	FlagExpr hcl.Expression `hcl:"flags,optional"`
}

var SetSpec = hcldec.ObjectSpec{
	"packages": &hcldec.AttrSpec{
		Name:     "packages",
		Type:     cty.List(cty.String),
		Required: true,
	},
	"constraints": &hcldec.BlockSpec{
		TypeName: "constraints",
		Nested:   ConstraintSpec,
		Required: false,
	},
	"action": &hcldec.BlockLabelSpec{
		Index: 0,
		Name:  "action",
	},
}
var SetRemainSpec = hcldec.ObjectSpec{
	"flags": &hcldec.AttrSpec{
		Name:     "flags",
		Type:     cty.List(cty.String),
		Required: false,
	},
}

func (s *Set) Prepare(
	ctx *hcl.EvalContext, manager *CustomManager, action *Action,
) (diags hcl.Diagnostics) {
	pkgs, err := gocty.ToCtyValue(s.Packages, cty.List(cty.String))
	if err != nil {
		diag := &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "error converting list of packages to cty value",
		}
		diags = append(diags, diag)
	}
	ctx.Variables = map[string]cty.Value{"pkgs": pkgs}

	managerRemain, moreDiags := hcldec.Decode(manager.Remain, CustomManagerRemainSpec, ctx)
	diags = append(diags, moreDiags...)
	globalCmd := utils.ValueToString(managerRemain.GetAttr("cmd"))
	managerFlags := utils.MapValueToString(managerRemain.GetAttr("flags"))

	actionRemain, moreDiags := hcldec.Decode(action.Remain, ActionRemainSpec, ctx)
	diags = append(diags, moreDiags...)
	actionCmd := utils.ValueToString(actionRemain.GetAttr("cmd"))
	actionFlags := utils.MapValueToString(actionRemain.GetAttr("flags"))
	actionInline := utils.MapValueToString(actionRemain.GetAttr("inline"))

	setRemain, moreDiags := hcldec.Decode(s.Remain, ActionRemainSpec, ctx)
	diags = append(diags, moreDiags...)
	setFlags := utils.MapValueToString(setRemain.GetAttr("flags"))

	var flags []string
	if len(actionInline) == 0 {
		flags = append(managerFlags, append(actionFlags, setFlags...)...)
	}

	if len(actionInline) > 0 {
		if actionCmd == "" {
			actionCmd = "/bin/sh"
			flags = []string{"-c"}
		}
		flags = append(flags, append(actionFlags, strings.Join(actionInline, "\n"))...)
	} else if actionCmd == "" && globalCmd != "" {
		actionCmd = globalCmd
	} else {

		diag := &hcl.Diagnostic{
			Severity: hcl.DiagWarning,
			Summary:  fmt.Sprintf("no global command and no command or inline defined on action %s", action.Type),
		}
		diags = append(diags, diag)
	}
	s.command = append([]string{actionCmd}, flags...)

	command := strings.Join(s.command, " ")
	appendPackages := true
	for _, s2 := range s.Packages {
		if strings.Contains(command, s2) {
			appendPackages = false
		}
	}
	if appendPackages {
		s.command = append(s.command, s.Packages...)
	}

	return diags
}

func (s *Set) Run(ctx context.Context) error {
	if err := runCommand(ctx, s.command[0], s.command[1:]...); err != nil {
		return errors.Wrapf(err, "error on run command on set of action %s", s.Action)
	}
	return nil
}
