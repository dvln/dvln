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
// globals.go file contains cfg (viper) package "globals" for the dvln
// tool
package cmds

//	"dvln/lib/out"
//	cfg "dvln/lib/3rd/viper"
import (
	"github.com/dvln/out"
	cfg "github.com/spf13/viper"
)

// initAppDefaultSettings sets up default settings for any variables used
// throughout the dvln tool... "globals" so to speak. These will be
// stashed in the 'cfg' (viper) package at the default level (lowest
// priority essentially) and can be overriden via config file, CLI
// flags, sometimes codebase level overrides, etc
//
// Note: this contains *all* app defaults regardless of top level dvln command
// or subcommands or dvln focused pkg/library "globals" (however, any package
// that is targeted at being generic/standalone should NOT use this as it will
// no longer be generic as it couldn't be used without this cmds package)
// - eriknow: this could be moved to lib/globs/dvln.go potentially or we could
//            put all dvln specific libs in just lib so lib/dvlnglobs.go in
//            which case the pkg would be 'dvlnlib' or something like that?,
//            and all those are in the same VCS 'pkg' (import 'dvlnlib') and
//            no "generic" sub-packages (eg: out, 3rd/*, etc) should use
//            anything from within 'dvlnlib'.
//
// Note: for any new CLI focused option you need to modify cmds/dvln.go
//       so pushCLIOptsTofg() pushes the CLI option into the 'cfg' (viper)
//       package... otherwise you're stuck with the CLI not working  ;)... and
//       yes, that should be fixed (as flags should be a 1st class citizen
//       and not have to use cfg.Set() to push them into cfg/viper, ugh)
func initAppDefaultSettings() {
	// Note: if you want aliases for keys you can add them like so, note
	//       that cfg (viper) is "case independent" so Taxonomies and
	//       taxonomies are identical as far as 'cfg' is concerned

	// cfg.SetDefault("Taxonomies", map[string]string{"tag": "tags", "category": "categories"})
	// cfg.RegisterAlias("indexes", "taxonomies")
	// NewSubCommand: if you have a new subcommand with new CLI options you'll
	// want to add a variable for it here and set up default settings,
	// description and additional data such as user level that will use
	// the option/variable and how the variable can be set.

	// Note: this is currently in sections related to the scope of how the
	//       variable can be set, feel free to set subsections within those
	//       sections if needed (eg: path variables, clitool name vars)...
	//       essentially any grouping you see fit at this point but try and
	//       at least get the top level Section right

	// Section: CLIGlobal class options, vars that can come in from the CLI
	// - please add them alphabetically and don't reuse existing opts/vars
	cfg.SetDefault("analysis", false)
	cfg.SetDesc("analysis", "memory and timing analytics", cfg.ExpertUser, cfg.CLIGlobal)

	cfg.SetDefault("codebase", "") // no default code base to start with
	cfg.SetDesc("codebase", "codebase name or URL", cfg.NoviceUser, cfg.CLIGlobal)

	cfg.SetDefault("config", "~/.dvlncfg/") // defaults to .dvlncfg/config.json|yaml|..
	cfg.SetDesc("config", "file|path, path scans cfg.json|toml|yml", cfg.ExpertUser, cfg.CLIGlobal)

	cfg.SetDefault("debug", false)
	cfg.SetDesc("debug", "control debug output", cfg.NormalUser, cfg.CLIGlobal)

	cfg.SetDefault("devline", "") // no default devline to start with
	cfg.SetDesc("devline", "development line name", cfg.NoviceUser, cfg.CLIGlobal)

	cfg.SetDefault("fatalon", 1) // exits on 1st VCS error
	cfg.SetDesc("fatalon", "# of VCS clone errs to choke on", cfg.ExpertUser, cfg.CLIGlobal)

	cfg.SetDefault("force", false) // fail on dangerous ops
	cfg.SetDesc("force", "force bypass of protections", cfg.ExpertUser, cfg.CLIGlobal)

	cfg.SetDefault("interact", false) // the default is no user interaction
	cfg.SetDesc("interact", "control client prompting", cfg.NormalUser, cfg.CLIGlobal)

	cfg.SetDefault("jobs", "all") // default: use all CPU's
	cfg.SetDesc("jobs", "# of CPU's to use for jobs", cfg.ExpertUser, cfg.CLIGlobal)

	cfg.SetDefault("look", "text") // text or json
	cfg.SetDesc("look", "output look, text|json", cfg.ExpertUser, cfg.CLIGlobal)

	cfg.SetDefault("pkg", "") // no default package(s) to start with
	cfg.SetDesc("pkg", "package selector, comma separated", cfg.NoviceUser, cfg.CLIGlobal)

	cfg.SetDefault("port", 3856) // port when serving
	cfg.SetDesc("port", "port # for --serve mode", cfg.ExpertUser, cfg.CLIGlobal)

	cfg.SetDefault("quiet", false) // normal output to start
	cfg.SetDesc("quiet", "silent running", cfg.NormalUser, cfg.CLIGlobal)

	cfg.SetDefault("record", "off") // no output log to start
	cfg.SetDesc("record", "log to file or 'tmp'", cfg.NoviceUser, cfg.CLIGlobal)

	cfg.SetDefault("serve", false) // serve defaults off
	cfg.SetDesc("serve", "activate REST serve mode", cfg.ExpertUser, cfg.CLIGlobal)

	cfg.SetDefault("terse", false) // regular non-terse mode
	cfg.SetDesc("terse", "output brevity", cfg.NormalUser, cfg.CLIGlobal)

	cfg.SetDefault("verbose", false) // not verbose to start
	cfg.SetDesc("verbose", "output verbosity, extends debug", cfg.NormalUser, cfg.CLIGlobal)

	cfg.SetDefault("version", false)
	cfg.SetDesc("version", "show tool version details", cfg.NormalUser, cfg.CLIGlobal)

	cfg.SetDefault("wkspcdir", ".") // assume current dir is where workspace is
	cfg.SetDesc("wkspcdir", "workspace directory", cfg.NormalUser, cfg.CLIGlobal)


	// Section: BasicGlobal variables to store data (env, config file, default)
	// - please add them alphabetically and don't reuse existing opts/vars
	cfg.SetDefault("logfileLevel", int(out.LevelInfo)) // default (if activate)
	cfg.SetDesc("logfileLevel", "log file output level (used if logging on)", cfg.ExpertUser, cfg.BasicGlobal)

	cfg.SetDefault("screenLevel", int(out.LevelInfo)) // default print lvl
	cfg.SetDesc("screenLevel", "screen output level", cfg.ExpertUser, cfg.BasicGlobal)


	// Section: ConstGlobal variables to store data (default value only, no overrides)
	// - please add them alphabetically and don't reuse existing opts/vars
	cfg.SetDefault("dvlnToolVer", "0.0.1") // current version of the dvln tool
	cfg.SetDesc("dvlnToolVer", "current version of the dvln tool", cfg.InternalUse, cfg.ConstGlobal)

	// Section: <add more sections as needed>
}
