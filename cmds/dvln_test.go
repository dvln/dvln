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

package cmds

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"strings"
	"testing"

	cli "github.com/dvln/cobra"
	"github.com/dvln/out"
)

type resulter struct {
	Error   error
	Output  string
	Command *cli.Command
}

func initialize() *cli.Command {
	var c = dvlnCmd
	return c
}

func setupDvlnCmdTest(input string) resulter {
	c := initialize()
	return simpleTester(c, input)
}

// sipmleTester passes the "input" (command line flags/opts) to the 'dvln' tool
func simpleTester(c *cli.Command, input string) resulter {
	// The output from globs (cobra) is set to a buffer, also the 'out' output
	// package is redirected to the same buffer
	buf := new(bytes.Buffer)

	// Tell cli pkg (cobra) the args we are passing on the "fake" command line:
	args := strings.Split(input, " ")
	c.SetArgs(args)
	fullArgs := []string{"dvln"}
	fullArgs = append(fullArgs, args...)

	// In case the user running tests has mucked with their own log file
	// config settings lets just forcibly turn off log file writing for now
	os.Setenv("DVLN_LOGFILE_OFF", "1")

	// Use the dvln.go Execute() method to fire up the globs (cobra) Execute()
	// method and check on it's results and such, all output from the local
	// Execute() call goes through the 'out' pkg so we'll just grab all screen
	// output into our io.Writer buf buffer and then fire it up:
	out.SetWriter(out.LevelAll, buf, out.ForScreen)
	Execute(fullArgs)

	// Flip it back on in case it's desired for some reason (another method/etc)
	os.Setenv("DVLN_LOGFILE_OFF", "0")

	// Turn the output into a single string w/newlines included, return results
	output := buf.String()
	return resulter{nil, output, c}
}

func logErr(t *testing.T, found, expected string) {
	out := new(bytes.Buffer)

	_, _, line, ok := runtime.Caller(2)
	if ok {
		fmt.Fprintf(out, "Line: %d ", line)
	}
	fmt.Fprintf(out, "Unexpected response.\nExpecting to contain: \n %q\nGot:\n %q\n", expected, found)
	t.Errorf(out.String())
}

func checkResultContains(t *testing.T, x resulter, check string) {
	if !strings.Contains(x.Output, check) {
		logErr(t, x.Output, check)
	}
}

func checkResultOmits(t *testing.T, x resulter, check string) {
	if strings.Contains(x.Output, check) {
		logErr(t, x.Output, check)
	}
}

// TestNoArgs sees what dvln does with no args at all...
func TestNoArgs(t *testing.T) {
	os.Setenv("PKG_OUT_NO_EXIT", "1")
	x := setupDvlnCmdTest("")
	os.Setenv("PKG_OUT_NO_EXIT", "0")
	checkResultContains(t, x, "Issue #2001: Please use a valid subcommand")
	//eriknow, coming back double
}

// TestNoArgsCommand sees what dvln does with no args at all...
func TestBogusArgsCommand(t *testing.T) {
	os.Setenv("PKG_OUT_NO_EXIT", "1")
	x := setupDvlnCmdTest("--baloney")
	os.Setenv("PKG_OUT_NO_EXIT", "0")
	checkResultContains(t, x, "Issue #2000: Error: unknown flag: --baloney")

}

// TestHelpInterface runs through basic help for the dvln meta-cmd to see if the
// three entry points work, note that each subcommands will test it's own help
func TestHelpInterface(t *testing.T) {
	x := setupDvlnCmdTest("help")
	checkResultContains(t, x, "dvln: Multi-package development line and workspace management tool")
	checkResultContains(t, x, "Available Commands:")
	checkResultContains(t, x, "get packages")
	checkResultContains(t, x, "-D, --debug           control debug output")
	x = setupDvlnCmdTest("--help")
	checkResultContains(t, x, "dvln: Multi-package development line and workspace management tool")
	checkResultContains(t, x, "Available Commands:")
	checkResultContains(t, x, "get packages")
	checkResultContains(t, x, "-D, --debug           control debug output")
	x = setupDvlnCmdTest("-h")
	checkResultContains(t, x, "dvln: Multi-package development line and workspace management tool")
	checkResultContains(t, x, "Available Commands:")
	checkResultContains(t, x, "get packages")
	checkResultContains(t, x, "-D, --debug           control debug output")
	x = setupDvlnCmdTest("-hLjson")
	checkResultContains(t, x, "\"apiVersion\": ")
	checkResultContains(t, x, "\"id\": 0,")
	checkResultContains(t, x, "\"kind\": \"usage\"")
	checkResultContains(t, x, "\"verbosity\": ")
	checkResultContains(t, x, "\"items\": [")
	checkResultContains(t, x, "\"helpMsg\":")
	checkResultContains(t, x, "dvln: Multi-package development line and workspace management tool")
	checkResultContains(t, x, "\"userId\":")
	// Now lets see if the JSON returned was good
	var result interface{}
	err := json.Unmarshal([]byte(x.Output), &result)
	if err != nil {
		t.Fatalf("Unable to unmarshal JSON: -hLjson options used: %s", err)
	}
	// Flip it back off so later tests don't have it turned on
	x = setupDvlnCmdTest("-h -Ltext")
	checkResultContains(t, x, "dvln: Multi-package development line and workspace management tool")
	checkResultOmits(t, x, "\"apiVersion\": ")
}

