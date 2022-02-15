package main

import (
	"github.com/cmcpasserby/sxtp"
	"github.com/spf13/cobra"
	"log"
	"os"
)

const (
	version = "0.0.1"
)

func main() {
	var (
		formatFlag string
		suffixFlag string
	)

	rootCmd := &cobra.Command{
		Use:           "sxtp [atlasPath masksPath outputPath]",
		Version:       version,
		Short:         "Tool used for packing secondary textures in spine atlas format",
		Long:          "Tool used for packing secondary textures in spine atlas format",
		SilenceUsage:  false,
		SilenceErrors: true,
		Args:          cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			atlasPath, masksPath, outPath := args[0], args[1], args[2]

			f, err := os.Open(atlasPath)
			if err != nil {
				return err
			}
			defer f.Close()

			atlases, err := sxtp.DecodeAtlas(f)
			if err != nil {
				return err
			}

			return sxtp.PackMasks(atlases, sxtp.FileFormat(formatFlag), masksPath, outPath, suffixFlag, log.Default())
		},
	}

	rootCmd.Flags().StringVarP(
		&suffixFlag,
		"suffix",
		"s",
		"masks",
		"provides the suffix to use for the new atlas, useful when multiple atlases need to be packed for the same purpose",
	)
	rootCmd.Flags().StringVarP(&formatFlag, "format", "f", "png", "defines file format to save masks as options: [png, jpg]")

	if err := rootCmd.Execute(); err != nil {
		rootCmd.PrintErrln(err)
		os.Exit(1)
	}
}
