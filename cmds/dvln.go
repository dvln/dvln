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
// library (dvln/lib/3rd/cobra) which will be imported under "cli".
package cmds

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/dvln/api"
	"github.com/dvln/lib"
	"github.com/dvln/out"
	"github.com/kr/pretty"
	"github.com/spf13/cast"
	cli "github.com/spf13/cobra"
	analysis "github.com/spf13/nitro"
	flag "github.com/spf13/pflag"
	globs "github.com/spf13/viper"
)

// dvlnCmd is dvln's root command. Every other command attached to dvlnCmd
// is a child or "subcommand" to it, currently dvln is only one level deep
// as far as the meta-cmd sub-cmd structure.
var dvlnCmd = &cli.Command{
	Use:   "dvln",
	Short: "dvln package/workspace mgmt tool",
	Long: `dvln: Multi-package development line and workspace management tool
  - written by Erik and friends in Go

For complete documentation see: http://dvln.org`,
	Run: func(cmd *cli.Command, args []string) {
		run(cmd, args)
	},
}

// Timer used by analysis code via the 'analysis' (nitro) pkg
var Timer *analysis.B

// cliPkgOut is an io.Writer used by the 'cli' (cobra) package to capture any
// output from it so the dvln tool can dump it via the 'out' output mgmt Go pkg
// to get the levels of data sent to the screen and optional log files.
var cliPkgOut = new(bytes.Buffer)

// tmpLogfileUsed indicates if we're using a tmp logfile to capture/mirror
// the screen output, if so we'll want to always dump the path to that file
// so the client knows where to find the file.
var tmpLogfileUsed = false

// init() preps the analysis pkg, scans in app globals, adds in subcommands
// and makes a 1st pass at prepping the CLI options/descriptions/defaults
// for the 'cli' (cobra) Go pkg being used to drive this CLI tool.
func init() {
	// Init the analysis package in case we turn analysis on
	Timer = analysis.Initalize()

	// Set up "global" key/value (variable) defaults in the 'globs' (viper) pkg,
	initPkgGlobs()

	// Add in the subcommands for the dvln command (get, update, ..), this
	// will allow all CLI opts to be traversed fully in the initial loading
	// of the CLI arguments into the 'globs' (viper) Go pkg
	addSubCommands()

	// Do an early pass on the CLI options, defaults may shift so this
	// function will be called again during runtime
	reloadCLIFlags := false
	setupDvlnCmdCLIArgs(reloadCLIFlags)

	// Feature: for any user defined options from hooks/plugins consider how to
	// let the 'cli' package know about them, likely via a pre-pass before one
	// of the setupDvlnCmdCLIArgs (which is called here and once more to attempt
	// to reset default values to correspond to a users config file settings/CLI)
}