func TestAnalysisArg(t *testing.T) {
	// We'll combine it with help output for a few
	x := setupDvlnCmdTest("-Ah")
	checkResultContains(t, x, "cmds.Execute(): loaded dvln user config, early setup and output prep done:")
	checkResultContains(t, x, "cmds.Execute(): complete:")
	checkResultContains(t, x, "Allocs")
	x = setupDvlnCmdTest("--analysis --help")
	checkResultContains(t, x, "cmds.Execute(): loaded dvln user config, early setup and output prep done:")
	checkResultContains(t, x, "cmds.Execute(): complete:")
	checkResultContains(t, x, "Allocs")

	x = setupDvlnCmdTest("--analysis=false --help")
	checkResultOmits(t, x, "cmds.Execute(): loaded dvln user config, early setup and output prep done:")
	// The 1st pass to reverse the analysis only kicks on after Execute() runs
	// so this one should still be there as it's run before dvlnCmd.Execute()
	// fires (ie: this happens at the top of cmds.Execute()):
	checkResultContains(t, x, "init() complete (defaults set, subcmds added, CLI args set up):")

	x = setupDvlnCmdTest("--analysis=false --help")
	// The 2nd false pass should no longer contain that since analysis is now
	// false as we go into dvlnCmd.Execute(), so no print:
	checkResultOmits(t, x, "init() complete (defaults set, subcmds added, CLI args set up):")
}

func TestDebugArg(t *testing.T) {
	// We'll combine it with help output for a few samples
	x := setupDvlnCmdTest("-Dh")
	checkResultContains(t, x, "Debug: CLI (cobra) package dvlnCmd.Execute() completed successfully")
	x = setupDvlnCmdTest("--debug --help")
	checkResultContains(t, x, "Debug: CLI (cobra) package dvlnCmd.Execute() completed successfully")
	x = setupDvlnCmdTest("--debug=false --help")
	checkResultOmits(t, x, "Debug: CLI (cobra) package dvlnCmd.Execute() completed successfully")
	os.Setenv("PKG_OUT_NO_EXIT", "1")
	// turn off help for the next tests... and, while doing so, make sure that
	// option works as expected as well
	x = setupDvlnCmdTest("--help=false")
	checkResultContains(t, x, "Issue #2001: Please use a valid subcommand")
	os.Setenv("PKG_OUT_NO_EXIT", "0")
}

