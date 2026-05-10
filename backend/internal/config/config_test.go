package config_test

import (
	"testing"

	"github.com/KENTA0326/run-sync-pro/internal/config"
)

func TestResolveString_table(t *testing.T) {
	// t.Setenv と t.Parallel は併用不可のため、このテーブルは直列実行する。
	cases := []struct {
		name       string
		cli        string
		envKey     string
		setEnv     bool
		envVal     string
		defaultVal string
		want       string
	}{
		{
			name:       "cli_wins_over_env_and_default",
			cli:        "from-cli",
			envKey:     "RUNSYNCPRO_RESOLVE_A",
			setEnv:     true,
			envVal:     "from-env",
			defaultVal: "from-default",
			want:       "from-cli",
		},
		{
			name:       "trimmed_cli_empty_falls_through_to_env",
			cli:        "   ",
			envKey:     "RUNSYNCPRO_RESOLVE_B",
			setEnv:     true,
			envVal:     "from-env",
			defaultVal: "from-default",
			want:       "from-env",
		},
		{
			name:       "env_wins_over_default",
			cli:        "",
			envKey:     "RUNSYNCPRO_RESOLVE_C",
			setEnv:     true,
			envVal:     "only-env",
			defaultVal: "from-default",
			want:       "only-env",
		},
		{
			name:       "default_when_env_absent",
			cli:        "",
			envKey:     "RUNSYNCPRO_RESOLVE_SHOULD_NOT_EXIST_XYZ",
			setEnv:     false,
			envVal:     "",
			defaultVal: "fallback",
			want:       "fallback",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if tc.setEnv {
				t.Setenv(tc.envKey, tc.envVal)
			}
			got := config.ResolveString(tc.cli, tc.envKey, tc.defaultVal)
			if got != tc.want {
				t.Fatalf("ResolveString(...)=%q want %q", got, tc.want)
			}
		})
	}
}
