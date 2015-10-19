package render

import (
	"html/template"
	"time"
)

//公共模板函数
var teamplateFuncs template.FuncMap

func init() {
	teamplateFuncs = template.FuncMap{
		"formatTime": Tpl_FormatTime,
	}
}

func Tpl_FormatTime(t interface{}, layout string) string {
	var t64 int64
	switch t.(type) {
	case int64:
		t64 = t.(int64)
	case int:
		t64 = int64(t.(int))
	default:
	}
	return time.Unix(t64, 0).Format(layout)
}
