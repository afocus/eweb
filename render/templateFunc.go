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
		"math":       Tpl_Math,
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

func Tpl_Math(method string, a, b interface{}) float64 {
	var num1 float64
	var num2 float64
	switch a.(type) {
	case int64:
		num1 = float64(a.(int64))
		num2 = float64(b.(int64))
	case float64:
		num1 = a.(float64)
		num2 = b.(float64)
	case float32:
		num1 = float64(a.(float32))
		num2 = float64(b.(float32))
	case int:
		num1 = float64(a.(int))
		num2 = float64(b.(int))
	case uint:
		num1 = float64(a.(uint))
		num2 = float64(b.(uint))
	default:
		return 0
	}
	switch method {
	case "+":
		return num1 + num2
	case "-":
		return num1 - num2
	case "*":
		return num1 * num2
	case "/":
		return num1 / num2
	default:
		return 0
	}
}
