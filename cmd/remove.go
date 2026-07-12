package cmd

import (
	"fmt"

	"github.com/HriThik-MaNoj/vhoster/internal/root"
	"github.com/HriThik-MaNoj/vhoster/internal/validate"
	"github.com/spf13/cobra"
)

var removeIP string

var removeCmd = &cobra.Command{
	Use:   "remove <host> [host...] [--ip <ip>]",
	Short: "Remove vhosts from the managed block",
	RunE: func(cmd *cobra.Command, args []string) error {
		root.EnsureRoot(skipRoot())
		m := manager()

		if removeIP != "" {
			ip, err := validate.ValidateIP(removeIP)
			if err != nil {
				return err
			}
			if flagDryRun {
				fmt.Printf("[dry-run] would remove all hosts under %s\n", ip.String())
				return nil
			}
			if !confirm(fmt.Sprintf("Remove all hosts under %s?", ip.String())) {
				fmt.Println("aborted")
				return nil
			}
			if err := m.Remove(nil, ip.String()); err != nil {
				return err
			}
			fmt.Printf("Removed all hosts under %s\n", ip.String())
			return nil
		}

		if len(args) == 0 {
			return fmt.Errorf("provide a hostname to remove, or use --ip <ip>")
		}
		var errs []string
		for _, h := range args {
			if err := validate.ValidateHostname(h); err != nil {
				errs = append(errs, err.Error())
			}
		}
		if len(errs) > 0 {
			return fmt.Errorf("validation failed:\n%s", validate.FormatErrors(errs))
		}
		if flagDryRun {
			fmt.Printf("[dry-run] would remove hosts: %v\n", args)
			return nil
		}
		if !confirm(fmt.Sprintf("Remove %v?", args)) {
			fmt.Println("aborted")
			return nil
		}
		if err := m.Remove(args, ""); err != nil {
			return err
		}
		fmt.Printf("Removed %v\n", args)
		return nil
	},
}

func init() {
	removeCmd.Flags().StringVar(&removeIP, "ip", "", "remove all hosts for this IP")
	rootCmd.AddCommand(removeCmd)
}
