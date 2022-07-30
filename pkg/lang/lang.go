package lang

import (
	"context"
	"fmt"
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
)

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

func PrepareFlags(ctx *hcl.EvalContext, flagExprs []hcl.Expression) (hcl.Diagnostics, [][]string) {
	var diags hcl.Diagnostics
	var flags [][]string

	for _, flagExpr := range flagExprs {

		var flag []string
		moreDiags := gohcl.DecodeExpression(flagExpr, ctx, &flag)
		diags = append(diags, moreDiags...)
		flags = append(flags, flag)
	}

	return diags, flags
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
				Nested:   SetSpec,
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