// addSubCommands adds sub-commands to the top level dvln meta-command (dvlnCmd),
// note that dvlnCmd has been bootstrapped via the init() method already.
func addSubCommands() {
	//dvlnCmd.AddCommand(accessCmd) //    % dvln access ..
	//dvlnCmd.AddCommand(addCmd) //       % dvln add ..
	//dvlnCmd.AddCommand(blameCmd) //     % dvln blame ..
	//dvlnCmd.AddCommand(branchCmd) //    % dvln branch ..
	//dvlnCmd.AddCommand(catCmd) //       % dvln cat ..
	//dvlnCmd.AddCommand(checkCmd) //     % dvln check ..
	//dvlnCmd.AddCommand(commitCmd) //    % dvln commit ..
	//dvlnCmd.AddCommand(configCmd) //    % dvln config ..
	//dvlnCmd.AddCommand(copyrightCmd) // % dvln copyright ..
	//dvlnCmd.AddCommand(createCmd) //    % dvln create ..
	//dvlnCmd.AddCommand(dependCmd) //    % dvln depend ..
	//dvlnCmd.AddCommand(describeCmd) //  % dvln describe ..
	//dvlnCmd.AddCommand(diffCmd) //      % dvln diff ..
	//dvlnCmd.AddCommand(freezeCmd) //    % dvln freeze ..
	dvlnCmd.AddCommand(getCmd) //         % dvln get ..
	//dvlnCmd.AddCommand(issueCmd) //     % dvln issue ..
	//dvlnCmd.AddCommand(logCmd) //       % dvln log ..
	//dvlnCmd.AddCommand(manCmd) //       % dvln man ..
	//dvlnCmd.AddCommand(mergeCmd) //     % dvln merge ..
	//dvlnCmd.AddCommand(mirrorCmd) //    % dvln mirror ..
	//dvlnCmd.AddCommand(mvCmd) //        % dvln mv ..
	//dvlnCmd.AddCommand(patchCmd) //     % dvln patch ..
	//dvlnCmd.AddCommand(pushCmd) //      % dvln push ..
	dvlnCmd.AddCommand(pullCmd) //        % dvln pull ..
	//dvlnCmd.AddCommand(releaseCmd) //   % dvln release ..
	//dvlnCmd.AddCommand(retireCmd) //    % dvln retire ..
	//dvlnCmd.AddCommand(revertCmd) //    % dvln revert ..
	//dvlnCmd.AddCommand(rmCmd) //        % dvln rm ..
	//dvlnCmd.AddCommand(snapshotCmd) //  % dvln snapshot ..
	//dvlnCmd.AddCommand(statusCmd) //    % dvln status ..
	//dvlnCmd.AddCommand(tagCmd) //       % dvln tag ..
	//dvlnCmd.AddCommand(thawCmd) //      % dvln thaw ..
	//dvlnCmd.AddCommand(trackCmd) //     % dvln track ..
	dvlnCmd.AddCommand(versionCmd) //     % dvln version ..
}

