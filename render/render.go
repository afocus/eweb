package render

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
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
		debug:    isdebug,
	}
}

type Render struct {
	t        *template.Template
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

func getFileList(path string) []string {
	var files = make([]string, 0)
	err := filepath.Walk(path, func(path string, f os.FileInfo, err error) error {
		if f == nil {
			return err
		}
		if f.IsDir() {
			return nil
		}
		files = append(files, path)
		return nil
	})
	if err != nil {
		panic(err)
	}
	return files
}

func getTemplateIns(root string) *template.Template {
	tplFils := getFileList(root)
	t := template.New("views").Funcs(teamplateFuncs)
	for _, path := range tplFils {
		b, err := ioutil.ReadFile(path)
		if err != nil {
			panic(err)
		}
		//如果是windows下编译的 还"\"转为"/"
		name := string([]byte(filepath.ToSlash(path))[len(root):])
		t, err = t.New(name).Parse(string(b))
		println("parse tpl>>", name)
		if err != nil {
			panic(err)
		}
	}
	return t
}

func (this *Render) Html(w http.ResponseWriter, code int, tplname string,
	data interface{}, controlname string) error {
	var t *template.Template
	if this.debug {
		t = getTemplateIns(this.basePath)
	} else {
		if this.t == nil {
			this.t = getTemplateIns(this.basePath)
			t = this.t
		}
	}

	w.Header().Set("Content-type", CONTENTTYPE_HTML)
	w.WriteHeader(code)
	err := t.ExecuteTemplate(w, tplname, data)
	return err
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
