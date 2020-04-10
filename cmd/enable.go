package cmd

/*
Copyright © 2020 Blayne Campbell
All rights reserved.

Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions are met:

1. Redistributions of source code must retain the above copyright notice,
   this list of conditions and the following disclaimer.

2. Redistributions in binary form must reproduce the above copyright notice,
   this list of conditions and the following disclaimer in the documentation
   and/or other materials provided with the distribution.

3. Neither the name of the copyright holder nor the names of its contributors
   may be used to endorse or promote products derived from this software
   without specific prior written permission.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE
LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
POSSIBILITY OF SUCH DAMAGE.
*/

import (
	"fmt"
	"os"

	"github.com/bcambl/rpda/internal/pkg/rp"
	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
)

// enableCmd represents the enable command
var enableCmd = &cobra.Command{
	Use:   "enable",
	Short: "Enable direct access mode for the latest copy",
	Long: `Enable direct access mode for the latest copy
examples:

rpda enable --group EXAMPLE_CG --copy Test_Copy

rpda enable --group EXAMPLE_CG --test

rpda enable --group EXAMPLE_CG --dr

rpda enable --all --test

rpda enable --all --dr

	`,
	Run: func(cmd *cobra.Command, args []string) {

		// Load API Configuration
		c := &rp.Config{}
		c.Load(cmd)

		// Load Consistency Group Name Identifiers
		i := &rp.Identifiers{}
		i.Load(cmd)

		a := &rp.App{}
		a.Config = c
		a.Identifiers = i

		group, err := cmd.Flags().GetString("group")
		if err != nil {
			log.Fatal(err)
		}
		copyByName, err := cmd.Flags().GetString("copy")
		if err != nil {
			log.Fatal(err)
		}
		testCopy, err := cmd.Flags().GetBool("test")
		if err != nil {
			log.Fatal(err)
		}
		drCopy, err := cmd.Flags().GetBool("dr")
		if err != nil {
			log.Fatal(err)
		}
		all, err := cmd.Flags().GetBool("all")
		if err != nil {
			log.Fatal(err)
		}

		if a.Config.Debug {
			a.Debugger()
			fmt.Println("enable command 'group' flag value: ", group)
			fmt.Println("enable command 'copy' flag value: ", copyByName)
			fmt.Println("enable command 'test' flag value: ", testCopy)
			fmt.Println("enable command 'dr' flag value: ", drCopy)
			fmt.Println("enable command 'all' flag value: ", all)
		}

		// preflight checks

		// ensure group or all flags were provided
		if all == false && group == "" {
			log.Error("Either --all or --group must be specified.")
			cmd.Usage()
			os.Exit(1)
		}

		// if --all flag was specified, --copy cannot be used
		if all == true && copyByName != "" {
			log.Error("--copy cannot be used with --all")
			cmd.Usage()
			os.Exit(1)
		}

		a.Group = group
		a.CopyName = copyByName

		// if an exact copy name was not provided, ensure an image copy flag was provided
		if copyByName == "" && testCopy == false && drCopy == false {
			if all {
				log.Error("One of --test or --dr must be specified")
			} else {
				log.Error("One of --test --dr or --copy must be specified")
			}
			cmd.Usage()
			os.Exit(1)
		}

		// also ensure user did not provide BOTH image copy flags
		if copyByName != "" && (testCopy == true || drCopy == true) {
			log.Error("--copy cannot be combined with --test or --dr")
			cmd.Usage()
			os.Exit(1)
		}

		if drCopy == true {
			a.CopyRegexp = a.Identifiers.CopyNodeRegexp
		}
		if testCopy == true {
			a.CopyRegexp = a.Identifiers.TestNodeRegexp
		}

		if group != "" {
			// display status of single group if a group name was provided
			a.EnableOne()
		} else if all {
			// display status of all groups if the --all flag was provided
			a.EnableAll()
		} else {
			// otherwise, display command usage
			cmd.Usage()
		}

	},
}

func init() {
	rootCmd.AddCommand(enableCmd)

	// command flags and configuration settings.
	enableCmd.PersistentFlags().Bool("all", false, "Enable Direct Image Access for All Consistency Groups")
	enableCmd.PersistentFlags().String("group", "", "Enable Direct Image Access for Consistency Group by Name")
	enableCmd.PersistentFlags().String("copy", "", "Use Latest Test Copy Image By Name (only usable with --group)")
	enableCmd.PersistentFlags().Bool("test", false, "Use Latest Test Copy Image")
	enableCmd.PersistentFlags().Bool("dr", false, "Use Latest DR Copy Image")
}