// setupDvlnCmdCLIArgs sets up the CLI args available to the top level 'dvln'
// tool by telling the 'cli' (cobra) pkg what CLI opts the user can use.
// This is used by init() to bootstrap the data and re-used later to further
// update default value settings based on user config file settings and such.
// Note: there are "like" funcs, eg: cmds/get.go setupGetCLIArgs for 'dvln get'
func setupDvlnCmdCLIArgs(reloadCLIFlags bool) {
	var desc string

	if reloadCLIFlags {
		// this function is called multiple times, any 2nd (or 3rd) calls
		// must set reloadCLI flags otherwise it will panic within the 'cli'
		// (cobra) pkg (in the pflags pkg it uses).  The net effect of a reload
		// is that defaults for existing options will be updated, new options
		// can be added but that is, lets us say, less tested at this point.
		// - the primary use is to reload defaults so users config file settings
		//   are properly reflected in help screen/usage output and such
		dvlnCmd.Flags().SetDefValueReparseOK(true)
		dvlnCmd.PersistentFlags().SetDefValueReparseOK(true)
	}

	// Basic alphabetical listing of persistent flags (top and sub-level cmds),
	// note that if you look in dvln/cmd/globals.go in initPkgGlobs() it should
	// be pretty clear which options need to have CLI set up here, within the
	// local only block below or possibly within a given subcommands file
	// such as dvln/cmd/get.go for the 'dvln get' subcommand:
	desc, _, _ = globs.Desc("analysis")
	dvlnCmd.PersistentFlags().BoolVarP(&analysis.AnalysisOn, "analysis", "A", globs.GetBool("analysis"), desc)
	desc, _, _ = globs.Desc("config")
	dvlnCmd.PersistentFlags().StringP("config", "C", globs.GetString("config"), desc)
	desc, _, _ = globs.Desc("debug")
	dvlnCmd.PersistentFlags().BoolP("debug", "D", globs.GetBool("debug"), desc)
	desc, _, _ = globs.Desc("force")
	dvlnCmd.PersistentFlags().BoolP("force", "f", globs.GetBool("force"), desc)
	desc, _, _ = globs.Desc("fatalon")
	dvlnCmd.PersistentFlags().IntP("fatalon", "F", globs.GetInt("fatalon"), desc)
	desc, _, _ = globs.Desc("globs")
	dvlnCmd.PersistentFlags().StringP("globs", "G", globs.GetString("globs"), desc)
	desc, _, _ = globs.Desc("interact")
	dvlnCmd.PersistentFlags().BoolP("interact", "i", globs.GetBool("interact"), desc)
	desc, _, _ = globs.Desc("jobs")
	dvlnCmd.PersistentFlags().StringP("jobs", "J", globs.GetString("Jobs"), desc)
	desc, _, _ = globs.Desc("look")
	dvlnCmd.PersistentFlags().StringP("look", "L", globs.GetString("Look"), desc)
	desc, _, _ = globs.Desc("quiet")
	dvlnCmd.PersistentFlags().BoolP("quiet", "q", globs.GetBool("quiet"), desc)
	desc, _, _ = globs.Desc("record")
	dvlnCmd.PersistentFlags().StringP("record", "R", globs.GetString("record"), desc)
	desc, _, _ = globs.Desc("terse")
	dvlnCmd.PersistentFlags().BoolP("terse", "t", globs.GetBool("terse"), desc)
	desc, _, _ = globs.Desc("verbose")
	dvlnCmd.PersistentFlags().BoolP("verbose", "v", globs.GetBool("verbose"), desc)

	// NewCLIOpts: if there were opts for the dvln meta-command only they would
	// be added below, for persistent ops visible across all subcommands add
	// them above.  Put them in alphabetically ordered on the long opt name.
	// Note that you'll need to modify cmds/global.go as well otherwise your
	// globs.Desc() call and globs.GetBool("myopt") will not work.

	// The below opts apply *only* to the 'dvln' command itself, not subcommands
	desc, _, _ = globs.Desc("port")
	dvlnCmd.Flags().IntP("port", "P", globs.GetInt("Port"), desc)
	desc, _, _ = globs.Desc("serve")
	dvlnCmd.Flags().BoolP("serve", "S", globs.GetBool("serve"), desc)
	desc, _, _ = globs.Desc("version")
	dvlnCmd.Flags().BoolP("version", "V", globs.GetBool("version"), desc)

	dvlnCmd.Run = run
	if reloadCLIFlags {
		dvlnCmd.Flags().SetDefValueReparseOK(false)
		dvlnCmd.PersistentFlags().SetDefValueReparseOK(false)
	}
}

