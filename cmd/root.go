// Package cmd implements the vhoster CLI (cobra) and dispatches the TUI.
package cmd

import (
	"bufio"
	"fmt"
	"os"

	"github.com/HriThik-MaNoj/vhoster/internal/hosts"
	"github.com/HriThik-MaNoj/vhoster/internal/root"
	"github.com/HriThik-MaNoj/vhoster/internal/tui"
	"github.com/spf13/cobra"
)

var (
	flagHosts     string
	flagBackupDir string
	flagYes       bool
	flagDryRun    bool
	flagForce     bool
)

const (
	defaultHosts     = "/etc/hosts"
	defaultBackupDir = "/var/lib/vhoster/backups"
)

var rootCmd = &cobra.Command{
	Use:   "vhoster",
	Short: "Manage /etc/hosts vhosts via TUI or CLI",
	Long: "vhoster manages virtual-host entries in /etc/hosts inside a managed block.\n" +
		"Run with sudo and no arguments for the TUI, or use a subcommand for scripting:\n" +
		"  vhoster add <ip> <host> [host...]\n" +
		"  vhoster remove <host>... [--ip <ip>]\n" +
		"  vhoster list\n" +
		"  vhoster restore [backup] [--list]",
	RunE: func(cmd *cobra.Command, args []string) error {
		root.EnsureRoot(skipRoot())
		return tui.Run(manager())
	},
}

func init() {
	rootCmd.PersistentFlags().StringVar(&flagHosts, "hosts", defaultHosts, "path to hosts file")
	rootCmd.PersistentFlags().StringVar(&flagBackupDir, "backup-dir", defaultBackupDir, "backup directory")
	rootCmd.PersistentFlags().BoolVarP(&flagYes, "yes", "y", false, "skip confirmation prompts")
	rootCmd.PersistentFlags().BoolVar(&flagDryRun, "dry-run", false, "show changes without writing")
	rootCmd.PersistentFlags().BoolVar(&flagForce, "force", false, "override safety checks where possible")
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func manager() *hosts.Manager {
	return &hosts.Manager{HostsPath: flagHosts, BackupDir: flagBackupDir, Force: flagForce}
}

// skipRoot reports whether the root requirement is waived (a non-default
// --hosts path is in use, e.g. for local testing).
func skipRoot() bool {
	return flagHosts != defaultHosts
}

// confirm prompts the user unless --yes is set. Returns true on an
// affirmative reply (accepts y, Y, yes, YES, etc.).
func confirm(prompt string) bool {
	if flagYes {
		return true
	}
	fmt.Printf("%s [y/N]: ", prompt)
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	resp := scanner.Text()
	return len(resp) > 0 && (resp[0] == 'y' || resp[0] == 'Y')
}
