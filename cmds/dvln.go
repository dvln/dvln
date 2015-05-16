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
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/dvln/lib"
	"github.com/dvln/out"
	"github.com/spf13/cast"
	cli "github.com/spf13/cobra"
	analysis "github.com/spf13/nitro"
	globs "github.com/spf13/viper"
)

// DvlnCmd is dvln's root command. Every other command attached to DvlnCmd
// is a child or "subcommand" to it, currently dvln is only one level deep
// as far as the meta-cmd sub-cmd structure.
var DvlnCmd = &cli.Command{
	Use:   "dvln",
	Short: "dvln package/workspace mgmt tool",
	Long: `dvln: Multi-package development line and workspace management tool
  - written by Erik and friends in Go

For complete documentation see: http://dvln.org`,
	Run: func(cmd *cli.Command, args []string) {
		run(cmd, args)
	},
}

// package local CLI flags for dvln/subcommands, needed until 'cli' (cobra) can
// push these automatically into 'globs' (viper), ugh.  See cmd/globals.go for
// descriptions of any of these items (just remove the 'cli' prefix and look
// for a case insensitive match on therest of the names):
var cliConfig, cliCodeBase, cliDevLine, cliJobs, cliLook, cliPkg, cliRecord, cliWkspcDir string
var cliDebug, cliForce, cliInteract, cliQuiet, cliServe, cliTerse, cliVerbose, cliVersion bool
var cliFatalOn, cliPort int

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

// insure that the globs (viper) packages env prefix is set *before* any
// init() functions are run so they all honor hte DVLN_ prefix
var envPrefix = globs.SetEnvPrefix("DVLN")

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
	releadCLIFlags := false
	setupDvlnCmdCLIArgs(releadCLIFlags)

	// Feature: for any user defined options from hooks/plugins consider how to
	// let the 'cli' package know about them, likely via a pre-pass before one
	// of the setupDvlnCmdCLIArgs (which is called here and once more to attempt
	// to reset default values to correspond to a users config file settings/CLI)
}