// Execute is called by main(), it basically finishes prepping the 'dvln'
// configuration data (combined with init() setting up options and available
// subcommands and such) and then kicks off the 'cli' (cobra) package to run
// subcommands and such via the dvlnCmd.Execute() call in the routine.
func Execute() {
	Timer.Step("cmds.Execute(): init() complete (defaults set, subcmds added, CLI args set up)")

	// Shove the CLI args into the 'globs' (viper) package before we even kick
	// into the 'cli' package Execute() call below, allows us to turn on debug
	// early as well as adjust the help screen to reflect opts the user has set:
	prepCLIArgs()

	// Load up the users dvln config file (ie: ~/.dvlncfg/cfg.json|toml/yaml..).
	// This may alter settings/configuration further so we'll again make a pass
	// at setting up the 'out' package with any new settings:
	scanUserCfgFile()

	// Now that we've read in the CLI args and the users config file we have a
	// full picture of the settings that will be used... now we'll take a 2nd
	// pass through 'cli' (cobra) and the underlying 'pflags' package it uses to
	// make the defaults for the CLI options match what the user has configured
	// or used via CLI opts and config file settings for the current tool run
	// - debatable but I like it for now, --help now reflects users full config
	reloadCLIDefaults()

	// Full opt/config file setup is now set up, now wrap up any early prep of
	// the dvln tool before kicking off the 'cli' (cobra) libraries Execute()
	// method (ie: start up commands/subcommands and finish processing opts)...
	// so we can set up # of CPU's to use, handle easy requests the user gives
	// such as what version of the tool is running (-V|--version), show settings
	// available via env or config file (--globs|-G {cfg|env}), etc.
	dvlnFinalPrep()

	//dvlnCmd.DebugFlags() // can be useful for debugging purposes now and then

	// Capture 'cli' (cobra) pkg output into the cliPkgOut byte buffer, note
	// that this only affects the 'cli' (cobra) packages output (which also,
	// btw, indirectly controls and affects the 'pflags' package used by it).
	// The reason we do this is so we can control all output via the 'out' pkg
	// so we'll grab any results from 'dvlnCmd.Execute()' and dump it below:
	dvlnCmd.SetOutput(cliPkgOut)

	Timer.Step("cmds.Execute(): loaded dvln user config, early setup and output prep done")

	// Allow partial command matching, shortest unique match
	cli.EnablePrefixMatching = true

	// Kick off 'cli' (cobra) pkg, will parse args and the cmd/subcmd tree
	// structure and, if no help output requested or error encountered, it will
	// then kick into requested cmd PersistentPreRun, PreRun, Run, PostRun,
	// and any PersistentPostRun functions... also added a PersistentHelpRun
	// and PersistentErrRun set of funcs so if help or errors we can still
	// deal with CLI opts for debugging and verbosity and such (and recording)
	err := dvlnCmd.Execute()
	Timer.Step("cmds.Execute(): dvlnCmd.Execute() complete, post ops next")
	anyOutput := cast.ToString(cliPkgOut)
	if err != nil {
		// Identify the issue..
		out.Issue(anyOutput)
		if tmpLogfileUsed {
			out.Noteln("Temp output logfile:", globs.GetString("record"))
		}
		out.Exit(-1)
	}
	// If any output remains from the cli (cobra) pkg dump it here (eg: usage)
	if anyOutput != "" {
		out.Print(anyOutput)
	}
	if tmpLogfileUsed {
		out.Noteln("Temp output logfile:", globs.GetString("record"))
	}
	Timer.Step("cmds.Execute(): complete")
}

// scanUserConfigFile initializes a viper/globs config file with sensible default
// configuration flags and sets up any activities that have been requested
// via the CLI and config settings (recording, debugging, verbosity, etc)
func scanUserConfigFile() {
	// What config file?, default: ~/.dvlncfg/cfg.{json|toml|yaml|yml}
	// the globs package uses config.json|toml|.. and we prefer less typing
	// so we're going with cfg.json|toml|<ext> as the default name
	globs.SetConfigName("cfg")

	// Now grab the config file path info from the 'globs' (viper) Go pkg which
	// has our globals and CLI opts and overrides set (except for the config
	// file as we haven't read it yet of course, that's what we're doing):
	configFile := globs.GetString("config")

	// Handle $HOME and ~ and such in the config file name
	configFullPath := globs.AbsPathify(configFile)

	// Typically Config defaults to a path (dir) to look for config.<extension>
	// files in but it can also be a full path to a file, try and detect which:
	if fileInfo, err := os.Stat(configFullPath); err == nil && fileInfo.IsDir() {
		// if it's a dir then just add the path, default looks for config.<etc>
		globs.AddConfigPath(configFile)
	} else {
		// if it's not a visible dir assume it's a file, if no file no problem
		globs.SetConfigFile(configFullPath)
	}
	globs.ReadInConfig()
}

// currentCmd is a package globs that will be 'dvln' if no subcommand was
// used, else it will be the subcommand, so if 'dvln get ..' then it'll be get'
var currentCmd string

