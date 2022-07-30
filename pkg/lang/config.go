package lang

import (
	"context"
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/pkg/errors"
)

type Config struct {
	Managers         []ManagerOperation `hcl:"manager,block"`
	CustomManagers   []*CustomManager   `hcl:"custom_manager,block"`
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
		customManager, ok := c.CustomManagerMap[manager.Name]
		if !ok {
			diag := &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  fmt.Sprintf("manager %s does not exist", manager.Name),
				Detail:   fmt.Sprintf("manager declared for %s but is no known CustomManager", manager.Name),
			}
			diags = append(diags, diag)
		}
		moreDiags := customManager.PrepareAction(ctx, "update")
		diags = append(diags, moreDiags...)

		moreDiags = customManager.PrepareAction(ctx, "clean")
		diags = append(diags, moreDiags...)

		moreDiags = customManager.PrepareAction(ctx, "refresh")
		diags = append(diags, moreDiags...)

		moreDiags = manager.PrepareSets(ctx, customManager)
		diags = append(diags, moreDiags...)

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