// addSubCommands adds sub-commands to the top level dvln meta-command (DvlnCmd),
// note that DvlnCmd has been bootstrapped via the init() method already.
func addSubCommands() {
	//DvlnCmd.AddCommand(accessCmd) //    % dvln access ..
	//DvlnCmd.AddCommand(addCmd) //       % dvln add ..
	//DvlnCmd.AddCommand(blameCmd) //     % dvln blame ..
	//DvlnCmd.AddCommand(branchCmd) //    % dvln branch ..
	//DvlnCmd.AddCommand(catCmd) //       % dvln cat ..
	//DvlnCmd.AddCommand(checkCmd) //     % dvln check ..
	//DvlnCmd.AddCommand(commitCmd) //    % dvln commit ..
	//DvlnCmd.AddCommand(configCmd) //    % dvln config ..
	//DvlnCmd.AddCommand(copyrightCmd) // % dvln copyright ..
	//DvlnCmd.AddCommand(createCmd) //    % dvln create ..
	//DvlnCmd.AddCommand(dependCmd) //    % dvln depend ..
	//DvlnCmd.AddCommand(describeCmd) //  % dvln describe ..
	//DvlnCmd.AddCommand(diffCmd) //      % dvln diff ..
	//DvlnCmd.AddCommand(freezeCmd) //    % dvln freeze ..
	DvlnCmd.AddCommand(getCmd) //         % dvln get ..
	//DvlnCmd.AddCommand(issueCmd) //     % dvln issue ..
	//DvlnCmd.AddCommand(logCmd) //       % dvln log ..
	//DvlnCmd.AddCommand(manCmd) //       % dvln man ..
	//DvlnCmd.AddCommand(mergeCmd) //     % dvln merge ..
	//DvlnCmd.AddCommand(mirrorCmd) //    % dvln mirror ..
	//DvlnCmd.AddCommand(mvCmd) //        % dvln mv ..
	//DvlnCmd.AddCommand(patchCmd) //     % dvln patch ..
	//DvlnCmd.AddCommand(pushCmd) //      % dvln push ..
	//DvlnCmd.AddCommand(pullCmd) //      % dvln pull ..
	//DvlnCmd.AddCommand(releaseCmd) //   % dvln release ..
	//DvlnCmd.AddCommand(retireCmd) //    % dvln retire ..
	//DvlnCmd.AddCommand(revertCmd) //    % dvln revert ..
	//DvlnCmd.AddCommand(rmCmd) //        % dvln rm ..
	//DvlnCmd.AddCommand(snapshotCmd) //  % dvln snapshot ..
	//DvlnCmd.AddCommand(statusCmd) //    % dvln status ..
	//DvlnCmd.AddCommand(tagCmd) //       % dvln tag ..
	//DvlnCmd.AddCommand(thawCmd) //      % dvln thaw ..
	//DvlnCmd.AddCommand(trackCmd) //     % dvln track ..
	DvlnCmd.AddCommand(versionCmd) //     % dvln version ..
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
		DvlnCmd.Flags().SetDefValueReparseOK(true)
		DvlnCmd.PersistentFlags().SetDefValueReparseOK(true)
	}

	// Basic alphabetical listing of persistent flags (top and sub-level cmds),
	// note that if you look in dvln/cmd/globals.go in initPkgGlobs() it should
	// be pretty clear which options need to have CLI set up here, within the
	// local only block below or possibly within a given subcommands file
	// such as dvln/cmd/get.go for the 'dvln get' subcommand:
	desc, _, _ = globs.Desc("analysis")
	DvlnCmd.PersistentFlags().BoolVarP(&analysis.AnalysisOn, "analysis", "A", globs.GetBool("analysis"), desc)
	desc, _, _ = globs.Desc("config")
	DvlnCmd.PersistentFlags().StringVarP(&cliConfig, "config", "C", globs.GetString("config"), desc)
	desc, _, _ = globs.Desc("debug")
	DvlnCmd.PersistentFlags().BoolVarP(&cliDebug, "debug", "D", globs.GetBool("debug"), desc)
	desc, _, _ = globs.Desc("force")
	DvlnCmd.PersistentFlags().BoolVarP(&cliForce, "force", "f", globs.GetBool("force"), desc)
	desc, _, _ = globs.Desc("fatalon")
	DvlnCmd.PersistentFlags().IntVarP(&cliFatalOn, "fatalon", "F", globs.GetInt("fatalon"), desc)
	desc, _, _ = globs.Desc("interact")
	DvlnCmd.PersistentFlags().BoolVarP(&cliInteract, "interact", "i", globs.GetBool("interact"), desc)
	desc, _, _ = globs.Desc("jobs")
	DvlnCmd.PersistentFlags().StringVarP(&cliJobs, "jobs", "j", globs.GetString("Jobs"), desc)
	desc, _, _ = globs.Desc("look")
	DvlnCmd.PersistentFlags().StringVarP(&cliLook, "look", "L", globs.GetString("Look"), desc)
	desc, _, _ = globs.Desc("quiet")
	DvlnCmd.PersistentFlags().BoolVarP(&cliQuiet, "quiet", "q", globs.GetBool("quiet"), desc)
	desc, _, _ = globs.Desc("record")
	DvlnCmd.PersistentFlags().StringVarP(&cliRecord, "record", "R", globs.GetString("record"), desc)
	desc, _, _ = globs.Desc("terse")
	DvlnCmd.PersistentFlags().BoolVarP(&cliTerse, "terse", "t", globs.GetBool("terse"), desc)
	desc, _, _ = globs.Desc("verbose")
	DvlnCmd.PersistentFlags().BoolVarP(&cliVerbose, "verbose", "v", globs.GetBool("verbose"), desc)

	// As well as just top level dvln only command flags
	desc, _, _ = globs.Desc("port")
	DvlnCmd.Flags().IntVarP(&cliPort, "port", "P", globs.GetInt("Port"), desc)
	desc, _, _ = globs.Desc("serve")
	DvlnCmd.Flags().BoolVarP(&cliServe, "serve", "S", globs.GetBool("serve"), desc)
	desc, _, _ = globs.Desc("version")
	DvlnCmd.Flags().BoolVarP(&cliVersion, "version", "V", globs.GetBool("version"), desc)
	DvlnCmd.Run = run
	if reloadCLIFlags {
		DvlnCmd.Flags().SetDefValueReparseOK(false)
		DvlnCmd.PersistentFlags().SetDefValueReparseOK(false)
	}
}

