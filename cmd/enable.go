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

	log "github.com/sirupsen/logrus"

	"github.com/bcambl/rpda/internal/pkg/rpa"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// enableCmd represents the enable command
var enableCmd = &cobra.Command{
	Use:   "enable",
	Short: "Enable direct access mode for the latest copy",
	Long: `Enable direct access mode for the latest copy
examples:

rpda enable --group EXAMPLE_CG --latest-test

rpda enable --group EXAMPLE_CG --latest-dr

rpda enable --all --latest-test

rpda enable --all --latest-dr

	`,
	Run: func(cmd *cobra.Command, args []string) {

		a := &rpa.App{}
		a.RPAURL = viper.GetString("api.url")
		a.Username = viper.GetString("api.username")
		a.Password = viper.GetString("api.password")
		a.Delay = viper.GetInt("api.delay")
		a.Debug = viper.GetBool("debug")
		a.Identifiers.ProductionNode = viper.GetString("identifiers.production_node_name_contains")
		a.Identifiers.CopyNode = viper.GetString("identifiers.dr_copy_name_contains")
		a.Identifiers.TestCopy = viper.GetString("identifiers.test_copy_name_contains")

		group, err := cmd.Flags().GetString("group")
		if err != nil {
			log.Fatal(err)
		}
		latestTest, err := cmd.Flags().GetBool("latest-test")
		if err != nil {
			log.Fatal(err)
		}
		latestDR, err := cmd.Flags().GetBool("latest-dr")
		if err != nil {
			log.Fatal(err)
		}
		all, err := cmd.Flags().GetBool("all")
		if err != nil {
			log.Fatal(err)
		}

		if a.Debug {
			a.Debugger()
			fmt.Println("enable command 'group' flag value: ", group)
			fmt.Println("enable command 'latest-test' flag value: ", latestTest)
			fmt.Println("enable command 'latest-dr' flag value: ", latestDR)
			fmt.Println("enable command 'all' flag value: ", all)
		}

		// preflight checks

		// ensure group or all flags were provided
		if all == false && group == "" {
			cmd.Usage()
			os.Exit(1)
		}
		a.Group = group

		// ensure A image copy flag was provided
		if latestTest == false && latestDR == false {
			cmd.Usage()
			os.Exit(1)
		}

		// ensure user did not provide BOTH image copy flags
		if latestTest == true && latestDR == true {
			cmd.Usage()
			os.Exit(1)
		}

		// assign the image copy flag
		if latestDR == true {
			a.Copy = a.Identifiers.CopyNode
		}
		if latestTest == true {
			a.Copy = a.Identifiers.TestCopy
		}

		if group != "" {
			// display status of single group if a group name was provided
			a.StartOne()
		} else if all {
			// display status of all groups if the all flag was provided
			a.StartAll()
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
	enableCmd.PersistentFlags().Bool("latest-test", false, "Use Latest Test Copy Image")
	enableCmd.PersistentFlags().Bool("latest-dr", false, "Use Latest DR Copy Image")
}
