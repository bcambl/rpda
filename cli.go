package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

func validValue(value string, list []string) bool {
	for _, t := range list {
		tp := strings.ToLower(value)
		if t == tp {
			return true
		}
	}
	return false
}

func main() {

	a := &App{}
	a.readConfig("settings.json")

	validTasks := []string{"display", "enable", "disable"}
	taskPtr := flag.String("task", "",
		"Task to perform. Valid Options: "+strings.Join(validTasks, ", "))

	groupPtr := flag.String("group", "",
		"Consisntency group which to perform operation")

	validCopies := []string{"dr", "test"}
	copyPtr := flag.String("copy", "",
		"Copy to enable. Valid Options: "+strings.Join(validCopies, ", "))

	allPtr := flag.Bool("all", false,
		"Operate on all consistency groups which user has permissions")

	delayPtr := flag.Int("delay", 3,
		"Delay to wait when operating on multiple groups")

	debugPtr := flag.Bool("debug", false,
		"Enable Debug Mode")

	examplesPtr := flag.Bool("examples", false, "Show CLI Examples")

	flag.Parse()

	// Validate Task
	var valid bool
	valid = validValue(*taskPtr, validTasks)
	if !valid {
		fmt.Printf("\nPlease provide a valid Task (See: -help).\nOptions: %s\n\n",
			strings.Join(validTasks, ", "))
		os.Exit(1)
	}
	a.Task = *taskPtr

	// Validate Copy
	valid = validValue(*copyPtr, validCopies)
	// copy is not required to perform a display
	if !valid && a.Task != "display" {
		fmt.Printf("\nPlease provide a valid Copy. "+
			"(See: -help)\nOptions: %s\n\n",
			strings.Join(validCopies, ", "))
		os.Exit(1)
	}
	a.Copy = *copyPtr

	// If All and Group were passed. Default to single group
	if *allPtr {
		a.All = *allPtr
		if *groupPtr != "" {
			a.All = false
		}
	}

	if *groupPtr == "" {
		if !a.All {
			fmt.Printf("\nPlease provide a Consistency Group Name. " +
				"(See: -help)\n")
			os.Exit(1)
		}
	}
	a.Group = *groupPtr

	a.Delay = *delayPtr
	a.Debug = *debugPtr

	if a.Debug {
		a.debug()
	}

	if *examplesPtr {
		a.usageExamples()
		os.Exit(0)
	}

	a.Run()
}
