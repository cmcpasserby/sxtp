package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/cmcpasserby/scli"
	"github.com/cmcpasserby/sxtp"
	"log"
	"os"
)

const (
	version = "1.0.0"
)

func main() {
	var (
		rootFlagSet      = flag.NewFlagSet("sxtp", flag.ExitOnError)
		versionFlag      = rootFlagSet.Bool("v", false, "prints sxtp's version")
		suffixFlag       = rootFlagSet.String("s", "masks", "provides the suffix to use for new atlas")
		formatFlag       = rootFlagSet.String("f", "png", "defines file format to saves masks as options: [png, jpg]")
		includeAlphaFlag = rootFlagSet.Bool("a", false, "should alpha channel be included in packed secondary texture")
	)

	cmd := &scli.Command{
		Usage:     "sxtp [flags] <atlasPath masksPath> [outPath]",
		ShortHelp: "Tool used for packing secondary textures in spine atlas format",
		LongHelp:  "Tool used for packing secondary textures in spine atlas format", // TODO long help
		FlagSet:   rootFlagSet,
		Exec: func(ctx context.Context, args []string) error {
			if *versionFlag {
				fmt.Printf("sxtp version %s\n", version)
				return nil
			}

			var (
				err                          error
				atlasPath, maskPath, outPath string
			)

			switch len(args) {
			case 3:
				atlasPath, maskPath, outPath = args[0], args[1], args[2]

			case 2:
				atlasPath, maskPath = args[0], args[1]
				outPath, err = os.Getwd()
				if err != nil {
					return err
				}

			default:
				return fmt.Errorf("expected 2 to 3 args got %d", len(args))
			}

			f, err := os.Open(atlasPath)
			if err != nil {
				return err
			}
			defer f.Close()

			atlases, err := sxtp.DecodeAtlas(f)
			if err != nil {
				return err
			}

			l := log.New(os.Stdout, "", 0)

			return sxtp.PackMasks(atlases, sxtp.FileFormat(*formatFlag), maskPath, outPath, *suffixFlag, *includeAlphaFlag, l)
		},
	}

	if err := cmd.ParseAndRun(context.Background(), os.Args[1:]); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
	}
}
