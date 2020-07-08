package rpa

import (
	"regexp"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// Load will populate the application config based on the configuration file & cli flags
func (c *Config) Load() *Config {
	c.RPAURL = viper.GetString("api.url")
	c.Username = viper.GetString("api.username")
	c.Password = viper.GetString("api.password")
	c.Delay = viper.GetInt("api.delay")
	c.PollDelay = viper.GetInt("api.polldelay")
	c.PollMax = viper.GetInt("api.pollmax")
	c.CheckMode = viper.GetBool("check")
	c.Debug = viper.GetBool("debug")

	log.WithFields(log.Fields{
		"RPAURL":    c.RPAURL,
		"Username":  c.Username,
		"Password":  "REDACTED",
		"Delay":     c.Delay,
		"PollDelay": c.PollDelay,
		"PollMax":   c.PollMax,
		"CheckMode": c.CheckMode,
		"Debug":     c.Debug,
	}).Debug("Config struct variable assignments")

	return c
}

// Load will compile the application copy regular expression identifiers
func (i *Identifiers) Load() *Identifiers {
	i.ProductionNodeRegexp = regexp.MustCompile(viper.GetString("identifiers.production_node_regexp"))
	i.CopyNodeRegexp = regexp.MustCompile(viper.GetString("identifiers.copy_node_regexp"))
	i.TestNodeRegexp = regexp.MustCompile(viper.GetString("identifiers.test_node_regexp"))

	log.WithFields(log.Fields{
		"ProductionNodeRegexp": i.ProductionNodeRegexp.String(),
		"CopyNodeRegexp":       i.CopyNodeRegexp.String(),
		"TestNodeRegexp":       i.TestNodeRegexp.String(),
	}).Debug("Identifiers struct variable assignments")

	return i
}
