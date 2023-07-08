package utils

import (
	"encoding/json"
	"fmt"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"github.com/spf13/viper"
)

// LoadAdminConfig creates a caddy config from the viper config provided. This only contains the admin
// and logging portion of the configuration. The config loaders (`manager/configman`) will be responsible to load the
// configuration of the applications.
func LoadAdminConfig() (*caddy.Config, error) {
	logLevel := viper.GetString("caddy.log-level")
	persist := false
	return &caddy.Config{
		Admin: &caddy.AdminConfig{
			Disabled: true,
			Config: &caddy.ConfigSettings{
				Persist: &persist,
			},
		},
		Logging: &caddy.Logging{
			Logs: map[string]*caddy.CustomLog{
				"default": {
					Level: logLevel,
				},
			},
		},
	}, nil
}

// GetCaddyMatcherSet returns a caddy matcher set
func GetCaddyMatcherSet(path []string, methods []string, headers map[string][]string) []caddy.ModuleMap {
	// We will always need to match based on the path
	set := map[string]json.RawMessage{
		"path": GetByteStringArray(path...),
	}

	// Match on method if provided
	if len(methods) > 0 {
		set["method"] = GetByteStringArray(methods...)
	}

	// Match on headers if provided
	if len(headers) > 0 {
		data, _ := json.Marshal(headers)
		set["header"] = data
	}

	// Return the match set
	return []caddy.ModuleMap{set}
}

// GetByteStringArray returns an array of string in json form
func GetByteStringArray(val ...string) []byte {
	data, _ := json.Marshal(val)
	return data
}

// GetCaddyHandler returns a marshaled caddy handler config
func GetCaddyHandler(handlerName string, params map[string]interface{}) []json.RawMessage {
	handler := make(map[string]interface{})

	// Add the handler name / identifier
	handler["handler"] = fmt.Sprintf("sc_%s_handler", handlerName)

	// Add the params the handler needs
	for k, v := range params {
		handler[k] = v
	}

	data, _ := json.Marshal(handler)
	return []json.RawMessage{data}
}

// GetCaddySubrouter returns a marshaled caddy subrouter
func GetCaddySubrouter(routes ...caddyhttp.Route) []json.RawMessage {
	handler := map[string]interface{}{
		"handler": "subroute",
		"routes":  routes,
	}

	data, _ := json.Marshal(handler)
	return []json.RawMessage{data}
}
