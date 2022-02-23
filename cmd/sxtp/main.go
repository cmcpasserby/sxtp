package main

import (
	"context"
	"flag"
	"github.com/cmcpasserby/sxtp"
	"github.com/peterbourgon/ff/v3/ffcli"
	"log"
	"os"
)

const (
	version = "0.0.1"
)

func main() {
	var (
		rootFlagSet = flag.NewFlagSet("sxtp", flag.ExitOnError)
		suffixFlag  = rootFlagSet.String("s", "masks", "provides the suffix to use for new atlas")
		formatFlag  = rootFlagSet.String("f", "png", "defines file format to saves masks as, options: [png, jpg]")
	)

	rootCmd := ffcli.Command{
		Name:       "sxtp",
		ShortUsage: "sxtp [flags] atlasPath masksPath outPath",
		ShortHelp:  "Tool used for packing secondary textures in spine atlas format",
		LongHelp:   "Tool used for packing secondary textures in spine atlas format",
		FlagSet:    rootFlagSet,
		Exec: func(ctx context.Context, args []string) error {
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

			return sxtp.PackMasks(atlases, sxtp.FileFormat(*formatFlag), maskPath, outPath, *suffixFlag, log.Default())
		},
	}

	if err := rootCmd.ParseAndRun(context.Background(), os.Args[1:]); err != nil {
		log.Fatal(err)
	}
}
