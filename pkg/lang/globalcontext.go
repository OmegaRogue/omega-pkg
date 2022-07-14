package lang

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/pkg/errors"
	"github.com/zcalusic/sysinfo"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

func BuildGlobalContext() (*hcl.EvalContext, error) {
	var si sysinfo.SysInfo
	si.GetSysInfo()
	typ, err := gocty.ImpliedType(si)
	if err != nil {
		return nil, errors.Wrap(err, "error on convert sysinfo to cty.Type")
	}
	val, err := gocty.ToCtyValue(si, typ)
	if err != nil {
		return nil, errors.Wrap(err, "error on convert sysinfo to cty.Value")
	}
	ctx := &hcl.EvalContext{
		Variables: map[string]cty.Value{
			"sysinfo": val,
			"variant": cty.StringVal(""),
		},
	}
	return ctx, nil
}
