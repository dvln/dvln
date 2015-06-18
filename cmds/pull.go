// Copyright © 2015 Erik Brady <brady@dvln.org>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package cmds pull.go module implements the 'dvln pull' subcommand
// framework for the 'cli' (aka: cobra) package.  Lets pull some packages!!!
package cmds

import (
	"github.com/dvln/out"
	cli "github.com/dvln/cobra"
	globs "github.com/dvln/viper"
)

var pullCmd = &cli.Command{
	Use:   "pull",
	Short: "pull/add/remove packages using current or given devline",
	Long: `Pull/add/remove packages in a workspace using the current or a specified devline, eg:
  % dvln pull --devline=proj_x
  % dvln pull -d proj_x
  % dvln u    (will pull using versions from the workspaces current base devline)`,
	Run: pull,
}

// init bootstraps the options used for the pull subcommand and descriptions
// and initial defaults for those options and such.
func init() {
	reloadCLIFlags := false
	setupPullCmdCLIArgs(reloadCLIFlags)
}

// setupPullCmdCLIArgs is used from init() to set up the 'globs' (viper) pkg CLI
// options available to this subcommand (other options were already set up in
// the "parent" dvln subcommand in a like-named method). Every subcommand has
// a like named method "setup<subcmd>CmdCLIArgs()", called in init() above and
// called from dvln.go
func setupPullCmdCLIArgs(reloadCLIFlags bool) {
	var desc string
	if reloadCLIFlags {
		pullCmd.Flags().SetDefValueReparseOK(true)
	}
	//desc, _, _ = globs.Desc("codebase")
	//pullCmd.Flags().StringP("codebase", "c", globs.GetString("codebase"), desc)
	desc, _, _ = globs.Desc("devline")
	pullCmd.Flags().StringP("devline", "d", globs.GetString("devline"), desc)
	desc, _, _ = globs.Desc("pkg")
	pullCmd.Flags().StringP("pkg", "p", globs.GetString("pkg"), desc)
	pullCmd.Run = pull
	// NewCLIOpts: if there were opts for the subcmd set them here and note that
	// "persistent" opts are set in cmds/dvln.go, only opts specific to the
	// 'dvln pull' subcommand are set here
	// Note that you'll need to modify cmds/global.go as well otherwise your
	// globs.Desc() call and globs.GetBool("myopt") will not work.
	if reloadCLIFlags {
		pullCmd.Flags().SetDefValueReparseOK(false)
	}
}

// pull defines the 'dvln pull' sub-command in terms of it's options and making
// sure global config is setup correctly with all settings/controls the user
// requsted via the CLI
func pull(cmd *cli.Command, args []string) {
	out.Debugln("Initialization done, firing up pull()")
	out.Println("Look up devline")
	//devline := cmd.Flags().Lookup("devline").Value.String()
	devline := globs.GetString("devline")
	out.Printf("Pulling packages based on devline %s\n", devline)
}