// pushCLIOptsToGlobs is a bit of a hack, basically it "hacks" the 'cli' (cobra)
// package and the 'flag' (pflags) package under it to be able to pre-scan and
// parse all CLI args.  For the dvlnCmd meta-cmd and any subcmd used it will
// do a 'cli' (cobra) package Find() and ParseFlags() on them in a special
// "ignore bad flags" mode.  The idea is that if the user turns on debugging
// and perhaps asks to record output to a tmp log file, even if given with
// other invalid options, we want to accept those good options and shove them
// into the 'globs' (viper) package so we can kick on debugging and such as
// early as possible.  Note that we do choose to catch some 'cli' pkg errors
// here not related to flags (eg: a bad subcommand name used on the CLI).
func pushCLIOptsToGlobs(c *cli.Command, topCmd bool) {
	var args []string
	args = os.Args[1:]
	currErrHndl := c.Flags().ErrorHandling()

	// Tell the pflags package (used by 'cli') to ignore bad flags for this
	// early pass of flag parsing, the dvlnCmd.Execute() call will catch those
	c.Flags().SetErrorHandling(flag.IgnoreError)

	// Parse the CLI args into likely subcmd, flags given and any errors found:
	cmd, flags, err := c.Find(args)
	if err != nil && topCmd {
		// If this is the 1st pass on the top level dvlnCmd object (not the
		// subcommand getCmd or versionCmd objects) and if we are ignore flag
		// errors (as above) then any error coming back from Find will be from
		// non-flag issues (eg: bad subcommand name), will fail here if so:
		out.Issuef("Unable to parse command line: %s\n", err)
		out.IssueExitf(-1, "Please run 'dvln help' for usage\n")
	}
	// For nice errors lets stash either 'dvln' or, if a subcommand was used,
	// into the 'currentCmd' unexported package global so we know what the user
	// is running and can work (and error) with respect to that
	if currentCmd == "" {
		currentCmd = cmd.Name()
	}
	c.ParseFlags(flags)
	// Scan all flags to see what was used on CLI, shove ONLY used flags into
	// the 'globs' (viper) pkg so it's pflags and overide config levels focus
	// just on those CLI options actually used (I prefer that personally)
	globs.SetPFlags(c.Flags())
	c.Flags().SetErrorHandling(currErrHndl)
	// if running 'dvln <subcmd> ..' we'll also scan the <subcmd> args here
	// recursively, but if just 'dvln ..' with no subcmd then no need
	if c.HasSubCommands() && cmd.Name() != c.Root().Name() {
		topCmd = false // this is a subcmd, not the top 'dvln' cmd any longer
		pushCLIOptsToGlobs(cmd, topCmd)
	}
}

