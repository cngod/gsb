// Copyright 2016 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Command sblookup is a tool for looking up URLs via the command-line.
//
// The tool reads one URL per line from STDIN and checks every URL against
// the Safe Browsing API. The "Safe" or "Unsafe" verdict is printed to STDOUT.
// If an error occurred, debug information may be printed to STDERR.
//
// To build the tool:
//	$ go get github.com/google/safebrowsing/cmd/sblookup
//
// Example usage:
//	$ sblookup -apikey $APIKEY
//	https://google.com
//	Safe URL: https://google.com
//	http://bad1url.org
//	Unsafe URL: [{bad1url.org {MALWARE ANY_PLATFORM URL_EXPRESSION}}]
package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"

	"github.com/google/safebrowsing"
)

var (
	apiKeyFlag   = flag.String("apikey", "", "specify your Safe Browsing API key")
	databaseFlag = flag.String("db", "", "path to the Safe Browsing database. By default persistent storage is disabled (not recommended).")
)

const usage = `sblookup: command-line tool to lookup URLs with Safe Browsing.

Tool reads one URL per line from STDIN and checks every URL against the
Safe Browsing API. The Safe or Unsafe verdict is printed to STDOUT. If an error
occurred, debug information may be printed to STDERR.

Exit codes:
  0     if all URLs were looked up an are safe.
  1     if at least one URL is not safe.
  128   if at least one URL lookup failed.

Usage: %s -apikey=$APIKEY

`

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, usage, os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()
	if *apiKeyFlag == "" {
		fmt.Fprintln(os.Stderr, "No -apikey specified")
		os.Exit(1)
	}
	sb, err := safebrowsing.NewSafeBrowser(safebrowsing.Config{
		APIKey: *apiKeyFlag,
		DBPath: *databaseFlag,
		Logger: os.Stderr,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "Unable to initialize Safe Browsing client: ", err)
		os.Exit(1)
	}

	scanner := bufio.NewScanner(os.Stdin)
	code := 0
	for scanner.Scan() {
		url := scanner.Text()
		threats, err := sb.LookupURLs([]string{url})
		if err != nil {
			fmt.Fprintln(os.Stderr, "Lookup error:", err)
			if code != 0 {
				code = 128 // Invalid argument.
			}
		}
		if len(threats[0]) == 0 {
			fmt.Fprintln(os.Stdout, "Safe URL:", url)
		} else {
			fmt.Fprintln(os.Stdout, "Unsafe URL:", threats[0])
			if code != 0 {
				code = 1
			}
		}
	}
	if scanner.Err() != nil {
		fmt.Fprintln(os.Stderr, "Unable to read input:", scanner.Err())
		if code != 0 {
			code = 128 // Invalid argument.
		}
	}
	os.Exit(code)
}