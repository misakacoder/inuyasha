package configs

import (
	"bytes"
	_ "embed"
	"github.com/jinzhu/configor"
	"github.com/misakacoder/inuyasha/pkg/db/orm"
	"github.com/misakacoder/kagome/file"
	"github.com/misakacoder/kagome/maps"
	"github.com/misakacoder/kagome/str"
	"github.com/misakacoder/logger"
	"gopkg.in/yaml.v3"
	"path/filepath"
	"reflect"
	"sync"
)

const (
	configDir  = "configs"
	configName = "application.yml"
)

var (
	Config    = configuration{}
	listeners = []Listener{{
		Config: &Config,
		Reload: func(config any) {
			level, _ := logger.Parse(Config.Log.Level)
			logger.SetLevel(level)
		},
	}}
	once sync.Once
)

type Listener struct {
	Config any
	Reload func(config any)
}

type configuration struct {
	Server struct {
		Bind string
		Port int
	}
	Db  orm.Config
	Log struct {
		Directory string
		Level     string
	}
	Swagger struct {
		Enabled bool
		Auth    struct {
			Enabled  bool
			Username string
			Password string
		}
	}
	Pprof struct {
		Enabled  bool
		Username string
		Password string
	}
}

func AddListener(config any, reload func(config any)) {
	rt := reflect.TypeOf(config)
	if rt.Kind() != reflect.Ptr || rt.Elem().Kind() != reflect.Struct {
		logger.Panic("config must be a pointer to struct")
	}
	listeners = append(listeners, Listener{Config: config, Reload: reload})
}

func ListenConfig() {
	once.Do(func() {
		configFilepath := filepath.Join(configDir, configName)
		if !file.ExistFile(configFilepath) {
			var configs []map[string]interface{}
			for _, v := range listeners {
				config := v.Config
				data, _ := yaml.Marshal(config)
				dataMap := map[string]any{}
				yaml.Unmarshal(data, &dataMap)
				for _, mp := range configs {
					for key := range mp {
						if dataValue, ok := dataMap[key]; ok {
							maps.MergeMap(map[string]interface{}{key: dataValue}, mp)
							delete(dataMap, key)
						}
					}
				}
				if len(dataMap) > 0 {
					configs = append(configs, dataMap)
				}
			}
			result := str.NewJoiner("\n", "", "")
			for _, config := range configs {
				var buffer bytes.Buffer
				encoder := yaml.NewEncoder(&buffer)
				encoder.SetIndent(2)
				encoder.Encode(config)
				encoder.Close()
				result.Append(buffer.String())
			}
			file.WriteFile(configFilepath, []byte(result.String()))
			logger.Panic("please modify the config in the path %s", configFilepath)
		}
		for _, v := range listeners {
			config := v.Config
			reload := v.Reload
			conf := &configor.Config{
				AutoReload: true,
				AutoReloadCallback: func(config any) {
					if reload != nil {
						reload(config)
					}
				},
			}
			err := configor.New(conf).Load(config, configFilepath)
			if err != nil {
				logger.Panic(err.Error())
			}
		}
	})
}
