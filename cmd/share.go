package cmd

import (
	"github.com/aesadde/easyshare/internal/service"
	"github.com/spf13/cobra"
)

// shareCmd represents the share command
var shareCmd = &cobra.Command{
	Use:   "share",
	Short: "Share your posts to webflow and other tools",
	Run: func(cmd *cobra.Command, args []string) {

		token, err := cmd.Flags().GetString("webflow-api-token")
		cobra.CheckErr(err)

		collection, err := cmd.Flags().GetString("webflow-collection-id")

		template, err := cmd.Flags().GetString("template")
		cobra.CheckErr(err)

		rPath, err := cmd.Flags().GetString("resource-path")
		cobra.CheckErr(err)

		svc := service.NewEasyShare(token, collection, rPath)

		err = svc.NewPost(args[0], template)
		cobra.CheckErr(err)
	},
}

func init() {

	rootCmd.AddCommand(shareCmd)
}
