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
	"os/user"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/dvln/api"
	"github.com/dvln/cast"
	cli "github.com/dvln/cobra"
	"github.com/dvln/lib"
	analysis "github.com/dvln/nitro"
	"github.com/dvln/out"
	flag "github.com/dvln/pflag"
	"github.com/dvln/pretty"
	"github.com/dvln/util/homedir"
	"github.com/dvln/util/path"
	globs "github.com/dvln/viper"
	"github.com/dvln/wkspc"
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

// Used when temp logging is active, the boolean will be true if that is the
// case and the tmpLogfileMsg will be the message to print to the screen before
// exit occurs (that string will be empty if in JSON mode since we'll have
// already stored the tmp logfile name in a note in the JSON data)
var tmpLogfileMsg string
var tmpLogfileActive = false

// At startup time we'll initialize subcmds, opts, etc
func init() {
	doDvlnInit()
}

// doDvlnInit() preps the analysis pkg, scans in app globals, adds in subcommands
// and makes a 1st pass at prepping the CLI options/descriptions/defaults
// for the 'cli' (cobra) Go pkg being used to drive this CLI tool.
func doDvlnInit() {
	// Init the analysis package in case we turn analysis on
	Timer = analysis.Initalize()

	// Set up "global" key/value (variable) defaults in the 'globs' (viper) pkg,
	doPkgCmdGlobsInit()

	// Add in the subcommands for the dvln command (get, update, ..), this
	// will allow all CLI opts to be traversed fully in the initial loading
	// of the CLI arguments into the 'globs' (viper) Go pkg
	addSubCommands(dvlnCmd)

	// Do a first pass on the CLI options, defaults may shift so this
	// function will be called again during runtime
	reloadCLIFlags := false
	setupDvlnCmdCLIArgs(dvlnCmd, reloadCLIFlags)

	// Feature: for any user defined options from hooks/plugins consider how to
	// let the 'cli' package know about them, likely via a pre-pass before one
	// of the setupDvlnCmdCLIArgs (which is called here and once more to attempt
	// to reset default values to correspond to a users config file settings/CLI)
}

// addSubCommands adds sub-commands to the top level dvln meta-command (dvlnCmd),
// note that dvlnCmd has been bootstrapped via the init() method already.
func addSubCommands(c *cli.Command) {
	//c.AddCommand(accessCmd) //    % dvln access ..
	//c.AddCommand(addCmd) //       % dvln add ..
	//c.AddCommand(blameCmd) //     % dvln blame ..
	//c.AddCommand(branchCmd) //    % dvln branch ..
	//c.AddCommand(catCmd) //       % dvln cat ..
	//c.AddCommand(checkCmd) //     % dvln check ..
	//c.AddCommand(commitCmd) //    % dvln commit ..
	//c.AddCommand(configCmd) //    % dvln config ..
	//c.AddCommand(copyrightCmd) // % dvln copyright ..
	//c.AddCommand(createCmd) //    % dvln create ..
	//c.AddCommand(dependCmd) //    % dvln depend ..
	//c.AddCommand(describeCmd) //  % dvln describe ..
	//c.AddCommand(diffCmd) //      % dvln diff ..
	//c.AddCommand(freezeCmd) //    % dvln freeze ..
	c.AddCommand(getCmd) //         % dvln get ..
	//c.AddCommand(issueCmd) //     % dvln issue ..
	//c.AddCommand(logCmd) //       % dvln log ..  (or maybe dvln list?)
	//c.AddCommand(manCmd) //       % dvln man ..
	//c.AddCommand(mergeCmd) //     % dvln merge ..
	//c.AddCommand(mirrorCmd) //    % dvln mirror ..
	//c.AddCommand(mvCmd) //        % dvln mv ..
	//c.AddCommand(patchCmd) //     % dvln patch ..
	//c.AddCommand(pushCmd) //      % dvln push ..
	//c.AddCommand(pullCmd) //      % dvln pull ..
	//c.AddCommand(releaseCmd) //   % dvln release ..
	//c.AddCommand(retireCmd) //    % dvln retire ..
	//c.AddCommand(revertCmd) //    % dvln revert ..
	//c.AddCommand(rmCmd) //        % dvln rm ..
	//c.AddCommand(snapshotCmd) //  % dvln snapshot ..
	//c.AddCommand(statusCmd) //    % dvln status ..
	//c.AddCommand(tagCmd) //       % dvln tag ..
	//c.AddCommand(thawCmd) //      % dvln thaw ..
	//c.AddCommand(trackCmd) //     % dvln track ..
	c.AddCommand(updateCmd)  //     % dvln update ..
	c.AddCommand(versionCmd) //     % dvln version ..
}

