package cmd

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List managed vhost entries (read-only)",
	RunE: func(cmd *cobra.Command, args []string) error {
		entries, err := manager().List()
		if err != nil {
			return err
		}
		if len(entries) == 0 {
			fmt.Println("no managed entries")
			return nil
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "IP\tHOSTS")
		for _, e := range entries {
			fmt.Fprintf(w, "%s\t%s\n", e.IP, strings.Join(e.Hosts, ", "))
		}
		return w.Flush()
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
