package lang

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/zclconf/go-cty/cty"
)

type Variable struct {
	Name         string         `hcl:"name,label"`
	Type         hcl.Expression `hcl:"type"`
	DefaultValue hcl.Expression `hcl:"default"`
	Value        cty.Value
}

type Variables []*Variable

func (v Variables) GetMap() map[string]*Variable {
	varMap := make(map[string]*Variable)
	for _, variable := range v {
		varMap[variable.Name] = variable
	}
	return varMap
}

func (v *Variable) ApplyDefault(ctx *hcl.EvalContext) hcl.Diagnostics {
	var diags hcl.Diagnostics
	if v.Value.IsNull() {
		val, moreDiags := v.DefaultValue.Value(ctx)
		diags = append(diags, moreDiags...)
		v.Value = val
	}

	return diags
}

func (v Variables) ApplyDefaults(ctx *hcl.EvalContext) hcl.Diagnostics {
	var diags hcl.Diagnostics
	for _, variable := range v {
		moreDiags := variable.ApplyDefault(ctx)
		diags = append(diags, moreDiags...)
	}
	return diags
}

func (v Variables) GetCtyObject() map[string]cty.Value {
	varMap := make(map[string]cty.Value)
	for _, variable := range v {
		varMap[variable.Name] = variable.Value
	}
	return varMap
}

type VariableConfig struct {
	Variables Variables `hcl:"variable,block"`
	Remain    hcl.Body  `hcl:",remain"`
}

func DecodeVariable(body hcl.Body, ctx *hcl.EvalContext) (
	vars map[string]cty.Value, remain hcl.Body, diags hcl.Diagnostics,
) {
	var vari VariableConfig
	moreDiags := gohcl.DecodeBody(body, ctx, &vari)
	diags = append(diags, moreDiags...)

	moreDiags = vari.Variables.ApplyDefaults(ctx)
	diags = append(diags, moreDiags...)

	vars = vari.Variables.GetCtyObject()
	remain = vari.Remain
	return
}
