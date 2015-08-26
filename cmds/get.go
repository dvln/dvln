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

// Package cmds get.go module implements the 'dvln get' subcommand
// framework for the 'cli' (aka: cobra) package.  Lets get some packages!!!
package cmds

import (
	cli "github.com/dvln/cobra"
	"github.com/dvln/out"
	globs "github.com/dvln/viper"
)

var getCmd = &cli.Command{
	Use:   "get",
	Short: "get packages",
	Long: `Get packages via a static or dynamic/generated devline, eg:
  % dvln get [ --codebase=cb_x ] [ --pkg=pkg_y ] [ --devline=dl_z ]
  % dvln get [ -c cb_x ] [ -p=pkg_y ] [ -d dl_z ]
  % dvln g [ -d dl_z ]    (set cfgfile:codebase|env:DVLN_CODEBASE, not req'd)`,
	Run: get,
}

// init bootstraps the options used for the get subcommand and descriptions
// and initial defaults for those options and such.
func init() {
	reloadCLIFlags := false
	setupGetCmdCLIArgs(getCmd, reloadCLIFlags)
}

// setupGetCmdCLIArgs is used from init() to set up the 'globs' (viper) pkg CLI
// options available to this subcommand (other options were already set up in
// the "parent" dvln subcommand in a like-named method). Every subcommand has
// a like named method "setup<subcmd>CmdCLIArgs()", called in init() above and
// called from dvln.go
func setupGetCmdCLIArgs(c *cli.Command, reloadCLIFlags bool) {
	var desc string
	if reloadCLIFlags {
		c.Flags().SetDefValueReparseOK(true)
	}
	desc, _, _ = globs.Desc("codebase")
	c.Flags().StringP("codebase", "c", globs.GetString("codebase"), desc)
	desc, _, _ = globs.Desc("devline")
	c.Flags().StringP("devline", "d", globs.GetString("devline"), desc)
	desc, _, _ = globs.Desc("pkg")
	c.Flags().StringP("pkg", "p", globs.GetString("pkg"), desc)
	desc, _, _ = globs.Desc("wkspcdir")
	c.Flags().StringP("wkspcdir", "w", globs.GetString("wkspcdir"), desc)
	c.Run = get
	// NewCLIOpts: if there were opts for the subcmd set them here and note that
	// "persistent" opts are set in cmds/dvln.go, only opts specific to the
	// 'dvln get' subcommand are set here
	// Note that you'll need to modify cmds/global.go as well otherwise your
	// globs.Desc() call and globs.GetBool("myopt") will not work.
	if reloadCLIFlags {
		c.Flags().SetDefValueReparseOK(false)
	}
}

// get defines the 'dvln get' sub-command in terms of it's options and making
// sure global config is setup correctly with all settings/controls the user
// requsted via the CLI
func get(cmd *cli.Command, args []string) {
	out.Debugln("Initialization done, firing up get()")
	wkspcRoot := globs.GetString("wkspcRoot")
	if wkspcRoot == "" {
		out.Println("Workspace root: no workspace yet exists")
	} else {
		out.Println("Workspace root:", globs.GetString("wkspcRoot"))
	}
	out.Println("Look up codebase")
	//codebase := cmd.Flags().Lookup("codebase").Value.String()
	codebase := globs.GetString("codebase")
	out.Println("Look up devline")
	//devline := cmd.Flags().Lookup("devline").Value.String()
	devline := globs.GetString("devline")
	out.Printf("Getting packages from codebase %s, devline %s\n", codebase, devline)
}
