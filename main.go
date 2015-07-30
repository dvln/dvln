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

// package main immediatly kicks off the 'cli' management package (cobra
// commander created by spf13) which manages subcommands and opts and such.
package main

import (
	"os"

	"github.com/dvln/dvln/cmds"
	"github.com/dvln/out"
)

func main() {
	// Kick off the the 'cli' mgmt package (Cobra commander) for the dvln
	// command and the various subcommands and opts:
	exitVal := cmds.Execute(os.Args)
	out.Exit(exitVal)
	// Note, the 2nd exit shouldn't happen but in case someone told
	//       the 'out' pkg to bypass exitting (for test) lets exit now:
	os.Exit(exitVal)
}