// TestVersionFunctionality runs through basic version checks, both the
// subcmd and the --version and related options, including different levels
// of verbosity in text and JSON output mode.
func TestVersionFunctionality(t *testing.T) {
	os.Setenv("PKG_OUT_NO_EXIT", "1")
	x := setupDvlnCmdTest("version")
	checkResultContains(t, x, "Version: ")
	checkResultContains(t, x, "API Rev: ")
	checkResultContains(t, x, "Build Date: ")
	checkResultOmits(t, x, "Exec Name: ")
	x = setupDvlnCmdTest("version -v")
	checkResultContains(t, x, "Version: ")
	checkResultContains(t, x, "API Rev: ")
	checkResultContains(t, x, "Build Date: ")
	checkResultContains(t, x, "Exec Name: ")
	x = setupDvlnCmdTest("version -t --verbose=false")
	checkResultContains(t, x, "Version: ")
	checkResultOmits(t, x, "API Rev: ")
	checkResultOmits(t, x, "Build Date: ")
	checkResultOmits(t, x, "Exec Name: ")
	x = setupDvlnCmdTest("--version --terse=false")
	checkResultContains(t, x, "Version: ")
	checkResultContains(t, x, "API Rev: ")
	checkResultContains(t, x, "Build Date: ")
	checkResultOmits(t, x, "Exec Name: ")
	x = setupDvlnCmdTest("-Vt")
	checkResultContains(t, x, "Version: ")
	checkResultOmits(t, x, "API Rev: ")
	checkResultOmits(t, x, "Build Date: ")
	checkResultOmits(t, x, "Exec Name: ")
	x = setupDvlnCmdTest("-vVLjson --terse=false")
	checkResultContains(t, x, "\"apiVersion\": ")
	checkResultContains(t, x, "\"id\": 0,")
	checkResultContains(t, x, "\"kind\": \"version\"")
	checkResultContains(t, x, "\"verbosity\": \"verbose\",")
	checkResultContains(t, x, "\"fields\": [")
	checkResultContains(t, x, "\"items\": [")
	checkResultContains(t, x, "\"toolVersion\":")
	checkResultContains(t, x, "\"apiVersion\":")
	checkResultContains(t, x, "\"buildDate\":")
	checkResultContains(t, x, "\"execName\":")

	// Now lets see if the JSON returned was good
	var result interface{}
	err := json.Unmarshal([]byte(x.Output), &result)
	if err != nil {
		t.Fatalf("Unable to unmarshal JSON: -vVLjson options used: %s", err)
	}
	// Flip it back off so later tests don't have it turned on
	x = setupDvlnCmdTest("-VLtext --verbose=false --terse==false")
	checkResultContains(t, x, "Version: ")
	checkResultContains(t, x, "API Rev: ")
	x = setupDvlnCmdTest("--version=false --help=false")
	checkResultContains(t, x, "Issue #2001: Please use a valid subcommand")
	os.Setenv("PKG_OUT_NO_EXIT", "0")
}

