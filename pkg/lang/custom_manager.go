package lang

import (
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/zclconf/go-cty/cty"
)

type CustomManager struct {
	Name string `hcl:"name,label"`

	Actions   []*Action `hcl:"action,block"`
	ActionMap map[string]*Action
	Remain    hcl.Body `hcl:",remain"`
}
type CustomManagerRemain struct {
	CmdExpr   hcl.Expression `hcl:"cmd,optional"`
	FlagExprs hcl.Expression `hcl:"flags,optional"`
}

var CustomManagerRemainSpec = hcldec.ObjectSpec{
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
			//moreDiags := action.Validate(ctx.NewChild(), m.CmdExpr, m.FlagExprs)
			//diags = append(diags, moreDiags...)
			m.ActionMap[action.Type] = action
		}
	}
	return diags
}

func (m *CustomManager) PrepareAction(ctx *hcl.EvalContext, name string) hcl.Diagnostics {
	var diags hcl.Diagnostics
	action, ok := m.ActionMap[name]
	if !ok {
		diag := &hcl.Diagnostic{
			Severity: hcl.DiagWarning,
			Summary:  fmt.Sprintf("action %s does not exist on manager %s", name, m.Name),
		}
		diags = append(diags, diag)
	} else {
		moreDiags := action.Prepare(ctx.NewChild(), m)
		diags = append(diags, moreDiags...)
	}
	return diags
}