// Execute is called by main(), it basically finishes prepping the 'dvln'
// configuration data (combined with init() setting up options and available
// subcommands and such) and then kicks off the 'cli' (cobra) package to run
// subcommands and such via the DvlnCmd.Execute() call in the routine.
func Execute() {
	Timer.Step("cmds.Execute(): init() complete (defaults set, subcmds added, CLI args set up)")
	// Load up the users dvln config file from the correct location
	loadDvlnConfigFile()

	// Now that the full config is loaded into the 'globs' (viper) pkg lets handle
	// any early setup for the 'dvln' tool needed to give user any info needed
	// or setup number of CPU's to use and that sort of thing.  These are things
	// that can be done before kicking off subcommands (see Execute() below)
	dvlnEarlySetup()

	// Capture cli (cobra) output into the cliPkgOut byte buffer, note that
	// this only affects the 'cli' (cobra) packages output (which indirectly
	// controls also the 'pflags' package used by it)
	DvlnCmd.SetOutput(cliPkgOut)

	Timer.Step("cmds.Execute(): loaded dvln user config, early setup and output prep done")

	// Allow partial command matching, shortest unique match
	cli.EnablePrefixMatching = true

	// Kick off 'cli' (cobra) pkg, will parse args and the cmd/subcmd tree
	// structure and, if no help output requested or error encountered, it will
	// then kick into requested cmd PersistentPreRun, PreRun, Run, PostRun,
	// and any PersistentPostRun functions... also added a PersistentHelpRun
	// and PersistentErrRun set of funcs so if help or errors we can still
	// deal with CLI opts for debugging and verbosity and such (and recording)
	err := DvlnCmd.Execute()
	Timer.Step("cmds.Execute(): DvlnCmd.Execute() complete, post ops next")
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
		out.Notef("Temp output logfile is \"%s\"\n", globs.GetString("record"))
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

// pushCLIOptsToGlobs is used to peruse the flags used by the client on the CLI
// (in the 'cli' pkg via init() methods) and to now update the 'globs' (viper)
// package so that those CLI settings are recorded correctly there.
// Feature: The various globs.Set() calls should eventually be auto-handled by
// the 'cli' (cobra) package but currently aren't (when that is done the 'cli*'
// variables at the top of this pkg should be tossed and the *Flags() methods
// used shouldn't need a global to shove the flag results into, they should
// just automatically go into the globs/viper package).  Do not remove the 1st
// part of this method though when that is done, it is needed.
func pushCLIOptsToGlobs() {
	if len(os.Args) == 1 {
		return
	}
	var args []string
	args = os.Args[1:]

	// Feature: as hooks are done we can grab the 1st field from Find
	//        and get the cmd/subcmd info so we can load up the correct hooks
	//        for the command that is running (?needs consideration?), note
	//        that, as in the FIXME note below, the 1st return value which
	//        is the command isn't coming back correctly in shortened subcmd
	//        name situations.

	// Find the flags the user used, traversing commands, subcommands for
	// all allowed flags and such and storing them in the 'pflags' package
	// Flagset structure used by the 'cli' (cobra) package:
	_, flags, _ := DvlnCmd.Find(args)

	// FIXME: this is likely kind of weak... I've seen 'dvln ver -Ljson' return
	//        an error here (3rd arg) about ver not being a known command but it
	//        is a unique short match for 'version' so that shouldn't be the case.
	//        Right now we ignore the error and just do a 1st pass parse always
	//        as that gets the persistent flags into the config regardless.
	DvlnCmd.ParseFlags(flags)

	// Feature: ParseFlags should auto-push the below CLI settings into the
	//        'globs' (viper) pkg so we shouldn't have to do that below with all
	//        the Changed() calls... but that isn't done now so we do it here:

	// NewSubCommand: If you add a new subcommand you need to add a method to
	//     that subcommand named like what's below, see cmds/get.go for the
	//     pushGetCmdCLIOptsToGlobs() to get an idea.
	pushDvlnCmdCLIOptsToGlobs()
	pushGetCmdCLIOptsToGlobs()
	pushVersionCmdCLIOptsToGlobs()
}

func pushDvlnCmdCLIOptsToGlobs() {
	// Persistent flags are pushed into the 'globs' (viper) settings package here
	if DvlnCmd.PersistentFlags().Lookup("analysis").Changed {
		globs.Set("analysis", analysis.AnalysisOn)
	}
	if DvlnCmd.PersistentFlags().Lookup("config").Changed {
		globs.Set("config", cliConfig)
	}
	if DvlnCmd.PersistentFlags().Lookup("debug").Changed {
		globs.Set("debug", cliDebug)
	}
	if DvlnCmd.PersistentFlags().Lookup("fatalon").Changed {
		globs.Set("fatalon", cliFatalOn)
	}
	if DvlnCmd.PersistentFlags().Lookup("force").Changed {
		globs.Set("force", cliForce)
	}
	if DvlnCmd.PersistentFlags().Lookup("interact").Changed {
		globs.Set("interact", cliInteract)
	}
	if DvlnCmd.PersistentFlags().Lookup("jobs").Changed {
		globs.Set("jobs", cliJobs)
	}
	if DvlnCmd.PersistentFlags().Lookup("look").Changed {
		globs.Set("look", cliLook)
	}
	if DvlnCmd.PersistentFlags().Lookup("quiet").Changed {
		globs.Set("quiet", cliQuiet)
	}
	if DvlnCmd.PersistentFlags().Lookup("record").Changed {
		globs.Set("record", cliRecord)
	}
	if DvlnCmd.PersistentFlags().Lookup("terse").Changed {
		globs.Set("terse", cliTerse)
	}
	if DvlnCmd.PersistentFlags().Lookup("verbose").Changed {
		globs.Set("verbose", cliVerbose)
	}

	// local flags for dvln bootstrapped here
	if DvlnCmd.Flags().Lookup("port").Changed {
		globs.Set("port", cliPort)
	}
	if DvlnCmd.Flags().Lookup("serve").Changed {
		globs.Set("serve", cliServe)
	}
	if DvlnCmd.Flags().Lookup("version").Changed {
		globs.Set("version", cliVersion)
	}
}

// adjustOutLevels examines verbosity related options and sets up the 'out'
// output control package to dump what the client has requested, as well as
// record any output to a logfile and such.
func adjustOutLevels() {
	// Set screen output threshold (defaults to LevelInfo which is the 'out'
	// pkg default already, but someone can change the level now via cfg/env)
	out.SetThreshold(out.Level(globs.GetInt("screenlevel")), out.ForScreen)
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
	// for an env to tweak it's flags (output line metadata augmentation) on
	// the fly... so we'll let special DVLN env settings control that.  Note
	// that normally you would *NOT* want to do a hack like this, instead you
	// want to use 'globs' (viper) to store your key/values and, using that, you
	// get free env overrides and such (but the 'out' pkg is low level enough
	// that it can't depend upon the 'globs' config pkg (as it depends on 'out')
	// - note that we allow a setting of "none" to be special and to mean "",
	//   (see above DVLN_SCREEN_FLAG setting, maybe you don't want screen flags
	//   in which case using "none" will do that but "" would not)
	var flags string
	if flags = os.Getenv("DVLN_SCREEN_FLAGS"); flags != "" {
		if flags != "none" {
			os.Setenv("PKG_OUT_SCREEN_FLAGS", flags)
		}
	}
	if flags = os.Getenv("DVLN_LOGFILE_FLAGS"); flags != "" {
		if flags != "none" {
			os.Setenv("PKG_OUT_LOGFILE_FLAGS", flags)
		}
	}
	if flags = os.Getenv("DVLN_DEBUG_SCOPE"); flags != "" {
		if flags != "none" {
			os.Setenv("PKG_OUT_DEBUG_SCOPE", flags)
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
			out.SetLogFile(globs.AbsPathify(record))
			// quick little hack to trim out home dir and shove in ~, keeps
			// the usage output brief if --help is used and such
			homeDir := globs.UserHomeDir()
			if homeDir != "" && strings.HasPrefix(record, homeDir+string(filepath.Separator)) {
				length := len(homeDir)
				rest := record[length:]
				record = "~" + cast.ToString(rest)
			}
			globs.Set("Record", record)
		}
		currThresh := out.Threshold(out.ForLogfile)
		if currThresh == out.LevelDiscard {
			// if no threshold level set yet start with LevelInfo
			out.SetThreshold(out.Level(globs.GetInt("logfilelevel")), out.ForLogfile)
		}
	}
}

// loadDvlnConfigFile finishes updating the 'globs' (viper) pkg so that all
// CLI opts are fully visible and the users config file data is also loaded
// up as well, hurray!  With that data we'll also re-update dvln so that
// output data is going to the screen and any mirror logfile at the right
// output levels and such (and that any help screens reflect those final
// "full" settings from all this config data)
func loadDvlnConfigFile() {
	// Push all CLI options into globs (viper) pkg at which point we've taken into
	// account default opt settings (already set in globs via the init method),
	// env settings (handled in globs get calls), and, with this, CLI options
	// have been parsed and pushed into 'globs'.  All that remains is the users
	// config file (below) and any codebase or VCS pkg settings.
	pushCLIOptsToGlobs()

	// Do an early output level adjustment in case CLI opts are used that will
	// add debug/record/etc info so that our adjustOutLevels() actually has a
	// chance to dump any debug/trace/etc level output calls "early", final
	// adjustments will be done with another call down below.  Early setup:
	adjustOutLevels()

	// Scan the users config file, if any, honoring any CLI opts that might
	// override the location of the user config file and push those settings
	// into the 'globs' (viper) pkg as well.  Once done the 'globs' globals will
	// be complete (Feature: except for fugure codebase and VCS pkg settings):
	scanUserConfigFile()

	// Final output levels adjustements to take into account any tweaks from
	// the users config file settings.  Note, don't move this below the calls
	// to the "setup*CmdCLIArgs()" routines, we need it to make final tweaks
	// to things like the --record flag before we do the final option default
	// reload.
	adjustOutLevels()

	// init() here and in the subcmd files (cmds/get.go) all do a 1st pass of
	// loading in options and defaults for the entire cmd/subcommand structure.
	// Now do a 2nd pass on the CLI options, this one will take into account the
	// config file we just read and update the defaults for each option so the
	// 'cli' (cobra) pkg help screen is now accurate based on the users config
	// file settings and even CLI flags used:
	// NewSubCommand: If you add a new subcommand you need to add a method to
	//     that subcommand named like what's below, see cmds/get.go for the
	//     setupGetCmdCLIArgs() function to get an idea.
	reloadCLIFlags := true
	setupDvlnCmdCLIArgs(reloadCLIFlags)
	setupGetCmdCLIArgs(reloadCLIFlags)
	setupVersionCmdCLIArgs(reloadCLIFlags)

}

// dvlnEarlySetup basically does just that... now that the 'globs' config
// data is fully populated with CLI's, env's, config files, codebase/pkg
// settings and defaults, lets use it for 'dvln' level early setup
func dvlnEarlySetup() {
	// (re)Dump user config file info.  Possibly dumped already from the calls
	// within scanUserConfigFile() but, if output/logfile thresholds changed in
	// the users config file we may have missed logging it, so do it again as
	// it's useful for client/admin troubleshooting of dvln:
	if globs.ConfigFileUsed() != "" {
		out.Debugln("Used config file:", globs.ConfigFileUsed())
	}

	// Honor the parallel jobs setting (-j, --jobs, cfg file setting Jobs or env
	// var DVLN_JOBS can all control this), identifies # of CPU's to use.
	numCPU := runtime.NumCPU()
	if jobs := globs.GetString("jobs"); jobs != "" && jobs != "all" {
		if _, err := strconv.Atoi(jobs); err != nil {
			out.Issuef("Jobs value should be a number or 'all', \"%s\" was given\n", jobs)
			out.IssueExitln(-1, "Please run 'dvln [subcmd] --help' for usage")
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

	// If the developer asks for the version of the tool print that out
	if version := globs.GetBool("version"); version {
		out.Println(lib.DvlnVerStr())
		os.Exit(0)
	}
}

// run for the dvln cmd really doesn't do anything but recommend the user
// select a subcommand to run
func run(cmd *cli.Command, args []string) {
	out.IssueExitln(-1, "Please use a valid subcommand (for a list: 'dvln help')")
}