// adjustOutLevels examines verbosity related options and sets up the 'out'
// output control package to dump what the client has requested, as well as
// record any output to a logfile and such.
func adjustOutLevels() {
	// Set screen output threshold (defaults to LevelInfo which is the 'out'
	// pkg default already, but someone can change the level now via cfg/env)
	out.SetThreshold(out.LevelString2Level(globs.GetString("screenlevel")), out.ForScreen)

	// Note: for all of the below threshold settings the use of ForBoth means
	//       both screen and logfile output will be set at the given 'out' pkg
	//       levels, keep in mind that log file defaults to the writer
	//       ioutil.Discard (/dev/null) to start so you need to set up a writer
	//       which is done further below
	if globs.GetBool("debug") && globs.GetBool("verbose") {
		out.SetThreshold(out.LevelTrace, out.ForBoth)
		// For trace level (highest debug level) output we go crazy and turn
		// on many "prefix" flags (often used when writing to the logfile) so
		// that loglevels, timestamps, Go filename/line#/funcname, etc are all
		// displayed, set DVLN_SCREEN_FLAGS to "none" to suppress that
		if os.Getenv("DVLN_SCREEN_FLAGS") == "" {
			os.Setenv("DVLN_SCREEN_FLAGS", "debug")
		}
	} else if globs.GetBool("debug") {
		out.SetThreshold(out.LevelDebug, out.ForBoth)
	} else if globs.GetBool("verbose") {
		out.SetThreshold(out.LevelVerbose, out.ForBoth)
	} else if globs.GetBool("quiet") {
		out.SetThreshold(out.LevelError, out.ForScreen)
	}

	// Handle a few special case settings: pkg 'out' is low level and allows
	// for some env's and some API's to tweak it's settings (output line indent
	// and metadata augmentation), so we'll handle both the API's and the
	// env settings so that appropriate 'dvln' top level cmd settings get
	// pushed down into the 'out' package correctly.
	// Note: normally you would *NOT* want to do a hack like this, instead you
	// want to use 'globs' (viper) to store your key/values and, using that, you
	// get free env overrides and such (but the 'out' pkg is low level enough
	// that it can't depend upon the 'globs' config pkg (as it depends on 'out')
	// - note that we allow a setting of "none" to be special and to mean "",
	//   (see above DVLN_SCREEN_FLAG setting, maybe you don't want screen flags
	//   in which case using "none" will do that but "" would not)
	// Note: lean towards the above for future 'out' package tweaks
	var flags string
	if flags = os.Getenv("DVLN_DEBUG_SCOPE"); flags != "" {
		if flags != "none" {
			os.Setenv("PKG_OUT_DEBUG_SCOPE", flags)
		}
	}
	if flags = os.Getenv("DVLN_LOGFILE_FLAGS"); flags != "" {
		if flags != "none" {
			os.Setenv("PKG_OUT_LOGFILE_FLAGS", flags)
		}
	}
	if flags = os.Getenv("DVLN_NONZERO_EXIT_STRACKTRACE"); flags != "" {
		if flags != "none" {
			os.Setenv("PKG_OUT_NONZERO_EXIT_STACKTRACE", flags)
		}
	}
	if flags = os.Getenv("DVLN_PKG_OUT_SMART_FLAGS_PREFIX"); flags != "" {
		if flags != "none" {
			os.Setenv("PKG_OUT_SMART_FLAGS_PREFIX", flags)
		}
	}
	if flags = os.Getenv("DVLN_SCREEN_FLAGS"); flags != "" {
		if flags != "none" {
			os.Setenv("PKG_OUT_SCREEN_FLAGS", flags)
		}
	}

	jsonLevel := globs.GetInt("jsonindentlevel")
	api.SetJSONIndentLevel(jsonLevel)
	raw := globs.GetBool("jsonraw")
	api.SetJSONRaw(raw)
	jsonPrefix := globs.GetString("jsonprefix")
	api.SetJSONPrefix(jsonPrefix)

	// Like the 'out' package init above the 'pretty' package has no
	// dependencies on 'globs' (viper) but the reverse is true... so we
	// need to tell the 'pretty' package how we want our formatting done
	// (note that this honors DVLN_TEXTHUMANIZE, etc)
	humanize := globs.GetBool("texthumanize")
	pretty.SetHumanize(humanize)
	textLevel := globs.GetInt("textindentlevel")
	pretty.SetOutputIndentLevel(textLevel)
	textPrefixStr := globs.GetString("textprefix")
	pretty.SetOutputPrefixStr(textPrefixStr)

	// Lets handle recording of output..
	if record := globs.GetString("record"); record != "" && record != "off" {
		// If the user has requested recording/logging the below will set up
		// an io.Writer for a log file (via the 'out' package).  At this point
		// logging is enabled at the "Info/Print" (LevelInfo) level which
		// matches the default screen output setting
		//		out.SetThreshold(out.LevelInfo, out.ForLogfile)
		if record == "temp" || record == "tmp" {
			tmpLogfileUsed = true
			record = out.UseTempLogFile("dvln.")
			globs.Set("Record", record)
		} else {
			origRecord := record
			out.SetLogFile(globs.AbsPathify(record))
			// quick little hack to trim out home dir and shove in ~, keeps
			// the usage output brief if --help is used and such
			homeDir := globs.UserHomeDir()
			if homeDir != "" && strings.HasPrefix(record, homeDir+string(filepath.Separator)) {
				length := len(homeDir)
				rest := record[length:]
				record = "~" + cast.ToString(rest)
			}
			if origRecord != record {
				globs.Set("Record", record)
			}
		}
		currThresh := out.Threshold(out.ForLogfile)
		if currThresh == out.LevelDiscard {
			// if no threshold level set yet start with LevelInfo
			out.SetThreshold(out.LevelString2Level(globs.GetString("logfilelevel")), out.ForLogfile)
		}
	}
}

