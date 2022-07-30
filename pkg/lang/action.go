package lang

import (
	"context"
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/pkg/errors"
	"github.com/zclconf/go-cty/cty"
	"omega-pkg/pkg/utils"
	"strings"
)

type Action struct {
	Type    string `hcl:"type,label"`
	command []string
	Remain  hcl.Body `hcl:",remain"`
}

var ActionRemainSpec = hcldec.ObjectSpec{
	"cmd": &hcldec.AttrSpec{
		Name:     "cmd",
		Type:     cty.String,
		Required: false,
	},
	"flags": &hcldec.AttrSpec{
		Name:     "flags",
		Type:     cty.List(cty.String),
		Required: false,
	},
	"inline": &hcldec.AttrSpec{
		Name:     "inline",
		Type:     cty.List(cty.String),
		Required: false,
	},
}

func (a *Action) Prepare(
	ctx *hcl.EvalContext, manager *CustomManager,
) (diags hcl.Diagnostics) {
	managerRemain, moreDiags := hcldec.Decode(manager.Remain, CustomManagerRemainSpec, ctx)
	diags = append(diags, moreDiags...)
	globalCmd := utils.ValueToString(managerRemain.GetAttr("cmd"))
	managerFlags := utils.MapValueToString(managerRemain.GetAttr("flags"))

	actionRemain, moreDiags := hcldec.Decode(a.Remain, ActionRemainSpec, ctx)
	diags = append(diags, moreDiags...)
	actionCmd := utils.ValueToString(actionRemain.GetAttr("cmd"))
	actionFlags := utils.MapValueToString(actionRemain.GetAttr("flags"))
	actionInline := utils.MapValueToString(actionRemain.GetAttr("inline"))

	var flags []string
	if len(actionInline) == 0 {
		flags = append(managerFlags, actionFlags...)
	}
	if len(actionInline) > 0 {
		if actionCmd == "" {
			actionCmd = "/bin/sh"
			flags = []string{"-c"}
		}
		flags = append(actionFlags, append(flags, strings.Join(actionInline, "\n"))...)
	} else if actionCmd == "" && globalCmd != "" {
		actionCmd = globalCmd
	} else {
		diag := &hcl.Diagnostic{
			Severity: hcl.DiagWarning,
			Summary:  fmt.Sprintf("no global command and no command or inline defined on action %s", a.Type),
		}
		diags = append(diags, diag)
	}
	a.command = append([]string{actionCmd}, flags...)
	return diags

}

func (a *Action) Run(ctx context.Context) error {
	if err := runCommand(ctx, a.command[0], a.command[1:]...); err != nil {
		return errors.Wrapf(err, "run command on action %s", a.Type)
	}
	return nil
}