// TestGlobsFunctionality runs through basic version checks, both the
// subcmd and the --version and related options, including different levels
// of verbosity in text and JSON output mode.
func TestGlobsFunctionality(t *testing.T) {
	os.Setenv("PKG_OUT_NO_EXIT", "1")
	x := setupDvlnCmdTest("--globs")
	checkResultContains(t, x, "Issue #2000: Error: flag needs an argument: --globs")
	x = setupDvlnCmdTest("-vG=env")
	checkResultContains(t, x, "DVLN_ANALYSIS: ")
	checkResultContains(t, x, "  Description: memory and timing analytics")
	checkResultContains(t, x, "  Use Level:   EXPERT")
	checkResultContains(t, x, "  Value:       false")
	checkResultContains(t, x, "DVLN_DEBUG: ")
	checkResultContains(t, x, "  Description: control debug output")
	checkResultContains(t, x, "  Use Level:   STANDARD")
	checkResultContains(t, x, "  Value:       false")
	x = setupDvlnCmdTest("-tGenv --verbose=false")
	checkResultContains(t, x, "DVLN_ANALYSIS: ")
	checkResultOmits(t, x, "  Description: memory and timing analytics")
	checkResultOmits(t, x, "  Use Level:   EXPERT")
	checkResultContains(t, x, "  Value: false")
	checkResultContains(t, x, "DVLN_DEBUG: ")
	checkResultOmits(t, x, "  Description: control debug output")
	checkResultOmits(t, x, "  Use Level:   STANDARD")
	checkResultContains(t, x, "  Value: false")
	x = setupDvlnCmdTest("--terse=false --globs=env")
	checkResultContains(t, x, "DVLN_DEBUG: ")
	checkResultContains(t, x, "  Description: control debug output")
	checkResultOmits(t, x, "  Use Level:   STANDARD")
	checkResultContains(t, x, "  Value:       false")
	x = setupDvlnCmdTest("--globs=cfg")
	checkResultContains(t, x, "analysis: ")
	checkResultContains(t, x, "  Description: memory and timing analytics")
	checkResultOmits(t, x, "  Use Level:   EXPERT")
	checkResultContains(t, x, "  Value:       false")
	checkResultContains(t, x, "codebase: ")
	checkResultContains(t, x, "  Description: codebase name or URL")
	checkResultOmits(t, x, "  Use Level:   NOVICE")
	checkResultContains(t, x, "  Value:       ")
	x = setupDvlnCmdTest("--terse --globs=cfg")
	checkResultContains(t, x, "analysis: ")
	checkResultOmits(t, x, "  Description: memory and timing analytics")
	checkResultOmits(t, x, "  Use Level:   EXPERT")
	checkResultContains(t, x, "  Value: false")
	checkResultContains(t, x, "codebase: ")
	checkResultOmits(t, x, "  Description: codebase name or URL")
	checkResultOmits(t, x, "  Use Level:   NOVICE")
	checkResultContains(t, x, "  Value: ")
	x = setupDvlnCmdTest("--terse=false --verbose --globs=cfg")
	checkResultContains(t, x, "analysis: ")
	checkResultContains(t, x, "  Description: memory and timing analytics")
	checkResultContains(t, x, "  Use Level:   EXPERT")
	checkResultContains(t, x, "  Value:       false")
	checkResultContains(t, x, "codebase: ")
	checkResultContains(t, x, "  Description: codebase name or URL")
	checkResultContains(t, x, "  Use Level:   NOVICE")
	checkResultContains(t, x, "  Value:       ")
	x = setupDvlnCmdTest("-G=blah")
	checkResultContains(t, x, "Issue #2005: The --globs option (-G) can only be set to 'env' or 'cfg'")
	x = setupDvlnCmdTest("-vGcfg -Ljson")
	checkResultContains(t, x, "\"apiVersion\": ")
	checkResultContains(t, x, "\"id\": 0,")
	checkResultContains(t, x, "\"kind\": \"cfg\"")
	checkResultContains(t, x, "\"verbosity\": \"verbose\",")
	checkResultContains(t, x, "\"fields\": [")
	checkResultContains(t, x, "\"(name)\",")
	checkResultContains(t, x, "\"description\",")
	checkResultContains(t, x, "\"useLevel\",")
	checkResultContains(t, x, "\"value\"")
	checkResultContains(t, x, "\"startIndex\": 1,")
	checkResultContains(t, x, "\"items\": [")
	checkResultContains(t, x, "\"analysis\": {") // } (for matching)
	checkResultContains(t, x, "\"description\": \"memory and timing analytics\",")
	checkResultContains(t, x, "\"useLevel\": \"EXPERT\",")
	checkResultContains(t, x, "\"value\": false")
	// Now lets see if the JSON returned was good
	var result interface{}
	err := json.Unmarshal([]byte(x.Output), &result)
	if err != nil {
		t.Fatalf("Unable to unmarshal JSON: -vGcfg -Ljson options used: %s", err)
	}
	x = setupDvlnCmdTest("--verbose=false -tGenv -Ljson")
	checkResultContains(t, x, "\"apiVersion\": ")
	checkResultContains(t, x, "\"id\": 0,")
	checkResultContains(t, x, "\"kind\": \"env\"")
	checkResultContains(t, x, "\"verbosity\": \"terse\",")
	checkResultContains(t, x, "\"fields\": [")
	checkResultContains(t, x, "\"(name)\",")
	checkResultOmits(t, x, "\"description\",")
	checkResultOmits(t, x, "\"useLevel\",")
	checkResultContains(t, x, "\"value\"")
	checkResultContains(t, x, "\"startIndex\": 1,")
	checkResultContains(t, x, "\"items\": [")
	checkResultContains(t, x, "\"DVLN_ANALYSIS\": {") // } (for matching)
	checkResultOmits(t, x, "\"description\": \"memory and timing analytics\",")
	checkResultOmits(t, x, "\"useLevel\": \"EXPERT\",")
	checkResultContains(t, x, "\"value\": false")
	// Now lets see if the JSON returned was good
	err = json.Unmarshal([]byte(x.Output), &result)
	if err != nil {
		t.Fatalf("Unable to unmarshal JSON: --verbose=false -tGenv -Ljson options used: %s", err)
	}

	// Flip it back off so later tests don't have it turned on
	x = setupDvlnCmdTest("-Gcfg -Ltext --terse=false --verbose=false")
	checkResultOmits(t, x, "\"apiVersion\": ")
	checkResultContains(t, x, "analysis: ")
	checkResultContains(t, x, "  Description: memory and timing analytics")
	checkResultOmits(t, x, "  Use Level:   EXPERT")
	checkResultContains(t, x, "  Value:       false")
	// Use a cheesy test specific value, "skip", to tells the tool to ignore the
	// setting without fully clearing it (blasting the cli setting or unsetting
	// it is a pain currently so this "cheats" and bypasses that need)
	x = setupDvlnCmdTest("-Gskip")
	checkResultContains(t, x, "Issue #2001: Please use a valid subcommand")
	os.Setenv("PKG_OUT_NO_EXIT", "0")
}
