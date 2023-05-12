package cmd

import (
	"fmt"
	"runtime/debug"

	"github.com/spf13/cobra"
)

var (
	versionRevision = "unknown"
	versionTime     = "unknown"
	versionModified = "unknown"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Display version info",
	Run: func(cmd *cobra.Command, args []string) {
		m := map[string]string{
			"revision": versionRevision,
			"time":     versionTime,
			"modified": versionModified,
		}
		for k, v := range m {
			fmt.Printf("%s: %s\n", k, v)
		}
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)

	info, ok := debug.ReadBuildInfo()
	if !ok {
		panic("no version info")
	}

	m := map[string]func(string){
		"vcs.revision": func(s string) { versionRevision = s },
		"vcs.time":     func(s string) { versionTime = s },
		"vcs.modified": func(s string) { versionModified = s },
	}
	for _, kv := range info.Settings {
		if f, ok := m[kv.Key]; ok {
			f(kv.Value)
		}
	}
}
