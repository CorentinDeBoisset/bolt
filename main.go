package main

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/corentindeboisset/tera/pkg/cfg"
	"github.com/corentindeboisset/tera/pkg/iface"
	"github.com/corentindeboisset/tera/pkg/jobexec"
	"github.com/corentindeboisset/tera/pkg/servicemgmt"
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
		Use:           "tera",
		Version:       Version,
		Short:         i18n.Sprintf("Boost your development workflow by at least 10^12."),
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	// Manually add the flags to add translation on the usage strings
	rootCmd.PersistentFlags().BoolP("help", "h", false, i18n.Sprintf("Display help information about the command"))
	rootCmd.Flags().BoolP("version", "v", false, i18n.Sprintf("Display version information"))

	rootCmd.SetUsageFunc(iface.RenderUsage)
	rootCmd.SetHelpFunc(iface.RenderHelp)

	versionCmd := &cobra.Command{
		Use:   "version",
		Short: i18n.Sprintf("Print version and exit"),
		Run: func(cmd *cobra.Command, args []string) {
			_, _ = i18n.Printf("tera version %s\n", Version)
		},
	}
	rootCmd.AddCommand(versionCmd)

	runCmd := &cobra.Command{
		Use:   "run [job-name] [-c config]",
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
	runCmd.Flags().StringVarP(&confPath, "config", "c", "", i18n.Sprintf("Path to a configuration file. If left empty, it will recursively search in the parent directories for a tera.yml file"))
	_ = runCmd.MarkFlagFilename("config", "yaml", "yml")

	rootCmd.AddCommand(runCmd)

	serviceCmd := &cobra.Command{
		Use:   "services [-c config]",
		Short: i18n.Sprintf("Start the service management interface"),
		RunE: func(cmd *cobra.Command, args []string) error {
			return servicemgmt.StartServiceManagement(confPath)
		},
	}
	serviceCmd.Flags().StringVarP(&confPath, "config", "c", "", i18n.Sprintf("Path to a configuration file. If left empty, it will recursively search in the parent directories for a tera.yml file"))
	_ = serviceCmd.MarkFlagFilename("config", "yaml", "yml")

	rootCmd.AddCommand(serviceCmd)

	// Add translations to the default commands
	rootCmd.InitDefaultHelpCmd()
	rootCmd.InitDefaultCompletionCmd()
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		iface.RenderError(err)
		os.Exit(1)
	}
}
