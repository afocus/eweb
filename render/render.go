package render

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"html/template"
	"net/http"
)

const (
	CONTENTTYPE_TEXT = "text/plain"
	CONTENTTYPE_JSON = "application/json"
	CONTENTTYPE_XML  = "application/xml"
	CONTENTTYPE_HTML = "text/html"
)

func New(path string, isdebug bool) *Render {
	return &Render{
		basePath: path,
		list:     make(map[string]*template.Template, 0),
		debug:    isdebug,
	}
}

type Render struct {
	list     map[string]*template.Template
	basePath string
	debug    bool
}

func (this *Render) RegisterFuncs(funcs template.FuncMap) {
	for k, v := range funcs {
		if _, has := teamplateFuncs[k]; has {
			panic(errors.New("与已存在的模板函数重名 >>" + k))
		}
		teamplateFuncs[k] = v
	}
}
func (Render) Render(w http.ResponseWriter, typestr string, code int, data []byte) error {
	if len(typestr) == 0 {
		typestr = CONTENTTYPE_TEXT
	}
	w.WriteHeader(code)
	w.Header().Set("Content-Type", typestr)
	_, err := w.Write(data)
	return err
}

func (this *Render) Html(w http.ResponseWriter, code int, tplname string,
	data interface{}, controlname string) error {
	var t *template.Template
	if this.debug {
		t = template.Must(template.New(controlname).Funcs(teamplateFuncs).
			ParseGlob(this.basePath + "/" + controlname + "/*"))
	} else {
		//这部分应该加锁 稍后处理
		var has bool
		if t, has = this.list[controlname]; !has {
			t = template.Must(template.New(controlname).Funcs(teamplateFuncs).
				ParseGlob(this.basePath + "/" + controlname + "/*"))
			this.list[controlname] = t
		}
	}
	w.WriteHeader(code)
	w.Header().Set("Content-type", CONTENTTYPE_HTML)
	return t.ExecuteTemplate(w, tplname, data)
}

func (this *Render) Json(w http.ResponseWriter, code int, data interface{}) error {
	val, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return this.Render(w, CONTENTTYPE_JSON, code, val)
}

func (this *Render) Xml(w http.ResponseWriter, code int, data interface{}) error {
	val, err := xml.Marshal(data)
	if err != nil {
		return err
	}
	//xml默认不添加头部 这里补上
	xmls := []byte(xml.Header)
	return this.Render(w, CONTENTTYPE_XML, code, append(xmls, val...))
}
