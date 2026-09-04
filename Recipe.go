//go:build gobake

package bake_recipe

import (
	"fmt"

	"github.com/fezcode/gobake"
)

func Run(bake *gobake.Engine) error {
	if err := bake.LoadRecipeInfo("recipe.piml"); err != nil {
		return err
	}

	bake.Task("build", "Builds the binary for multiple platforms", func(ctx *gobake.Context) error {
		ctx.Log("Building %s v%s...", bake.Info.Name, bake.Info.Version)

		targets := []struct {
			os   string
			arch string
		}{
			{"linux", "amd64"},
			{"linux", "arm64"},
			{"windows", "amd64"},
			{"windows", "arm64"},
			{"darwin", "amd64"},
			{"darwin", "arm64"},
		}

		err := ctx.Mkdir("build")
		if err != nil {
			return err
		}

		ldflags := fmt.Sprintf("-X main.Version=%s", bake.Info.Version)

		for _, t := range targets {
			// The camera backend on macOS is AVFoundation, which needs cgo, so this target cannot be cross-compiled with
			// CGO disabled. Build it on a Mac machine instead of failing here.
			if t.os == "darwin" {
				ctx.Log("Skipping %s/%s: AVFoundation needs cgo (build on Mac)", t.os, t.arch)
				continue
			}

			output := "build/" + bake.Info.Name + "-" + t.os + "-" + t.arch
			if t.os == "windows" {
				output += ".exe"
			}

			ctx.Env = []string{
				"CGO_ENABLED=0",
				"GOOS=" + t.os,
				"GOARCH=" + t.arch,
			}

			// We use manual go build to inject ldflags
			err := ctx.Run("go", "build", "-ldflags", ldflags, "-o", output, ".")
			if err != nil {
				return err
			}
		}
		return nil
	})

	bake.Task("clean", "Removes build artifacts", func(ctx *gobake.Context) error {
		return ctx.Remove("build")
	})

	return nil
}
