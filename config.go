package eweb

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

var configList map[string]*Config

type Config struct {
	path string
	data map[string]interface{}
}

//加载已存在的配置文件
func GetConfig(name string) *Config {
	if config, has := configList[name]; has {
		return config
	}
	return nil
}

//如果不存在则直接加载
func MustGetConfig(name string) *Config {
	config := GetConfig(name)
	if config != nil {
		return config
	}
	return LoadConfig(name)
}
func LoadConfig(name string) *Config {
	if _, has := configList[name]; has {
		panic("已加载相同名称的配置文件")
	}
	path := "config/" + name + ".json"
	fi, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer fi.Close()
	fd, err := ioutil.ReadAll(fi)
	if err != nil {
		panic(err)
	}
	var data map[string]interface{}
	err = json.Unmarshal(fd, &data)
	if err != nil {
		panic(err)
	}
	if configList == nil {
		configList = make(map[string]*Config)
	}
	config := &Config{path, data}
	configList[name] = config
	return config
}

func (this *Config) GetInt(key string) int {
	if v, has := this.data[key]; has {
		return int(v.(float64))
	}
	return 0
}
func (this *Config) SetInt(key string, value int) {
	this.data[key] = value
}

func (this *Config) GetString(key string) string {
	if v, has := this.data[key]; has {
		return v.(string)
	}
	return ""
}
func (this *Config) SetString(key, value string) {
	this.data[key] = value
}

func (this *Config) GetBool(key string) bool {
	if v, has := this.data[key]; has {
		return v.(bool)
	}
	return false
}

func (this *Config) SetBool(key string, value bool) {
	this.data[key] = value
}

//吧当前的配置保存到磁盘上
func (this *Config) Save() bool {
	return false
}

func (this *Config) GetPath() string {
	return this.path
}
