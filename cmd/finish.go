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
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/bcambl/rpda/internal/pkg/rpa"
	"github.com/spf13/cobra"
)

// finishCmd represents the finish command
var finishCmd = &cobra.Command{
	Use:   "finish",
	Short: "Return a conistency group to a full replication state",
	Long: `Return a conistency group to a full replication state
examples:

rpda finish --group EXAMPLE_CG --test

rpda finish --all --test

	`,
	Run: func(cmd *cobra.Command, args []string) {

		// Load API Configuration
		c := &rpa.Config{}
		c.Load()

		// Load Consistency Group Name Identifiers
		i := &rpa.Identifiers{}
		i.Load()

		a := &rpa.App{}
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

		log.Debug("finish command 'group' flag value: ", group)
		log.Debug("enable command 'copy' flag value: ", copyByName)
		log.Debug("finish command 'test' flag value: ", testCopy)
		log.Debug("finish command 'dr' flag value: ", drCopy)
		log.Debug("finish command 'all' flag value: ", all)

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

		// if an exact copy name provided, ensure A image copy flag provided
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
			a.FinishOne()
		} else if all {
			// display status of all groups if the all flag was provided
			a.FinishAll()
		} else {
			// otherwise, display command usage
			cmd.Usage()
		}

	},
}

func init() {
	rootCmd.AddCommand(finishCmd)

	// command flags and configuration settings.
	finishCmd.PersistentFlags().Bool("all", false, "Display Status for All Consistency Groups")
	finishCmd.PersistentFlags().String("group", "", "Display Status of Consistency Group by Name")
	finishCmd.PersistentFlags().String("copy", "", "Use Latest Test Copy Image By Name (only usable with --group)")
	finishCmd.PersistentFlags().Bool("test", false, "Use Latest Test Copy Image")
	finishCmd.PersistentFlags().Bool("dr", false, "Use Latest DR Copy Image")
}
