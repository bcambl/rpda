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

// finishCmd represents the finish command
var finishCmd = &cobra.Command{
	Use:   "finish",
	Short: "Return a conistency group to a full replication state",
	Long: `Return a conistency group to a full replication state
examples:

rpda finish --group EXAMPLE_CG

rpda finish --all

	`,
	Run: func(cmd *cobra.Command, args []string) {

		a := &rpa.App{}
		a.RPAURL = viper.GetString("api.url")
		a.Username = viper.GetString("api.username")
		a.Password = viper.GetString("api.password")
		a.Delay = viper.GetInt("api.delay")

		group, err := cmd.Flags().GetString("group")
		if err != nil {
			log.Fatal(err)
		}
		all, err := cmd.Flags().GetBool("all")
		if err != nil {
			log.Fatal(err)
		}

		if viper.GetBool("debug") {
			a.Debug()
			fmt.Println("status command 'group' flag value: ", group)
			fmt.Println("status command 'all' flag value: ", all)
		}

		// preflight checks

		// ensure group or all flags were provided
		if all == false && group == "" {
			cmd.Usage()
			os.Exit(1)
		}
		a.Group = group

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
}