// prepCLIArgs scans all CLI opts and tries to shove them into 'globs' (viper)
// so we can then make a pass at turning on debugging, recording, etc as early
// as possible (now) if such options are used
func prepCLIArgs() {
	// Recursively traverse dvlnCmd and all subcommands and do an early
	// "ignore errors" pass at parsing the CLI flags and shoving any valid
	// flags into the "globs" (viper) package.  What could go wrong?  ;)
	if len(os.Args) != 1 {
		topCmd := true // passing in the top level cmd obj at this point, yes
		pushCLIOptsToGlobs(dvlnCmd, topCmd)
	}

	// Do an early output level adjustment in case CLI opts are used that will
	// add debug/record/etc info so that our adjustOutLevels() actually has a
	// chance to dump any debug/trace/etc level output calls "early", final
	// adjustments will be done with another call down below.  Early setup:
	adjustOutLevels()
}

// scanUserCfgFile finishes updating the 'globs' (viper) pkg so that all
// CLI opts are fully visible and the users config file data is also loaded
// up as well, hurray!  With that data we'll also re-update dvln so that
// output data is going to the screen and any mirror logfile at the right
// output levels and such (and that any help screens reflect those final
// "full" settings from all this config data)
func scanUserCfgFile() {
	// Scan the users config file, if any, honoring any CLI opts that might
	// override the location of the user config file and push those settings
	// into the 'globs' (viper) pkg as well.  Once done the 'globs' globals will
	// be complete (Feature: except for future codebase and VCS pkg settings):
	scanUserConfigFile()

	// Final output levels adjustements to take into account any tweaks from
	// the users config file settings.  Note, don't move this below the calls
	// to the "setup*CmdCLIArgs()" routines, we need it to make final tweaks
	// to things like the --record flag before we do the final option default
	// reload.
	adjustOutLevels()
}

// reloadCLIDefaults finishes updating the 'globs' (viper) pkg so that all
// CLI opts are fully visible and the users config file data is also reflected
// in the default settings for each arg, this happens after the users config
// file is scanned so any settings there are reflect in options help accurately
func reloadCLIDefaults() {
	// init() in dvln.go & subcmd files (eg:cmds/get.go) all do a 1st pass
	// loading in options and defaults for the entire cmd/subcommand structure.
	// - each subcommand init's it's own globs via setup<subcmd>CmdCLIArgs()
	// Now do a 2nd pass on the CLI options, this one will take into account the
	// config file we just read and update the defaults for each option so the
	// 'cli' (cobra) pkg help screen is now accurate based on the users config
	// file settings and even CLI flags used:
	reloadCLIFlags := true
	setupDvlnCmdCLIArgs(reloadCLIFlags)
	setupGetCmdCLIArgs(reloadCLIFlags)
	setupPullCmdCLIArgs(reloadCLIFlags)
	setupVersionCmdCLIArgs(reloadCLIFlags)
	// NewSubCommand: If you add a new subcommand you need to add a method to
	//     that subcommand named like what's above, see cmds/get.go for the
	//     setupGetCmdCLIArgs() function to get an idea (so if you add the
	//     'diff' subcommand create diff.go and add setupDiffCmdCLIArgs() in
	//     cmds/diff.go and call it from within init() in diff.go and also
	//     call it from directly above).
}