// setupDvlnCmdCLIArgs sets up the CLI args available to the top level 'dvln'
// tool by telling the 'cli' (cobra) pkg what CLI opts the user can use.
// This is used by init() to bootstrap the data and re-used later to further
// update default value settings based on user config file settings and such.
// Note: there are "like" funcs, eg: cmds/get.go setupGetCLIArgs for 'dvln get'
func setupDvlnCmdCLIArgs(c *cli.Command, reloadCLIFlags bool) {
	var desc string
	if reloadCLIFlags {
		// this function is called multiple times, any 2nd (or 3rd) calls
		// must set reloadCLI flags otherwise it will panic within the 'cli'
		// (cobra) pkg (in the pflags pkg it uses).  The net effect of a reload
		// is that defaults for existing options will be updated, new options
		// can be added but that is, lets us say, less tested at this point.
		// - the primary use is to reload defaults so users config file settings
		//   are properly reflected in help screen/usage output and such
		c.Flags().SetDefValueReparseOK(true)
		c.PersistentFlags().SetDefValueReparseOK(true)
	}

	// Basic alphabetical listing of persistent flags (top and sub-level cmds),
	// note that if you look in dvln/cmd/globals.go in initPkgGlobs() it should
	// be pretty clear which options need to have CLI set up here, within the
	// local only block below or possibly within a given subcommands file
	// such as dvln/cmd/get.go for the 'dvln get' subcommand:
	desc, _, _ = globs.Desc("analysis")
	c.PersistentFlags().BoolVarP(&analysis.AnalysisOn, "analysis", "A", globs.GetBool("analysis"), desc)
	desc, _, _ = globs.Desc("config")
	c.PersistentFlags().StringP("config", "C", globs.GetString("config"), desc)
	desc, _, _ = globs.Desc("debug")
	c.PersistentFlags().BoolP("debug", "D", globs.GetBool("debug"), desc)
	desc, _, _ = globs.Desc("help")
	c.PersistentFlags().BoolP("help", "h", globs.GetBool("help"), desc)
	desc, _, _ = globs.Desc("force")
	c.PersistentFlags().BoolP("force", "f", globs.GetBool("force"), desc)
	desc, _, _ = globs.Desc("fatalon")
	c.PersistentFlags().IntP("fatalon", "F", globs.GetInt("fatalon"), desc)
	desc, _, _ = globs.Desc("globs")
	c.PersistentFlags().StringP("globs", "G", globs.GetString("globs"), desc)
	desc, _, _ = globs.Desc("interact")
	c.PersistentFlags().BoolP("interact", "i", globs.GetBool("interact"), desc)
	desc, _, _ = globs.Desc("jobs")
	c.PersistentFlags().StringP("jobs", "J", globs.GetString("Jobs"), desc)
	desc, _, _ = globs.Desc("look")
	c.PersistentFlags().StringP("look", "L", globs.GetString("Look"), desc)
	desc, _, _ = globs.Desc("quiet")
	c.PersistentFlags().BoolP("quiet", "q", globs.GetBool("quiet"), desc)
	desc, _, _ = globs.Desc("record")
	c.PersistentFlags().StringP("record", "R", globs.GetString("record"), desc)
	desc, _, _ = globs.Desc("terse")
	c.PersistentFlags().BoolP("terse", "t", globs.GetBool("terse"), desc)
	desc, _, _ = globs.Desc("verbose")
	c.PersistentFlags().BoolP("verbose", "v", globs.GetBool("verbose"), desc)

	// NewCLIOpts: if there were opts for the dvln meta-command only they would
	// be added below, for persistent ops visible across all subcommands add
	// them above.  Put them in alphabetically ordered on the long opt name.
	// Note that you'll need to modify cmds/global.go as well otherwise your
	// globs.Desc() call and globs.GetBool("myopt") will not work.

	// The below opts apply *only* to the 'dvln' command itself, not subcommands
	desc, _, _ = globs.Desc("port")
	c.Flags().IntP("port", "P", globs.GetInt("Port"), desc)
	desc, _, _ = globs.Desc("serve")
	c.Flags().BoolP("serve", "S", globs.GetBool("serve"), desc)
	desc, _, _ = globs.Desc("version")
	c.Flags().BoolP("version", "V", globs.GetBool("version"), desc)

	c.Run = run
	if reloadCLIFlags {
		c.Flags().SetDefValueReparseOK(false)
		c.PersistentFlags().SetDefValueReparseOK(false)
	}
}

