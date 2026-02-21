package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/corentindeboisset/bolt/pkg/cfg"
	"github.com/corentindeboisset/bolt/pkg/iface"
	"github.com/corentindeboisset/bolt/pkg/jobexec"
	"github.com/corentindeboisset/bolt/pkg/servicemgmt"
)

//go:generate gotext -srclang=en update -out catalog.go -lang=fr,en

var (
	Version string
	rootCmd *cobra.Command

	confPath string
)

func init() {
	i18n := cfg.GetI18nPrinter()

	if Version == "" {
		Version = i18n.Sprintf("unknown (built from source)")
	}

	rootCmd = &cobra.Command{
		Use:     "bolt",
		Version: Version,
		Short:   i18n.Sprintf("Script management that rides the lighting."),
	}

	versionCmd := &cobra.Command{
		Use:   "version",
		Short: i18n.Sprintf("Print version and exit"),
		Run: func(cmd *cobra.Command, args []string) {
			_, _ = i18n.Printf("bolt version %s\n", Version)
		},
	}
	rootCmd.AddCommand(versionCmd)

	runCmd := &cobra.Command{
		Use:   "run [job-name]",
		Short: i18n.Sprintf("Run a job"),
		Args:  cobra.MaximumNArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return jobexec.GetJobList(confPath), cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			jobToRun := ""
			if len(args) >= 1 {
				jobToRun = args[0]
			}

			return jobexec.ExecuteJob(confPath, jobToRun)
		},
	}
	runCmd.Flags().StringVarP(&confPath, "config", "c", "", i18n.Sprintf("Path to a configuration file. If left empty, it will recursively search in the parent directories for a bolt.yml file"))
	_ = runCmd.MarkFlagFilename("config", "yaml", "yml")

	rootCmd.AddCommand(runCmd)

	serviceCmd := &cobra.Command{
		Use:   "services",
		Short: i18n.Sprintf("Start the service management interface"),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := servicemgmt.StartServiceManagement(confPath)
			if formattableErr, ok := errors.AsType[iface.FormattableError](err); ok {
				iface.PrintError(formattableErr)
				fmt.Printf("\n")
				os.Exit(1)
				return nil
			}

			return err
		},
	}
	serviceCmd.Flags().StringVarP(&confPath, "config", "c", "", i18n.Sprintf("Path to a configuration file. If left empty, it will recursively search in the parent directories for a bolt.yml file"))
	_ = serviceCmd.MarkFlagFilename("config", "yaml", "yml")

	rootCmd.AddCommand(serviceCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
