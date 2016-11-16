package eweb

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

type configSection map[string][]string

type configSections map[string]configSection

var configList = make(map[string]*configSections)

func GetConfig(name string) *configSections {
	if sections, has := configList[name]; has {
		return sections
	}
	var err error
	var f *os.File
	f, err = os.Open("config/" + name + ".cfg")
	if err != nil {
		panic("打开配置文件失败! error: " + err.Error())
	}
	defer f.Close()
	//解析ini文件
	r := bufio.NewReader(f)
	var line string
	var sec string
	sections := make(configSections)
	for err == nil {
		line, err = r.ReadString('\n')
		line = strings.TrimSpace(line)
		//空行或者注释跳过 注释支持;和#开头的行
		if line == "" || line[0] == ';' || line[0] == '#' {
			continue
		}
		//判断配置块[name]
		if line[0] == '[' && line[len(line)-1] == ']' {
			sec = line[1 : len(line)-1]
			_, has := sections[sec]
			if !has {
				sections[sec] = make(configSection)
			}
			continue
		}
		if sec == "" {
			continue
		}
		pair := strings.SplitN(line, "=", 2)
		if len(pair) != 2 {
			continue
		}
		key := strings.TrimSpace(pair[0])
		val := strings.TrimSpace(pair[1])
		if key == "" || val == "" {
			continue
		}
		if sections[sec][key] == nil {
			sections[sec][key] = make([]string, 0)
		}
		// 判断是否是数组
		if strings.HasSuffix(key, "[]") {
			sections[sec][key] = append(sections[sec][key], val)
		} else {
			sections[sec][key] = []string{val}
		}

	}
	configList[name] = &sections
	return configList[name]
}

func (this *configSections) GetString(sec, key, def string) string {
	m, ok := (*this)[sec]
	if !ok {
		return def
	}
	v, ok := m[key]
	if !ok {
		return def
	}
	return v[0]

}

func (this *configSections) GetSliceString(sec, key string) []string {
	m, ok := (*this)[sec]
	if !ok {
		return []string{}
	}
	v, ok := m[key]
	if !ok {
		return []string{}
	}
	return v
}

func (this *configSections) GetInt(sec, key string, def int) int {
	m, ok := (*this)[sec]
	if !ok {
		return def
	}
	v, ok := m[key]
	if !ok {
		return def
	}
	a := v[0]
	i, err := strconv.ParseInt(a, 10, 64)
	if err != nil {
		println(err.Error())
		return def
	}
	return int(i)
}

func (this *configSections) GetBool(sec, key string, def bool) bool {
	m, ok := (*this)[sec]
	if !ok {
		return def
	}
	v, ok := m[key]
	if !ok {
		return def
	}
	return v[0] != "0"
}
