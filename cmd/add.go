package cmd

import (
	"fmt"

	"github.com/HriThik-MaNoj/vhoster/internal/root"
	"github.com/HriThik-MaNoj/vhoster/internal/validate"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add <ip> <host> [host...]",
	Short: "Add one or more vhosts under an IP",
	Args:  cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		root.EnsureRoot(skipRoot())
		ip, hosts, errs := validate.ValidateAdd(args[0], args[1:])
		if len(errs) > 0 {
			return fmt.Errorf("validation failed:\n%s", validate.FormatErrors(errs))
		}
		ipCanonical := ip.String()
		if flagDryRun {
			fmt.Printf("[dry-run] would add under %s: %v\n", ipCanonical, hosts)
			return nil
		}
		if !confirm(fmt.Sprintf("Add %v under %s?", hosts, ipCanonical)) {
			fmt.Println("aborted")
			return nil
		}
		if err := manager().Add(ipCanonical, hosts); err != nil {
			return err
		}
		fmt.Printf("Added %v under %s\n", hosts, ipCanonical)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
}
