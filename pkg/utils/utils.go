package utils

import (
	"github.com/samber/lo"
	"github.com/zclconf/go-cty/cty"
)

func MapValueToString(value cty.Value) []string {
	if value.IsNull() {
		return []string{}
	}
	return lo.Map[cty.Value, string](
		value.AsValueSlice(), func(val cty.Value, i int) string { return ValueToString(val) },
	)
}
func ValueToString(value cty.Value) string {
	if value.IsNull() {
		return ""
	}
	return value.AsString()
}
