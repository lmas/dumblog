// Copyright © 2021 Alex
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/lmas/dumblog/internal"
)

var (
	initDir  = "./example" // Default dir for the "init" command
	confOut  = flag.String("out", "./public", "Output dir for generated site")
	confAddr = flag.String("addr", "127.0.0.1:8080", "Local IP address for hosting the demo web server")
)

type cmd struct {
	Name string
	Func func()
	Help string
}

var commands []cmd

func main() {
	flag.Parse()

	commands = []cmd{
		{"init", runInit, "Writes an example template (default output dir is `./example`)"},
		{"update", runUpdate, "Regenerate the static site"},
		{"web", runWeb, "Run a demo web server"},
		{"version", printVersion, "Print version and exit"},
		{"help", printHelp, "Print this help message and exit"},
	}

	cmd := flag.Arg(0)
	for _, c := range commands {
		if cmd == c.Name {
			c.Func()
			return
		}
	}
	printHelp()
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func print(msg string, args ...interface{}) {
	fmt.Printf(msg+"\n", args...)
}

func printFatal(msg string, args ...interface{}) {
	print(msg, args...)
	os.Exit(1)
}

func printVersion() {
	print("%s v%s", internal.Name, internal.Version)
}

func printHelp() {
	printVersion()

	print("\nFlags:")
	flag.CommandLine.SetOutput(os.Stdout) // to match the output of print()/fmt.Printf()
	flag.PrintDefaults()

	print("\nCommands:")
	for _, c := range commands {
		print("  %s\n  \t%s", c.Name, c.Help) // Matches indentation with PrintDefaults()
	}
	print("")
}

func runInit() {
	if err := internal.CreateTemplate(initDir); err != nil {
		printFatal("Error creating template: %s", err)
	}
	print("Wrote %s", initDir)
}

func runWeb() {
	handler := http.FileServer(http.Dir(*confOut))
	print("Running on http://%s", *confAddr)
	if err := http.ListenAndServe(*confAddr, handler); err != nil {
		printFatal("Error running web server: %s", err)
	}
}

func runUpdate() {
	dir := flag.Arg(1)
	gen := internal.New()

	if err := gen.ReadTemplate(dir); err != nil {
		printFatal("Error reading %q: %s", dir, err)
	}

	if err := gen.ExecuteTemplate(*confOut); err != nil {
		printFatal("Error writing %q: %s", *confOut, err)
	}
	print("Wrote %s", *confOut)
}
