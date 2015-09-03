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

// Package cmds defines and implements command-line commands and flags used by
// dvln.  Commands and flags are implemented using the cobra CLI commander
// library (dvln/lib/3rd/cobra) which will be imported under "cli".  The
// globals.go file contains globs (viper) package "globals" for the dvln
// tool
package cmds

import (
	"fmt"
	"path/filepath"

	"github.com/dvln/out"
	globs "github.com/dvln/viper"
)

// doPkgCmdGlobsInit sets up default settings for any variables/opts used
// for the dvln cmds pkg... "globals" so to speak.  These are currently
// stashed in the 'globs' (viper) package at the default level (lowest
// priority essentially) and can be overriden via config file, CLI
// flags, and codebase, package and devline level overrides (in time).
// - Any non-generic package can use globs to store "globals" which are then
//   effectively visible across all dvln commands and non-generic packages
// - Generic packages (eg: lib/dvln/out, or: github.com/dvln/out) should *NOT*
//   use globs (viper) for config so as to remain extremely generic (and usable
//   by the global community)... for such generic packages (or 3rd party/vendor
//   packages dvln is using the dvln or dvln get cmd can use globs (vipers) to
//   get DVLN configuration or opts/etc... and then it can use API's for those
//   generic package (or package variables) to config that pkg for dvln's use
//   as one can see with dvln/cmd/dvln.go init'ing the 'out' pkg on startup.
// - You will find that other packages (dvln/lib) that are not generic will
//   have more "global" settings, like these, for their own needs so as to
//   keep the pkg data "close" to that package (currently), for example see
//   the dvln/lib/json.go or dvln/lib/dvlnver.go files and their init() fcns.
//
// Aside: for any new CLI focused option for the dvln meta-cmd check out
//       setupDvlnCmdCLIArgs() in cmds/dvln.go and for subcommands see related
//       setup<Name>CmdCLIArgs() located within each cmds/<name>.go file, I'd
//       suggest searching on the string "NewCLIOpts" to find those locations.
func doPkgCmdGlobsInit() {
	// Note: if you want aliases for keys you can add them like so, note
	//       that "globs" (viper) is "case independent" so Taxonomies and
	//       taxonomies are identical as far as 'globs' is concerned

	// globs.SetDefault("Taxonomies", map[string]string{"tag": "tags", "category": "categories"})
	// globs.RegisterAlias("indexes", "taxonomies")
	// NewSubCommand: if you have a new subcommand with new CLI options you'll
	// want to add a variable for it here and set up default settings,
	// description and additional data such as user level that will use
	// the option/variable and how the variable can be set.

	// Note: this is currently in sections related to the scope of how the
	//       variable can be set, feel free to set subsections within those
	//       sections if needed (eg: path variables, clitool name vars)...
	//       essentially any grouping you see fit at this point but try and
	//       at least get the top level Section right

	// Section: ConstGlobal variables to store data (default value only, no overrides)
	// - please add them alphabetically and don't reuse existing opts/vars
	globs.SetDefault("wkspcMetaDir", ".dvln")
	globs.SetDesc("wkspcMetaDir", "where dvln config info exists in a workspace", globs.InternalUse, globs.ConstGlobal)

	// Section: BasicGlobal variables to store data (env, config file, default)
	// - please add them alphabetically and don't reuse existing opts/vars
	globs.SetDefault("logfilelevel", fmt.Sprintf("%s", out.LevelInfo)) // default log lvl (if activate)
	globs.SetDesc("logfilelevel", "log file output level (used if logging on)", globs.ExpertUser, globs.BasicGlobal)

	globs.SetDefault("screenlevel", fmt.Sprintf("%s", out.LevelInfo)) // default print lvl
	globs.SetDesc("screenlevel", "screen output level", globs.ExpertUser, globs.BasicGlobal)

	// Section: CLIGlobal class options, vars that can come in from the CLI
	// - please add them alphabetically and don't reuse existing opts/vars
	//
	// NewCLIOpts: if there were opts for the dvln meta-command or subcmds you
	// would define their starting default value and meta-data below.  Note that
	// for CLI opts you would need the last field to be something like the one
	// in the section below (globs.CLIGlobal) or some other scope that indicates
	// they can be set via the CLI (which is what the below "block" is for, but
	// if yours is special and maybe can't be set in the config file or
	// something special like that you might need another block to put em in).
	// Please add things alphabetically within the appropriate section.
	// - note: currently this contains CLIGlobal and CLIOnlyGlobal type flags
	globs.SetDefault("analysis", false)
	globs.SetDesc("analysis", "memory and timing analytics", globs.ExpertUser, globs.CLIGlobal)

	globs.SetDefault("codebase", "") // no default code base to start with
	globs.SetDesc("codebase", "codebase name or URL", globs.NoviceUser, globs.CLIGlobal)

	cfgDir := filepath.Join("~", ".dvlncfg")
	globs.SetDefault("config", cfgDir) // defaults to ~/.dvlncfg
	globs.SetDesc("config", "tool config dir|file", globs.ExpertUser, globs.CLIGlobal)

	globs.SetDefault("debug", false)
	globs.SetDesc("debug", "control debug output", globs.StandardUser, globs.CLIGlobal)

	globs.SetDefault("devline", "") // no default devline to start with
	globs.SetDesc("devline", "development line name", globs.NoviceUser, globs.CLIGlobal)

	globs.SetDefault("fatalon", 1) // exits on 1st VCS error
	globs.SetDesc("fatalon", "# of VCS errs needed to cause exit", globs.ExpertUser, globs.CLIGlobal)

	globs.SetDefault("force", false) // fail on dangerous ops
	globs.SetDesc("force", "force bypass of protections", globs.ExpertUser, globs.CLIGlobal)

	globs.SetDefault("globs", "") // show available cfg|env settings to user
	globs.SetDesc("globs", "show settings available, cfg|env", globs.ExpertUser, globs.CLIOnlyGlobal)

	globs.SetDefault("interact", false) // the default is no user prompting
	globs.SetDesc("interact", "prompting control", globs.StandardUser, globs.CLIGlobal)

	globs.SetDefault("jobs", "all") // default: use all CPU's
	globs.SetDesc("jobs", "# of CPU's to use for jobs", globs.ExpertUser, globs.CLIGlobal)

	globs.SetDefault("look", "text") // text or json
	globs.SetDesc("look", "output look, text|json", globs.ExpertUser, globs.CLIGlobal)

	globs.SetDefault("pkg", "") // no default package(s) to start with
	globs.SetDesc("pkg", "package selector, comma separated", globs.NoviceUser, globs.CLIOnlyGlobal)

	globs.SetDefault("port", 3856) // port when serving
	globs.SetDesc("port", "port # for --serve mode", globs.ExpertUser, globs.CLIGlobal)

	globs.SetDefault("quiet", false) // normal output to start
	globs.SetDesc("quiet", "silent running", globs.StandardUser, globs.CLIGlobal)

	globs.SetDefault("record", "off") // no output record/log to start
	globs.SetDesc("record", "to file|'tmp'", globs.NoviceUser, globs.CLIGlobal)

	globs.SetDefault("serve", false) // serve defaults off
	globs.SetDesc("serve", "activate REST serve mode", globs.ExpertUser, globs.CLIGlobal)

	globs.SetDefault("terse", false) // regular non-terse mode
	globs.SetDesc("terse", "output reduction", globs.StandardUser, globs.CLIGlobal)

	globs.SetDefault("verbose", false) // not verbose to start
	globs.SetDesc("verbose", "output verbosity, extends debug", globs.StandardUser, globs.CLIGlobal)

	globs.SetDefault("version", false)
	globs.SetDesc("version", "show tool version details", globs.StandardUser, globs.CLIOnlyGlobal)

	globs.SetDefault("wkspcdir", ".") // assume current dir is where workspace is
	globs.SetDesc("wkspcdir", "workspace directory", globs.StandardUser, globs.CLIOnlyGlobal)

	// Section: <add more sections as needed>

}
