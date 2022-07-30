package lang

import (
	"context"
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/pkg/errors"
)

type ManagerOperation struct {
	Name         string       `hcl:"name,label"`
	Update       bool         `hcl:"update,optional"`
	Cleanup      bool         `hcl:"clean,optional"`
	DryRun       bool         `hcl:"dry,optional"`
	Sets         []Set        `hcl:"set,block"`
	Repositories []Repository `hcl:"repo,block"`
}

func (m *ManagerOperation) Run(ctx context.Context) error {
	if m.DryRun {
		ctx = context.WithValue(ctx, DryrunContextKey, true)
	}
	customManager, ok := ctx.Value(CustomManagerContextKey).(*CustomManager)
	if !ok {
		return errors.New("customManager is nil")
	}
	if err := customManager.ActionMap["refresh"].Run(ctx); err != nil {
		return errors.Wrap(err, "refresh packages")
	}
	if m.Update {
		if err := customManager.ActionMap["update"].Run(ctx); err != nil {
			return errors.Wrap(err, "update packages")
		}
	}
	for _, set := range m.Sets {
		action := customManager.ActionMap[set.Action]
		ctx = context.WithValue(ctx, ActionContextKey, action)
		if err := set.Run(ctx); err != nil {
			return errors.Wrapf(err, "%s packages", action.Type)
		}
	}
	if m.Cleanup {
		if err := customManager.ActionMap["clean"].Run(ctx); err != nil {
			return errors.Wrap(err, "clean packages")
		}
	}
	return nil
}

func (m *ManagerOperation) PrepareSets(ctx *hcl.EvalContext, customManager *CustomManager) hcl.Diagnostics {
	var diags hcl.Diagnostics
	for i, set := range m.Sets {
		action, ok := customManager.ActionMap[set.Action]
		if !ok {
			diag := &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  fmt.Sprintf("action %s does not exist on manager %s", set.Action, m.Name),
			}
			diags = append(diags, diag)
		}

		moreDiags := set.Prepare(ctx.NewChild(), customManager, action)
		diags = append(diags, moreDiags...)
		m.Sets[i] = set
	}
	return diags
}
