package lang

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/ext/tryfunc"
	"github.com/pkg/errors"
	"github.com/zcalusic/sysinfo"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
	"github.com/zclconf/go-cty/cty/function/stdlib"
	"github.com/zclconf/go-cty/cty/gocty"
)

func Functions() map[string]function.Function {

	functions := map[string]function.Function{
		"abs":             stdlib.AbsoluteFunc,
		"can":             tryfunc.CanFunc,
		"ceil":            stdlib.CeilFunc,
		"chomp":           stdlib.ChompFunc,
		"coalescelist":    stdlib.CoalesceListFunc,
		"compact":         stdlib.CompactFunc,
		"concat":          stdlib.ConcatFunc,
		"contains":        stdlib.ContainsFunc,
		"csvdecode":       stdlib.CSVDecodeFunc,
		"distinct":        stdlib.DistinctFunc,
		"element":         stdlib.ElementFunc,
		"chunklist":       stdlib.ChunklistFunc,
		"flatten":         stdlib.FlattenFunc,
		"floor":           stdlib.FloorFunc,
		"format":          stdlib.FormatFunc,
		"formatdate":      stdlib.FormatDateFunc,
		"formatlist":      stdlib.FormatListFunc,
		"indent":          stdlib.IndentFunc,
		"index":           stdlib.IndexFunc,
		"join":            stdlib.JoinFunc,
		"jsondecode":      stdlib.JSONDecodeFunc,
		"jsonencode":      stdlib.JSONEncodeFunc,
		"keys":            stdlib.KeysFunc,
		"length":          stdlib.LengthFunc,
		"log":             stdlib.LogFunc,
		"lookup":          stdlib.LookupFunc,
		"lower":           stdlib.LowerFunc,
		"max":             stdlib.MaxFunc,
		"merge":           stdlib.MergeFunc,
		"min":             stdlib.MinFunc,
		"parseint":        stdlib.ParseIntFunc,
		"pow":             stdlib.PowFunc,
		"range":           stdlib.RangeFunc,
		"regex":           stdlib.RegexFunc,
		"regexall":        stdlib.RegexAllFunc,
		"replace":         stdlib.ReplaceFunc,
		"reverse":         stdlib.ReverseListFunc,
		"setintersection": stdlib.SetIntersectionFunc,
		"setproduct":      stdlib.SetProductFunc,
		"setsubtract":     stdlib.SetSubtractFunc,
		"setunion":        stdlib.SetUnionFunc,
		"signum":          stdlib.SignumFunc,
		"slice":           stdlib.SliceFunc,
		"sort":            stdlib.SortFunc,
		"split":           stdlib.SplitFunc,
		"strrev":          stdlib.ReverseFunc,
		"substr":          stdlib.SubstrFunc,
		"timeadd":         stdlib.TimeAddFunc,
		"title":           stdlib.TitleFunc,
		"tostring":        stdlib.MakeToFunc(cty.String),
		"tonumber":        stdlib.MakeToFunc(cty.Number),
		"tobool":          stdlib.MakeToFunc(cty.Bool),
		"toset":           stdlib.MakeToFunc(cty.Set(cty.DynamicPseudoType)),
		"tolist":          stdlib.MakeToFunc(cty.List(cty.DynamicPseudoType)),
		"tomap":           stdlib.MakeToFunc(cty.Map(cty.DynamicPseudoType)),
		"trim":            stdlib.TrimFunc,
		"trimprefix":      stdlib.TrimPrefixFunc,
		"trimspace":       stdlib.TrimSpaceFunc,
		"trimsuffix":      stdlib.TrimSuffixFunc,
		"try":             tryfunc.TryFunc,
		"upper":           stdlib.UpperFunc,
		"values":          stdlib.ValuesFunc,
		"zipmap":          stdlib.ZipmapFunc,
	}
	//functions["templatefile"] = funcs.MakeTemplateFileFunc(
	//	cwd, func() map[string]function.Function {
	//		// The templatefile function prevents recursive calls to itself
	//		// by copying this map and overwriting the "templatefile" entry.
	//		return functions
	//	},
	//)
	//
	//functions["file"] = funcs.MakeFileFunc(cwd, false)
	//functions["fileexists"] = funcs.MakeFileExistsFunc(s.BaseDir)
	//functions["fileset"] = funcs.MakeFileSetFunc(s.BaseDir)
	//functions["filebase64"] = funcs.MakeFileFunc(s.BaseDir, true)
	//functions["filebase64sha256"] = funcs.MakeFileBase64Sha256Func(s.BaseDir)
	//functions["filebase64sha512"] = funcs.MakeFileBase64Sha512Func(s.BaseDir)
	//functions["filemd5"] = funcs.MakeFileMd5Func(s.BaseDir)
	//functions["filesha1"] = funcs.MakeFileSha1Func(s.BaseDir)
	//functions["filesha256"] = funcs.MakeFileSha256Func(s.BaseDir)
	//functions["filesha512"] = funcs.MakeFileSha512Func(s.BaseDir)

	return functions
}

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
		Functions: map[string]function.Function{
			"concat": stdlib.ConcatFunc,
		},
	}
	return ctx, nil
}
