package lang

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/zclconf/go-cty/cty"
)

type Local struct {
	Remain hcl.Body `hcl:",remain"`
}
type LocalConfig struct {
	Locals []*Local `hcl:"locals,block"`
	Remain hcl.Body `hcl:",remain"`
}

func DecodeLocals(body hcl.Body, ctx *hcl.EvalContext) (
	locals map[string]cty.Value, remain hcl.Body, diags hcl.Diagnostics,
) {
	locals = make(map[string]cty.Value)
	var loc LocalConfig
	moreDiags := gohcl.DecodeBody(body, ctx, &loc)
	diags = append(diags, moreDiags...)
	for _, local := range loc.Locals {
		attrs, moreDiags := local.Remain.JustAttributes()
		diags = append(diags, moreDiags...)
		for key, attribute := range attrs {
			val, moreDiags := attribute.Expr.Value(ctx)
			diags = append(diags, moreDiags...)
			locals[key] = val
			ctx.Variables["local"] = cty.ObjectVal(locals)
		}
	}
	remain = loc.Remain
	return
}