// showCLIPkgOutput is a cheesy wrapper on dumping any output coming back from
// the 'cli' (cobra) package... typically usage and such.  Note that the
// output can be dumped in JSON format or text depending upon what is needed
// FIXME: see if output that isn't help can come back when no errors seen,
//        if so then JSON outputting here might need extending (as anything
//        else will not be dumped in JSON currently, even if in JSON "look")
func showCLIPkgOutput(theOutput string, look string) {
	//FIXME: help is coming back false when -h used... 'dvln get -h -Ljson', I
	//       think this is because help isn't being properly registered by Execute()
	//help := globs.GetBool("help")
	//if help && look == "json" {
	if look == "json" {
		type helpStruct struct {
			HelpMsg   string `json:"helpMsg"`
			RecordLog string `json:"recordLog,omitempty"`
			UserID    string `json:"userId,omitempty"`
		}
		fields := make([]string, 0, 0)
		items := make([]interface{}, 0, 0)
		var usage helpStruct
		cleanOutput := api.EscapeJSONString([]byte(theOutput))
		usage.HelpMsg = cast.ToString(cleanOutput)
		fields = append(fields, "helpMsg")
		if recTgt := globs.GetString("record"); recTgt != "" {
			usage.RecordLog = recTgt
			fields = append(fields, "recordLog")
			user, err := user.Current()
			if err == nil {
				usage.UserID = user.Username
				fields = append(fields, "userId")
			}
		}
		items = append(items, &usage)
		output, fatalProblem := api.GetJSONOutput(globs.GetString("apiver"), "dvlnHelp", "usage", "regular", fields, items)
		if fatalProblem {
			out.Print(output)
			out.Exit(-1)
		}
		out.Print(output)
	} else {
		out.Print(theOutput)
	}
}

