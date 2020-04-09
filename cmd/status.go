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

	log "github.com/sirupsen/logrus"

	"github.com/bcambl/rpda/internal/pkg/rp"
	"github.com/spf13/cobra"
)

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Display Consistency Group Status",
	Long: `Display Consistency Group Status
examples:

rpda status --all

rpda status --group Example_CG

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
		all, err := cmd.Flags().GetBool("all")
		if err != nil {
			log.Fatal(err)
		}

		if a.Config.Debug {
			a.Debugger()
			fmt.Println("status command 'group' flag value: ", group)
			fmt.Println("status command 'all' flag value: ", all)
		}

		if group != "" {
			// display status of single group if a group name was provided
			a.DisplayGroup(group)
		} else if all {
			// display status of all groups if the all flag was provided
			a.DisplayAllGroups()
		} else {
			// otherwise, display command usage
			cmd.Usage()
		}

	},
}

func init() {
	rootCmd.AddCommand(statusCmd)

	// command flags and configuration settings.
	statusCmd.PersistentFlags().Bool("all", false, "Display Status for All Consistency Groups")
	statusCmd.PersistentFlags().String("group", "", "Display Status of Consistency Group by Name")
}
