package app

import (
	"bytes"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

// default env variables.
const (
	EnvActiveProfiles       = "ACTIVE_PROFILES"
	EnvSpringCloudConfigUri = "SPRING_CLOUD_CONFIG_URI"
	EnvConfigPath           = "CONFIG_PATH"
	EnvAppName              = "APP_NAME"
)

var (
	once              sync.Once
	applicationConfig *ApplicationConfig
	re                = regexp.MustCompile(`\$\{([A-Za-z0-9_-]+)((:)(.+?))?\}`)
)

type DecoderConfigOption func(*mapstructure.DecoderConfig)

type ApplicationConfig struct {
	viper *viper.Viper
}

func LoadApplicationConfig() *ApplicationConfig {
	v := viper.New()
	v.SetEnvPrefix("")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.SetDefault(EnvActiveProfiles, "")
	v.AutomaticEnv()
	v.SetConfigType("yaml")
	loadFile(v)
	loadSpringCloudConfig(v)
	v.AutomaticEnv()
	replaceConfigurationVariables(v)
	return &ApplicationConfig{
		viper: v,
	}
}

func replaceSingleVariable(v *viper.Viper, value string) string {
	groups := re.FindAllStringSubmatch(value, -1)
	for _, group := range groups {
		match := group[0]
		variable := group[1]
		defaultValue := group[4]
		variableValue := v.GetString(variable)
		if variableValue == "" {
			variableValue = defaultValue
		}

		value = strings.Replace(value, match, variableValue, -1)
	}
	return value
}

func replaceConfigurationVariables(v *viper.Viper) {
	for _, k := range v.AllKeys() {
		value := v.Get(k)
		switch value.(type) {
		case []interface{}:
			v.Set(k, replaceConfigurationVariablesInSlice(v, value.([]interface{})))
		case map[interface{}]interface{}:
			v.Set(k, replaceConfigurationVariablesInMap(v, value.(map[interface{}]interface{})))
		default:
			v.Set(k, replaceSingleVariable(v, fmt.Sprintf("%v", value)))
		}
	}
}

func replaceConfigurationVariablesInSlice(v *viper.Viper, slice []interface{}) []interface{} {
	for i, item := range slice {
		switch item.(type) {
		case string:
			slice[i] = replaceSingleVariable(v, item.(string))
		case map[interface{}]interface{}:
			slice[i] = replaceConfigurationVariablesInMap(v, item.(map[interface{}]interface{}))
		}
	}

	return slice
}

func replaceConfigurationVariablesInMap(v *viper.Viper, object map[interface{}]interface{}) map[interface{}]interface{} {
	for key, value := range object {
		switch value.(type) {
		case string:
			object[key] = replaceSingleVariable(v, value.(string))
		case []interface{}:
			object[key] = replaceConfigurationVariablesInSlice(v, value.([]interface{}))
		case map[interface{}]interface{}:
			object[key] = replaceConfigurationVariablesInMap(v, object)
		}
	}

	return object
}

// loadFile loads the configuration from a file.
func loadFile(v *viper.Viper) {
	v.SetDefault(EnvConfigPath, "./resources/")
	configPath := filepath.Dir(v.GetString(EnvConfigPath))
	v.SetConfigName("application")
	v.AddConfigPath(configPath)
	v.AddConfigPath("./config")
	v.AddConfigPath(".")
	err := v.MergeInConfig()
	if err != nil {
		fmt.Printf("unable to load config file application.yaml, error: %s\n", err)
	}
	profiles := v.GetString(EnvActiveProfiles)
	for _, p := range strings.Split(profiles, ",") {
		if p == "" {
			continue
		}
		v.SetConfigName(fmt.Sprintf("application-%s", p))
		err = v.MergeInConfig()
		if err != nil {
			fmt.Printf("unable to load config file application-%s.yaml, error: %s\n", p, err)
		}
	}
}

// loadSpringCloudConfig loads the configuration from a spring cloud configuration server.
func loadSpringCloudConfig(v *viper.Viper) {
	if !v.IsSet(EnvSpringCloudConfigUri) {
		return
	}

	v.SetDefault(EnvAppName, "app")

	serverURL := v.GetString(EnvSpringCloudConfigUri)
	profiles := v.GetString(EnvActiveProfiles)
	appName := v.GetString(EnvAppName)
	if profiles != "" {
		profiles = "-" + profiles
	}
	url := fmt.Sprintf("%s/%s%s.yaml", serverURL, appName, profiles)
	v.SetConfigName("remote")
	resp := fetchConfiguration(url)
	err := v.MergeConfig(bytes.NewBuffer(resp))
	if err != nil {
		fmt.Printf("could not read config at :%s, viper error: %v.\n", url, err)
	}
}

func toViperOpts(opts []DecoderConfigOption) []viper.DecoderConfigOption {
	var viperOpts []viper.DecoderConfigOption
	for _, vo := range opts {
		viperOpts = append(viperOpts, viper.DecoderConfigOption(vo))
	}
	return viperOpts
}

func (c *ApplicationConfig) Unmarshal(rawVal interface{}, opts ...DecoderConfigOption) error {
	return c.viper.Unmarshal(rawVal, toViperOpts(opts)...)
}

func (c *ApplicationConfig) UnmarshalKey(key string, rawVal interface{}, opts ...DecoderConfigOption) error {
	return c.viper.UnmarshalKey(key, rawVal, toViperOpts(opts)...)
}

func (c *ApplicationConfig) Settings() map[string]interface{} {
	return c.viper.AllSettings()
}

func (c *ApplicationConfig) GetString(key string) string {
	return c.viper.GetString(key)
}

func (c *ApplicationConfig) Get(key string) interface{} {
	return c.viper.Get(key)
}

func Config() *ApplicationConfig {
	once.Do(func() {
		applicationConfig = LoadApplicationConfig()
	})
	return applicationConfig
}

func (c *ApplicationConfig) Viper() *viper.Viper {
	return c.viper
}
