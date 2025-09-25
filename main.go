package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/corentindeboisset/bolt/pkg"
)

var (
	Version string
	rootCmd *cobra.Command

	confPath string
)

func init() {
	if Version == "" {
		Version = "unknown (built from source)"
	}

	rootCmd = &cobra.Command{
		Use:     "bolt",
		Version: Version,
		Short:   "bolt is a tool to execute complex jobs on your local machine",
	}

	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print version and exit",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("bolt version %s\n", Version)
		},
	}
	rootCmd.AddCommand(versionCmd)

	runCmd := &cobra.Command{
		Use:   "run [job-name]",
		Short: "Run a job",
		Args:  cobra.MaximumNArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return pkg.RunAutocomplete(confPath), cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			jobToRun := ""
			if len(args) >= 1 {
				jobToRun = args[0]
			}

			return pkg.RunCmd(confPath, jobToRun)
		},
	}
	runCmd.Flags().StringVarP(&confPath, "config", "c", "", "Path to a configuration file. If left empty, it will recursively search in the parent directories for a bolt.yml file")
	_ = runCmd.MarkFlagFilename("config", "yaml", "yml")

	rootCmd.AddCommand(runCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
