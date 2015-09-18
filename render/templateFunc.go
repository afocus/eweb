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

func Tpl_FormatTime(t time.Time, layout string) string {
	return t.Format(layout)
}
