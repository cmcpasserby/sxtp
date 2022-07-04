package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/cmcpasserby/sxtp"
	"github.com/peterbourgon/ff/v3/ffcli"
	"log"
	"os"
)

const (
	version = "1.0.0" // TODO make command to print version
)

func main() {
	var (
		rootFlagSet      = flag.NewFlagSet("sxtp", flag.ExitOnError)
		versionFlag      = rootFlagSet.Bool("v", false, "prints sxtp's version")
		suffixFlag       = rootFlagSet.String("s", "masks", "provides the suffix to use for new atlas")
		formatFlag       = rootFlagSet.String("f", "png", "defines file format to saves masks as options: [png, jpg]")
		includeAlphaFlag = rootFlagSet.Bool("a", false, "should alpha channel be included in packed secondary texture")
	)

	rootCmd := ffcli.Command{
		Name:       "sxtp",
		ShortUsage: "sxtp [flags] atlasPath masksPath outPath",
		ShortHelp:  "Tool used for packing secondary textures in spine atlas format",
		LongHelp:   "Tool used for packing secondary textures in spine atlas format", // TODO long help
		FlagSet:    rootFlagSet,
		Exec: func(ctx context.Context, args []string) error {
			if *versionFlag {
				fmt.Printf("sxtp version %s\n", version)
				return nil
			}

			atlasPath, maskPath, outPath := args[0], args[1], args[2]

			f, err := os.Open(atlasPath)
			if err != nil {
				return err
			}
			defer f.Close()

			atlases, err := sxtp.DecodeAtlas(f)
			if err != nil {
				return err
			}

			return sxtp.PackMasks(atlases, sxtp.FileFormat(*formatFlag), maskPath, outPath, *suffixFlag, *includeAlphaFlag, log.Default())
		},
	}

	if err := rootCmd.ParseAndRun(context.Background(), os.Args[1:]); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
	}
}
