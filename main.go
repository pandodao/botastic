/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package main

import (
	"context"
	"os"

	_ "github.com/lib/pq"
	"github.com/pandodao/botastic/cmd/root"
	"github.com/pandodao/botastic/session"

	"github.com/spf13/cobra"
)

var (
	version = "2.0.0"
)

func main() {
	ctx := context.Background()
	s := &session.Session{Version: version}
	ctx = session.With(ctx, s)

	expandedArgs := []string{}
	if len(os.Args) > 0 {
		expandedArgs = os.Args[1:]
	}

	rootCmd := root.NewCmdRoot(version)

	rootCmd.SetArgs(expandedArgs)
	if err := rootCmd.ExecuteContext(ctx); err != nil {
		rootCmd.PrintErrln("execute failed:", err)
		os.Exit(1)
	}
}

func hasCommand(rootCmd *cobra.Command, args []string) bool {
	c, _, err := rootCmd.Traverse(args)
	return err == nil && c != rootCmd
}