// Execute is called by main(), it basically finishes prepping the 'dvln'
// configuration data (combined with init() setting up options and available
// subcommands and such) and then kicks off the 'cli' (cobra) package to run
// subcommands and such via the dvlnCmd.Execute() call within the routine.
func Execute(args []string) int {
	Timer.Step("cmds.Execute(): init() complete (defaults set, subcmds added, CLI args set up)")

	// Shove the CLI args into the 'globs' (viper) package before we even kick
	// into the 'cli' package Execute() call below, allows us to turn on debug
	// early as well as adjust the help screen to reflect opts the user has set:
	prepCLIArgs(dvlnCmd, args)

	// Load up the users dvln config file (ie: ~/.dvlncfg/cfg.json|toml|yaml..).
	// This may alter settings/configuration further so we'll again make a pass
	// at setting up the 'out' package with any new settings:
	// Note: hold on acting on the error in case we can set up JSON, see below
	err := scanUserCfgAndReinit()

	// See if in JSON output mode, if so set the 'out' pkg appropriately...
	look := globs.GetString("look")
	if look == "json" {
		// See the 'out' pkg and the "Formatter" interface there, this allows
		// dying msgs (fatal/error+exit/issue+exit) or not-dying msgs (no exit)
		// types of output msgs to be shown in JSON format.  Basically, in JSON
		// mode "dying" messages (fatal/error/issue levels + exit) come up as
		// 'error' and non-terminal (error/issue, no exit) come up as 'warning'
		// in the JSON output (w/output level, msg & error code shown within)
		var handleJSON handleLookJSONMsgs
		out.SetFormatter(out.LevelIssue, handleJSON)
		out.SetFormatter(out.LevelError, handleJSON)
		out.SetFormatter(out.LevelFatal, handleJSON)
	}
	if err != nil {
		out.Fatal(err)
	}

	// Full opt/config file setup is now complete, wrap up any early prep of
	// the dvln tool before kicking off the 'cli' (cobra) libraries Execute()
	// method (ie: start up commands/subcommands and finish processing opts)...
	// this sets up # of CPU's to leverage, handles easy requests the user gives
	// such as what version of the tool is running (-V|--version), show settings
	// available via env or config file (--globs|-G {cfg|env}), etc.  If we did
	// handle easy requests then complete will be true and we'll bail out (note,
	// complete can also be true on error, exitVal clarifies: 0=success)
	complete, exitVal := dvlnFinalPrep()
	if complete {
		return exitVal
	}

	//dvlnCmd.DebugFlags() // can be useful for debugging purposes now and then

	// Capture 'cli' (cobra) pkg output into the cliPkgOut byte buffer, note
	// that this only affects the 'cli' (cobra) packages output (which also,
	// BTW, indirectly controls and affects the 'pflags' package used by it).
	// Do this is so we can grab/control/dump all output via the 'out' pkg:
	var cliPkgOut = new(bytes.Buffer)
	dvlnCmd.SetOutput(cliPkgOut)

	// Let the 'analysis' pkg "time" things up to here..
	Timer.Step("cmds.Execute(): loaded dvln user config, early setup and output prep done")

	// Allow 'cli' (cobra) pkg partial command matching, shortest unique match
	cli.EnablePrefixMatching = true

	// Kick off 'cli' (cobra) pkg, will parse args and the cmd/subcmd tree
	// structure and, if no help output requested or error encountered, it will
	// then kick into requested cmd PersistentPreRun, PreRun, Run, PostRun,
	// and any PersistentPostRun functions... also added a PersistentHelpRun
	// and PersistentErrRun set of funcs so if help or errors we can still
	// deal with CLI opts for debugging and verbosity and such (and recording)
	err = dvlnCmd.Execute()
	Timer.Step("cmds.Execute(): dvlnCmd.Execute() complete, post ops next")
	theOutput := cast.ToString(cliPkgOut)
	if err != nil {
		if theOutput == "" {
			err = out.WrapErr(err, "Command line processing problem", 2000)
		} else {
			// Use cli pkg output, it matches the error w/more info:
			err = out.NewErr(theOutput, 2000)
		}
	}

	if err != nil {
		out.IssueExit(int(out.ErrorExitVal()), err)
		return int(out.ErrorExitVal())
	}
	out.Debugln("CLI (cobra) package dvlnCmd.Execute() completed successfully")
	// If any output remains from the cli (cobra) pkg dump it here (eg: usage)
	if theOutput != "" {
		showCLIPkgOutput(theOutput, look)
	}
	Timer.Step("cmds.Execute(): complete")
	if err != nil {
		out.Exit(int(out.ErrorExitVal()))
		return int(out.ErrorExitVal())
	}
	return 0
}

