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
	"github.com/dvln/out"
	cli "github.com/spf13/cobra"
	cfg "github.com/spf13/viper"
)

var getCmd = &cli.Command{
	Use:   "get",
	Short: "get packages for a codebase [+ devline]",
	Long: `Get packages for a codebase [+ devline], create/modify a workspace, eg:
  % dvln get --codebase=prod_x --devline=proj_x
  % dvln get -c prod_x -d proj_x
  % dvln g -d proj_x    (if cfg:CodeBase or env:DVLN_CODEBASE set)`,
	Run: get,
}

// init bootstraps the options used for the get subcommand and descriptions
// and initial defaults for those options and such.
func init() {
	reloadCLIFlags := false
	setupGetCmdCLIArgs(reloadCLIFlags)
}

// setupGetCmdCLIArgs is used from init() to set up the 'cfg' (viper) pkg CLI
// options available to this subcommand (other options were already set up in
// the "parent" dvln subcommand in a like-named method, every subcommand has
// a like named method "setup<subcmd>CmdCLIArgs()"
func setupGetCmdCLIArgs(reloadCLIFlags bool) {
	var desc string
	if reloadCLIFlags {
		getCmd.Flags().SetDefValueReparseOK(true)
	}
	desc, _, _ = cfg.Desc("codebase")
	getCmd.Flags().StringVarP(&cliCodeBase, "codebase", "c", cfg.GetString("codebase"), desc)
	desc, _, _ = cfg.Desc("devline")
	getCmd.Flags().StringVarP(&cliDevLine, "devline", "d", cfg.GetString("devline"), desc)
	desc, _, _ = cfg.Desc("interact")
	getCmd.Flags().BoolVarP(&cliInteract, "interact", "i", cfg.GetBool("interact"), desc)
	desc, _, _ = cfg.Desc("pkg")
	getCmd.Flags().StringVarP(&cliPkg, "pkg", "p", cfg.GetString("pkg"), desc)
	desc, _, _ = cfg.Desc("wkspcdir")
	getCmd.Flags().StringVarP(&cliWkspcDir, "wkspcdir", "w", cfg.GetString("wkspcdir"), desc)
	getCmd.Run = get
	if reloadCLIFlags {
		getCmd.Flags().SetDefValueReparseOK(false)
	}
}

// pushGetCmdCLIOptsToCfg shoves any user set CLI options into the 'cfg'
// (viper) package if it was used on the CLI (so viper has a full picture
// of all variables, their defaults, any env settings, and now any CLI
// settings as well (amongst other settings)
func pushGetCmdCLIOptsToCfg() {
	// local flags for get subcmd bootstrapped here
	if getCmd.Flags().Lookup("codebase").Changed {
		cfg.Set("codebase", cliCodeBase)
	}
	if getCmd.Flags().Lookup("devline").Changed {
		cfg.Set("devline", cliDevLine)
	}
	if getCmd.Flags().Lookup("interact").Changed {
		cfg.Set("interact", cliInteract)
	}
	if getCmd.Flags().Lookup("pkg").Changed {
		cfg.Set("pkg", cliPkg)
	}
	if getCmd.Flags().Lookup("wkspcdir").Changed {
		cfg.Set("wkspcdir", cliWkspcDir)
	}
}

// get defines the 'dvln get' sub-command in terms of it's options and making
// sure global config is setup correctly with all settings/controls the user
// requsted via the CLI
func get(cmd *cli.Command, args []string) {
	out.Debugln("Initialization done, firing up get()")
	out.Println("Look up codebase")
	codebase := cmd.Flags().Lookup("codebase").Value.String()
	out.Println("Look up devline")
	devline := cmd.Flags().Lookup("devline").Value.String()
	out.Printf("Getting packages from codebase %s, devline %s\n", codebase, devline)
}
