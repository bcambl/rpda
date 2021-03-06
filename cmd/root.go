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
	"syscall"

	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh/terminal"

	"github.com/spf13/cobra"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var (
	cfgFile       string
	userName      string
	promptPass    bool
	debugFlag     bool
	checkFlag     bool
	delayFlag     int
	pollDelayFlag int
	pollMaxFlag   int
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "rpda",
	Short: "RecoverPoint Direct Access",
	Long: `RecoverPoint Direct Access
==========================

examples:

# list all CG's by name
rpda list
 
# display replication status for all CG's
rpda status --all
 
# display replication status for a single CG
rpda status --group My_CG
 
# enable direct image access mode on latest test copy for ALL CG's
rpda enable --all --test
 
# enable direct image access mode on latest test copy for single CG
rpda enable --group My_CG --test
 
# enable direct image access mode on latest "DR" copy for single CG
rpda enable --group My_CG --dr
 
# enable direct image access mode on latest copy by name for single CG
rpda enable --group My_CG --copy <desired copy name>
 
# disable direct image access mode for ALL CG's (all copies)
rpda finish --all
 
# disable direct image access mode for a single CG (all copies)
rpda finish --group My_CG

    `,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global application flags
	rootCmd.PersistentFlags().StringVar(&userName, "user", "", "specify username")
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.rpda.yaml)")
	rootCmd.PersistentFlags().BoolVar(&checkFlag, "check", false, "enable check mode (no changes will be made)")
	rootCmd.PersistentFlags().BoolVar(&debugFlag, "debug", false, "enable debug mode")
	rootCmd.PersistentFlags().IntVar(&delayFlag, "delay", 0, "Seconds to wait between Consistency Groups with --all")
	rootCmd.PersistentFlags().IntVar(&pollDelayFlag, "polldelay", 3, "Seconds to wait between API status polling requests")
	rootCmd.PersistentFlags().IntVar(&pollMaxFlag, "pollmax", 30, "Number of status poll attempts with before failing")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	// skip config setup when caller simply wants to display the license or version
	if os.Args[1] == "license" || os.Args[1] == "version" {
		return
	}

	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Set path(s) to search for configuration file
		viper.AddConfigPath(home)
		viper.AddConfigPath(".")

		// Set default config name to search for (without extension)
		viper.SetConfigName(".rpda")

		// Set environment variable prefixes
		viper.SetEnvPrefix("RPDA")
	}

	viper.AutomaticEnv() // read in environment variables that match

	defaultURL := "https://recoverpoint_fqdn/"
	defaultUsername := "username"

	if debugFlag {
		log.SetLevel(log.DebugLevel)
	}

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {

			// Define placeholder configuration values for API
			viper.Set("api.url", defaultURL)
			viper.Set("api.username", defaultUsername)
			viper.Set("api.delay", 0)
			viper.Set("api.polldelay", 3)
			viper.Set("api.pollmax", 30)
			// Define placeholder copy identifiers
			viper.Set("identifiers.production_node_regexp", "_PN$")
			viper.Set("identifiers.copy_node_regexp", "_CN$")
			viper.Set("identifiers.test_node_regexp", "^TC_")

			home, err := homedir.Dir()
			if err != nil {
				log.Fatal(err)
			}
			newConfig := fmt.Sprintf(home + "/.rpda.yaml")
			err = viper.WriteConfigAs(newConfig)
			if err != nil {
				log.Fatal(err)
			}
			log.Info("New configuration created. Please Update: ", newConfig)
			os.Exit(0)
		} else {
			log.Fatal(err)
		}
	}

	fmt.Println("Using config file: ", viper.ConfigFileUsed())

	// add check and debug flags to viper
	viper.Set("check", checkFlag)
	viper.Set("debug", debugFlag)
	viper.Set("api.delay", delayFlag)
	viper.Set("api.polldelay", pollDelayFlag)
	viper.Set("api.pollmax", pollDelayFlag)

	// test for default url & username
	if viper.Get("api.url") == defaultURL || viper.Get("api.username") == defaultUsername {
		log.Fatal("Sample configuration detected. Please Update: ", viper.ConfigFileUsed())
	}

	// override username if provided
	if userName != "" {
		viper.Set("api.username", userName)
		// force a password prompt
		promptPass = true
	}

	// prompt for password if not saved
	passwordPrompt()
}

func passwordPrompt() {
	// password _can_ be saved to the config file; however, prompt by default.
	// consider this a hidden feature as passwords should not be stored in in plain text.
	if viper.Get("api.password") == nil || promptPass == true {
		fmt.Printf("provide password for user '%s' : ", viper.Get("api.username"))

		p, err := terminal.ReadPassword(int(syscall.Stdin))
		if err != nil {
			log.Fatal(err)
		}

		viper.Set("api.password", p)

		fmt.Println("")
	}
}
