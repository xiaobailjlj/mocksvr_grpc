package version

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	// Version is the current version of the application
	Version = "0.1.0"
	// Commit is the git commit hash
	Commit = "unknown"
	// BuildDate is the date when the binary was built
	BuildDate = "unknown"
)

// NewVersionCmd creates a new version command
func NewVersionCmd() *cobra.Command {
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version information",
		Long:  `Print the version, commit hash, and build date of the application.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Version: %s\n", Version)
			fmt.Printf("Commit: %s\n", Commit)
			fmt.Printf("Build Date: %s\n", BuildDate)
		},
	}

	return versionCmd
}
