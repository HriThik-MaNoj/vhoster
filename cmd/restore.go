package cmd

import (
	"fmt"

	"github.com/HriThik-MaNoj/vhoster/internal/root"
	"github.com/spf13/cobra"
)

var restoreList bool

var restoreCmd = &cobra.Command{
	Use:   "restore [backup]",
	Short: "Restore /etc/hosts from a backup",
	RunE: func(cmd *cobra.Command, args []string) error {
		m := manager()

		if restoreList {
			paths, err := m.BackupList()
			if err != nil {
				return err
			}
			if len(paths) == 0 {
				fmt.Println("no backups")
				return nil
			}
			for _, p := range paths {
				fmt.Println(p)
			}
			return nil
		}

		root.EnsureRoot(skipRoot())

		if len(args) == 1 {
			if flagDryRun {
				fmt.Printf("[dry-run] would restore %s from %s\n", flagHosts, args[0])
				return nil
			}
			if !confirm(fmt.Sprintf("Restore %s from %s?", flagHosts, args[0])) {
				fmt.Println("aborted")
				return nil
			}
			if err := m.RestoreBackup(args[0]); err != nil {
				return err
			}
			fmt.Printf("Restored %s from %s\n", flagHosts, args[0])
			return nil
		}

		latest, err := m.LatestBackupPath()
		if err != nil {
			return err
		}
		if latest == "" {
			return fmt.Errorf("no backups found")
		}
		if flagDryRun {
			fmt.Printf("[dry-run] would restore %s from latest backup: %s\n", flagHosts, latest)
			return nil
		}
		if !confirm(fmt.Sprintf("Restore %s from latest backup (%s)?", flagHosts, latest)) {
			fmt.Println("aborted")
			return nil
		}
		if err := m.RestoreLatest(); err != nil {
			return err
		}
		fmt.Printf("Restored %s from %s\n", flagHosts, latest)
		return nil
	},
}

func init() {
	restoreCmd.Flags().BoolVar(&restoreList, "list", false, "list available backups")
	rootCmd.AddCommand(restoreCmd)
}
