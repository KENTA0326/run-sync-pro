package config

import (
	"strings"

	"github.com/spf13/viper"
)

// ResolveString は優先順位 CLI > ENV > default で文字列設定を返す。
func ResolveString(cliValue string, envKey string, defaultValue string) string {
	if s := strings.TrimSpace(cliValue); s != "" {
		return s
	}
	v := viper.New()
	v.SetDefault(envKey, defaultValue)
	v.BindEnv(envKey)
	v.AutomaticEnv()
	return v.GetString(envKey)
}