// dvlnFinalPrep basically does just that... now that the 'globs' config
// data is fully populated with CLI's, env's, config files, codebase/pkg
// settings and defaults, handle any "easy" opts we can, eg: show version (-V),
// show available "global" cfg/env settings (-G), set up the number of parallel
// CPU's to leverage (-j<#>), etc... all stuff that can happen before we kick
// into the full 'cli' (cobra) commander package 'Execute()' method.
func dvlnFinalPrep() {
	// (re)Dump user config file info.  Possibly dumped already from the calls
	// within scanUserConfigFile() but, if output/logfile thresholds changed in
	// the users config file we may have missed logging it, so dump it again as
	// it's useful for client/admin troubleshooting of dvln:
	if globs.ConfigFileUsed() != "" {
		out.Debugln("Used config file:", globs.ConfigFileUsed())
	}
	cmdName := " [subcmd]"
	if currentCmd != "" {
		if currentCmd == dvlnCmd.Root().Name() {
			cmdName = ""
		} else {
			cmdName = " " + currentCmd
		}
	}

	// Honor the parallel jobs setting (-j, --jobs, cfg file setting Jobs or env
	// var DVLN_JOBS can all control this), identifies # of CPU's to use.
	numCPU := runtime.NumCPU()
	if jobs := globs.GetString("jobs"); jobs != "" && jobs != "all" {
		if _, err := strconv.Atoi(jobs); err != nil {
			out.Issuef("Jobs value should be a number or 'all', found: %s\n", jobs)
			out.IssueExitf(-1, "Please run 'dvln help%s' for usage\n", cmdName)
		}
		numJobs := cast.ToInt(jobs)
		if numJobs > numCPU {
			numJobs = numCPU
		}
		runtime.GOMAXPROCS(numJobs)
	} else {
		runtime.GOMAXPROCS(numCPU)
	}

	// Do some validation on the 'serve' mode:
	if serve := globs.GetBool("serve"); serve {
		out.Fatalln("Serve mode is not available yet, to contribute email 'brady@dvln.org'")
	}

	// Make sure that given --look|-l or cfgfile:Look or env:DVLN_LOOK are valid
	look := globs.GetString("look")
	if look != "text" && look != "json" {
		out.IssueExitf(-1, "The --look option (-l) can only be set to 'text' or 'json', found: '%s'\n", look)
	} else if look == "json" && globs.GetBool("interact") {
		out.Debugln("Interactive runs are not available for the 'json' output \"look\"")
		out.Debugln("- silently disabling interaction (client may have it set for text output)")
		globs.Set("interact", false)
	}

	// If the developer asks for the version of the tool print that out:
	if version := globs.GetBool("version"); version {
		out.Print(lib.DvlnVerStr())
		os.Exit(0)
	}

	// If trace level debug enabled (checked inside the routine) this will dump
	// the "globs" (viper) config via the 'out.Trace*()' calls run within the
	// given method:
	globs.Debug()

	globsCLI := globs.GetString("globs")
	if globsCLI != "" && globsCLI != "env" && globsCLI != "cfg" {
		out.Issuef("The --globs option (-G) can only be set to 'env' or 'cfg', found: '%s'\n", globsCLI)
		out.IssueExitf(-1, "Please run 'dvln help%s' for usage\n", cmdName)
	}
	// If the client asks for user available "globs" settable via env or cfgfile
	if globsCLI == "env" || globsCLI == "cfg" {
		out.Print(fmt.Sprintf("%v", globs.GetSingleton()))
		os.Exit(0)
	}
}

// run for the dvln cmd really doesn't do anything but recommend the user
// select a subcommand to run
func run(cmd *cli.Command, args []string) {
	out.IssueExitln(-1, "Please use a valid subcommand (for a list: 'dvln help')")
}
