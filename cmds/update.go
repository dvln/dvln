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

// Package cmds update.go module implements the 'dvln update' subcommand
// framework for the 'cli' (aka: cobra) package.  Lets update some packages!!!
package cmds

import (
	cli "github.com/dvln/cobra"
	"github.com/dvln/out"
	globs "github.com/dvln/viper"
)

var updateCmd = &cli.Command{
	Use:   "update",
	Short: "update/add/remove packages (dynamically or based on a devline)",
	Long: `Update/add/remove packages in a workspace dynamically or based ona devline, eg:
  % dvln update --devline=proj_x
  % dvln update -d proj_x
  % dvln u    (will update using versions from the workspaces current base devline)`,
	Run: update,
}

// init bootstraps the options used for the update subcommand and descriptions
// and initial defaults for those options and such.
func init() {
	reloadCLIFlags := false
	setupUpdateCmdCLIArgs(updateCmd, reloadCLIFlags)
}

// setupUpdateCmdCLIArgs is used from init() to set up the 'globs' (viper) pkg CLI
// options available to this subcommand (other options were already set up in
// the "parent" dvln subcommand in a like-named method). Every subcommand has
// a like named method "setup<subcmd>CmdCLIArgs()", called in init() above and
// called from dvln.go
func setupUpdateCmdCLIArgs(c *cli.Command, reloadCLIFlags bool) {
	var desc string
	if reloadCLIFlags {
		c.Flags().SetDefValueReparseOK(true)
	}
	desc, _, _ = globs.Desc("devline")
	c.Flags().StringP("devline", "d", globs.GetString("devline"), desc)
	desc, _, _ = globs.Desc("pkg")
	c.Flags().StringP("pkg", "p", globs.GetString("pkg"), desc)
	c.Run = update
	// NewCLIOpts: if there were opts for the subcmd set them here and note that
	// "persistent" opts are set in cmds/dvln.go, only opts specific to the
	// 'dvln update' subcommand are set here
	// Note that you'll need to modify cmds/global.go as well otherwise your
	// globs.Desc() call and globs.GetBool("myopt") will not work.
	if reloadCLIFlags {
		c.Flags().SetDefValueReparseOK(false)
	}
}

// update defines the 'dvln update' sub-command in terms of it's options and making
// sure global config is setup correctly with all settings/controls the user
// requsted via the CLI
func update(cmd *cli.Command, args []string) {
	out.Debugln("Initialization done, firing up update()")
	out.Println("Look up devline")
	//devline := cmd.Flags().Lookup("devline").Value.String()
	devline := globs.GetString("devline")
	out.Printf("Updateing packages based on devline %s\n", devline)
}