// scanUserConfigFile initializes a viper/globs config file with sensible default
// configuration flags and sets up any activities that have been requested
// via the CLI and config settings (recording, debugging, verbosity, etc)
func scanUserConfigFile() error {
	// What config file?, default: ~/.dvlncfg/cfg.{json|toml|yaml|yml}
	// the globs package uses config.json|toml|.. and we prefer less typing
	// so we're going with cfg.json|toml|<ext> as the default name
	globs.SetConfigName("cfg")

	// Now grab the config file dir info from the 'globs' (viper) Go pkg which
	// has our globals and CLI opts and overrides set (except for the config
	// file as we haven't read it yet of course, that's what we're doing):
	configPath := globs.GetString("config")

	// Handle $HOME and ~ and such in the config file name
	configCleanPath := path.AbsPathify(configPath)
	globs.Set("configdir", "")

	// Typically config defaults to a path (dir) to look for config.<extension>
	// files in but it can also be a full path to a file, try and detect which:
    var err error
    var fileInfo os.FileInfo
	if fileInfo, err = os.Stat(configCleanPath); err == nil && fileInfo.IsDir() {
		// if it's a dir then just add the path, default looks for cfg.json|..
		globs.AddConfigPath(configPath)
		globs.Set("configdir", configCleanPath)
	} else {
		// if it's not a visible dir assume it's a file, if no file no problem
        if err == nil && !fileInfo.IsDir() {
            globs.Set("configdir", filepath.Dir(configCleanPath))
            globs.SetConfigFile(configCleanPath)
        } else {
            out.Debugln("No config file located, normal, continuing")
            configCleanPath = ""
        }
	}
    if configCleanPath != "" {
        if err := globs.ReadInConfig(); err != nil {
            return out.WrapErr(err, "Configuration package failed to read config", 2002)
        }
    }
	return nil
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
func pushCLIOptsToGlobs(c *cli.Command, topCmd bool, args []string) {
	var opts []string
	opts = args[1:]
	currErrHndl := c.Flags().ErrorHandling()

	// Tell the pflags package (used by 'cli') to ignore bad flags for this
	// early pass of flag parsing, the dvlnCmd.Execute() call will catch those
	// later, for now lets get whatever "good" CLI opts we can and turn on debug
	// or JSON mode or whatever if we can for the client.
	c.Flags().SetErrorHandling(flag.IgnoreError)

	// Parse the CLI opts into likely subcmd, flags given and any errors found.
	// As above we're ignoring errors for this early pass
	cmd, flags, _ := c.Find(opts)

	// For clean errors lets stash the top level cmd name or, if a subcmd was
	// used stash that, into the 'currentCmd' unexported package global so we
	// know what the user is running and can work (and err) with respect to that
	if currentCmd == "" {
		currentCmd = cmd.Name()
	}

	c.ParseFlags(flags)

	// Scan all flags to see what was used on CLI, shove ONLY used flags into
	// the 'globs' (viper) pkg so it's pflags and overide config levels focus
	// just on those CLI options actually used (kind of a custom hack)
	globs.SetPFlags(c.Flags())
	c.Flags().SetErrorHandling(currErrHndl)

	// If running '<cmd> <subcmd> ..' we'll also scan the <subcmd> args here
	// recursively, but if just 'mycmd ..' with no subcmd then no need
	if c.HasSubCommands() && cmd.Name() != c.Root().Name() {
		topCmd = false // this is a subcmd, not the top level cmd any longer
		pushCLIOptsToGlobs(cmd, topCmd, args)
	}
}

// doBeforeExit() uses a feature of the 'out' package where one can register
// a "deferred" function that will fire before the 'out' packages calls the
// system os.Exit() method (this assumes one is using out.Exit() and various
// calls like out.Fatal(), out.IssueExit() and out.ErrorExit() to exit, if
// so then this will be called before final exit).  In our case if the user
// is working with a temp log file and we're in text mode (msg will be "" if
// we're in JSON output mode, see "record" handling below) then add a screen
// only note about the temp log files path+name so the user can find it.
func doBeforeExit(exitVal int) {
	if tmpLogfileMsg != "" {
		// Send screen note to STDERR if currently it is the default STDOUT
		currWriter := out.Writer(out.LevelNote, out.ForScreen)
		if currWriter == os.Stdout {
			out.SetWriter(out.LevelNote, os.Stderr, out.ForScreen)

		}

		// Don't put the tmp logfile note in the logfile itself...
		currThresh := out.Threshold(out.ForLogfile)
		out.SetThreshold(out.LevelDiscard, out.ForLogfile)

		out.Noteln(tmpLogfileMsg)

		// Just to be safe set them back to previous settings
		out.SetThreshold(currThresh, out.ForLogfile)
		out.SetWriter(out.LevelNote, currWriter, out.ForScreen)
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
	if flags = os.Getenv("DVLN_STACK_TRACE_CONFIG"); flags != "" {
		if flags != "none" {
			os.Setenv("PKG_OUT_STACK_TRACE_CONFIG", flags)
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

	// Like the 'out' package init above the 'pretty' and 'api' pkgs have
	// no dependencies on 'globs' (viper) but the reverse is true (in some
	// cases)... so tell the 'api' and 'pretty' package how we want dvln
	// formatting done (note that this honors DVLN_TEXTHUMANIZE, etc)

	api.SetJSONIndentLevel(globs.GetInt("jsonindentlevel"))
	api.SetJSONRaw(globs.GetBool("jsonraw"))
	api.SetJSONPrefix(globs.GetString("jsonprefix"))

	pretty.SetHumanize(globs.GetBool("texthumanize"))
	pretty.SetOutputIndentLevel(globs.GetInt("textindentlevel"))
	pretty.SetOutputPrefixStr(globs.GetString("textprefix"))

	// Handle recording of output to logfile if set up...
	if record := globs.GetString("record"); record != "" && record != "off" {
		// If the user has requested recording/logging the below will set up
		// an io.Writer for a log file (via the 'out' package).  At this point
		// logging is enabled at the "Info/Print" (LevelInfo) level which
		// matches the default screen output setting
		//		out.SetThreshold(out.LevelInfo, out.ForLogfile)
		if record == "temp" || record == "tmp" {
			if !tmpLogfileActive {
				record = out.UseTempLogFile("dvln.")
				tmpLogfileActive = true
				tmpLogfileMsg = fmt.Sprintf("Temp output logfile: %s", record)
				if globs.GetString("look") == "json" {
					// If in JSON stash the tmp logfile
					// name as early as possible in the JSON API so it will have
					// a "note" field with the temp logfile info "ready"
					tmpMsg := api.NewMsg(tmpLogfileMsg, 101, fmt.Sprintf("%s", out.LevelNote))
					api.SetStoredNote(tmpMsg)
					// with that we are done, no need for later screen prints
					tmpLogfileMsg = ""
				} else {
					out.SetDeferFunc(doBeforeExit)
				}

				// Here we try and override what the user gave us basically by
				// replacing it with the actual tmp file name
				globs.Set("record", record)
				// Since we're replacing the CLI opt temp|tmp with the true
				// temp file name we need to "force" globs (viper) to use the
				// new value and not the pflags value (if Set() is used *and*
				// that variable was used on the CLI it'll prefer the flag
				// setting unless this little hack is done):
				// FIXME: eventually it should be a ReplacePFlag call (or
				// something like that) but I had some issues using
				// flag.Value.Set() in such a routine and it not working
				// as expected... perhaps some caching/etc, needs research.
				globs.ClearPFlag("record")
			}
		} else {
			origRecord := record
			out.SetLogFile(path.AbsPathify(record))
			// quick little hack to trim out home dir and shove in ~, keeps
			// the usage output brief if --help is used and such
			homeDir := homedir.UserHomeDir()
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
	// This is mostly used for testing, not for direct use (although one can)
	if os.Getenv("DVLN_LOGFILE_OFF") == "1" {
		out.SetThreshold(out.LevelDiscard, out.ForLogfile)
	}
}

// prepCLIArgs scans all CLI opts and tries to shove them into 'globs' (viper)
// so we can then make a pass at turning on debugging, recording, etc as early
// as possible (now) if such options are used
func prepCLIArgs(c *cli.Command, args []string) {
	// Recursively traverse dvlnCmd and all subcommands and do an early
	// "ignore errors" pass at parsing the CLI flags and shoving any valid
	// flags into the "globs" (viper) package.  What could go wrong?  ;)
	if len(args) != 1 {
		topCmd := true // passing in the top level cmd obj at this point, yes
		pushCLIOptsToGlobs(c, topCmd, args)
	}

	// Do an early output level adjustment in case CLI opts are used that will
	// add debug/record/etc info so that our adjustOutLevels() actually has a
	// chance to dump any debug/trace/etc level output calls "early", final
	// adjustments will be done with another call down below.  Early setup:
	adjustOutLevels()
}

// scanUserCfgAndReinit finishes updating the 'globs' (viper) pkg so that all
// CLI opts are fully visible and the users config file data is also loaded
// up as well, hurray!  With that data we'll also re-update dvln so that
// output data is going to the screen and any mirror logfile at the right
// output levels and such (and that any help screens reflect those final
// "full" settings from all this config data)
func scanUserCfgAndReinit() error {
	// Scan the users config file, if any, honoring any CLI opts that might
	// override the location of the user config file and push those settings
	// into the 'globs' (viper) pkg as well.  Once done the 'globs' globals will
	// be complete (Feature: except for future codebase and VCS pkg settings):
	if err := scanUserConfigFile(); err != nil {
		return err
	}

	// Final output levels adjustements to take into account any tweaks from
	// the users config file settings.  Note, don't move this below the calls
	// to the "setup*CmdCLIArgs()" routines, we need it to make final tweaks
	// to things like the --record flag before we do the final option default
	// reload.
	adjustOutLevels()

	// Now that we've read in the CLI args and the users config file we have a
	// full picture of the settings that will be used... now we'll take a 2nd
	// pass through 'cli' (cobra) and the underlying 'pflags' package it uses to
	// make the defaults for the CLI options match what the user has configured
	// or used via CLI opts and config file settings for the current tool run
	// - debatable but I like it for now, --help now reflects users full config
	reloadCLIDefaults()
	return nil
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
	setupDvlnCmdCLIArgs(dvlnCmd, reloadCLIFlags)
	setupGetCmdCLIArgs(getCmd, reloadCLIFlags)
	setupUpdateCmdCLIArgs(updateCmd, reloadCLIFlags)
	setupVersionCmdCLIArgs(versionCmd, reloadCLIFlags)
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
// into the full 'cli' (cobra) commander package 'Execute()' method.  Returns:
// - programComplete: true if error or able to wrap up users needs
// - exitVal (int): if programComplete is true, then 0 means success, non-zero
//                  means failure (ignored if programComplete is false)
func dvlnFinalPrep() (bool, int) {
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
	errExit := int(out.ErrorExitVal())
	numCPU := runtime.NumCPU()
	if jobs := globs.GetString("jobs"); jobs != "" && jobs != "all" {
		if _, err := strconv.Atoi(jobs); err != nil {
			issueMsg := fmt.Sprintf("Jobs value should be a number or 'all', found: %s\n", jobs)
			issueMsg = fmt.Sprintf("%sPlease run 'dvln help%s' for usage\n", issueMsg, cmdName)
			out.IssueExit(errExit, out.NewErr(issueMsg, 2003))
			return true, errExit
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
		out.Fatalln("Serve mode is not available yet, to contribute email 'brady@dvln.org'  :)")
		// Test runs can make fatal not really fatal, so bail out if so
		return true, 0
	}

	// Make sure that given --look|-l or cfgfile:Look or env:DVLN_LOOK are valid
	look := globs.GetString("look")
	if look != "text" && look != "json" {
		issueMsg := fmt.Sprintf("The --look option (-l) can only be set to 'text' or 'json', found: '%s'\n", look)
		issueMsg = fmt.Sprintf("%sPlease run 'dvln help%s' for usage\n", issueMsg, cmdName)
		out.IssueExit(errExit, out.NewErr(issueMsg, 2004))
		return true, errExit
	} else if look == "json" && globs.GetBool("interact") {
		out.Debugln("Interactive runs are not available for the 'json' output \"look\"")
		out.Debugln("- silently disabling interaction (client may have it set for text output)")
		globs.Set("interact", false)
	}

	// If the developer asks for the version of the tool print that out:
	if printVersion := globs.GetBool("version"); printVersion {
		out.Print(lib.DvlnVerStr())
		out.Exit(0)
		return true, 0
	}

	// If trace level debug enabled (checked inside the routine) this will dump
	// the "globs" (viper) config via the 'out.Trace*()' calls run within the
	// given method:
	globs.Debug()

	globsCLI := globs.GetString("globs")
	if globsCLI != "" && globsCLI != "env" && globsCLI != "cfg" && globsCLI != "skip" {
		issueMsg := fmt.Sprintf("The --globs option (-G) can only be set to 'env' or 'cfg', found: '%s'\n", globsCLI)
		issueMsg = fmt.Sprintf("%sPlease run 'dvln help%s' for usage\n", issueMsg, cmdName)
		out.IssueExit(errExit, out.NewErr(issueMsg, 2005))
		return true, errExit
	}
	// If the client asks for user available "globs" settable via env or cfgfile
	if globsCLI == "env" || globsCLI == "cfg" {
		out.Print(fmt.Sprintf("%v", globs.GetSingleton()))
		out.Exit(0)
		return true, 0
	}

	rootDir := ""
	var err error
	// Find the workspaces root dir, will cache it in the wkspc module, if
	// there is no workspace can be "" at this point (note: some cmds take
	// a '--wkspcdir' option so that is taken into consideration 1st)
	wkspcdir := globs.GetString("wkspcdir") // CLI option --workspace/-w
	if wkspcdir != "." && wkspcdir != "" {
		rootDir, err = wkspc.RootDir(wkspcdir) // handle a CLI given dir/subdir
		out.Traceln("Workspace dir passed in:", wkspcdir)
	} else {
		out.Traceln("No workspace dir passed in, driving off of cwd")
		rootDir, err = wkspc.RootDir() // scan from CWD for a wkspc root dir
	}
	out.Debugln("Workspace root dir:", rootDir)
	if err != nil {
		// err means a failure (no workspace is not an error, it's normal)
		out.ErrorExit(errExit, out.WrapErr(err, "Unexpected problem scanning for a workspace", 2006))
		return true, errExit
	}
	// This will bootstrap the workspace root dir directory structure
	if err = wkspc.SetRootDir(rootDir); err != nil {
		out.ErrorExit(errExit, out.WrapErr(err, "Unexpected problem setting the workspace root dir", 2007))
		return true, errExit
	}
	return false, 0
}

type handleLookJSONMsgs struct{}

// FormatMessage in this context implements the Formatter interface from
// the 'out' package.  This is only called if in JSON mode and only for
// output levels Issue, Error and Fatal (note: these are double checked
// in the method as well "just in case").  If JSON output desired we'll
// handle terminal issues/errors/fatals and non-terminal issue/errors
// as follows:
// - if it's non-fatal then "register" the issue/error with the 'api' pkg
// so the JSON that is dumped at the end of our run includes a 'warning'
// which will give an error code and an 'out' package output level
// - if it is fatal then generate a full JSON error response using an 'api'
// pkg call and then control the 'out' pkg so a pure JSON error is displayed
// before dying off within the 'out' package.
//
// The idea is that the client wants JSON output fmt and should always get
// it even if any error has occurred (and we'll want to inform them of any
// non-fatal issues as well via the JSON 'warning' field to use as they
// see fit)
func (f handleLookJSONMsgs) FormatMessage(msg string, outLevel out.Level, code int, stack string, dying bool) (string, int, bool) {
	suppressOutputMask := 0
	suppressNativePrefixing := false
	look := globs.GetString("look")
	if look != "json" || outLevel < out.LevelIssue {
		return msg, suppressOutputMask, suppressNativePrefixing
	}
	problemMsg := api.NewMsg(msg, code, fmt.Sprintf("%s", outLevel))
	if dying {
		suppressNativePrefixing = true
		msg = api.FatalJSONMsg(globs.GetString("apiver"), problemMsg)
	} else {
		// Suppress output to the screen and the logfile, our final JSON dump
		// will include the warning and that will also go to the logfile (assuming
		// it has been set up of course)
		suppressOutputMask = out.ForBoth
		suppressNativePrefixing = true
		api.SetStoredNonFatalWarning(problemMsg, int(out.DefaultErrCode()))
		msg = ""
	}
	return msg, suppressOutputMask, suppressNativePrefixing
}

// run for the dvln cmd really doesn't do anything but recommend the user
// select a subcommand to run (as a raw dvln cmd with no subcommands should
// not get this far unless someone forgot to add a subcmd).  When a subcmd
// is used then the 'run()' method for that subcommand will be executed
// instead of this one (and all top level 'dvln' options like --globs=cfg
// are handled before we get this far).
func run(cmd *cli.Command, args []string) {
	out.IssueExit(-1, out.NewErr("Please use a valid subcommand (for a list: 'dvln help')", 2001))
}
