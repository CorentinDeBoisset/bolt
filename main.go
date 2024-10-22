package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/corentindeboisset/localci/pkg"
)

var (
	Version string
	rootCmd *cobra.Command

	confPath     string
	selectedStep string
)

func init() {
	if Version == "" {
		Version = "unknown (built from source)"
	}

	rootCmd = &cobra.Command{
		Use:     "localci",
		Version: Version,
		Short:   "localci is a tool to execute complex CI jobs on your local machine",
	}

	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print version and exit",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("localci version %s\n", Version)
		},
	}
	rootCmd.AddCommand(versionCmd)

	runCmd := &cobra.Command{
		Use:   "run",
		Short: "Run the local ci",
		RunE: func(cmd *cobra.Command, args []string) error {
			return pkg.RunCmd(confPath, selectedStep)
		},
	}
	runCmd.Flags().StringVarP(&confPath, "config", "c", "", "Path to a configuration file. If left empty, it will recursively search in the parent directories for a localci.yml file")
	runCmd.Flags().StringVarP(&selectedStep, "limit", "l", "", "Limit the execution to a specified step. Can be in the format \"step_foo:job_bar\" to run a specific job from a given step")
	rootCmd.AddCommand(runCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
