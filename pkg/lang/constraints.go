package lang

import (
	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/zclconf/go-cty/cty"
)

type Constraints struct {
	Value bool `hcl:"value"`
}

var ConstraintSpec = hcldec.ObjectSpec{
	"value": &hcldec.AttrSpec{
		Name:     "value",
		Type:     cty.Bool,
		Required: true,
	},
}

func (c *Constraints) Match() bool {
	if c == nil {
		return true
	}
	return c.Value
}
