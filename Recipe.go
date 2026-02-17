//go:build ignore

package main

import (
	"github.com/fezcode/gobake"
)

func main() {
	bake := gobake.NewEngine()
	bake.LoadRecipeInfo("recipe.piml")

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

		for _, t := range targets {
			output := "build/" + bake.Info.Name + "-" + t.os + "-" + t.arch
			if t.os == "windows" {
				output += ".exe"
			}
			
			// For atlas.cam, we try to build without CGO to allow cross-compilation where possible.
			// Note: Some camera drivers might require CGO (e.g. on macOS), so those builds might fail or have limited functionality.
			// Windows (MediaFoundation) often requires CGO with MinGW if not using pure Go impl.
			// Actually pion/mediadevices requires CGO for camera on Windows.
			if t.os == "windows" {
				// Assumes host has C compiler (MinGW)
				ctx.Env = []string{"CGO_ENABLED=1"}
			} else {
				ctx.Env = []string{"CGO_ENABLED=0"}
			}
			
			// We log but don't fail immediately on individual target errors to allow partial success
			err := ctx.BakeBinary(t.os, t.arch, output)
			if err != nil {
				ctx.Log("Warning: Failed to build for %s/%s: %v", t.os, t.arch, err)
			}
		}
		return nil
	})

	bake.Task("clean", "Removes build artifacts", func(ctx *gobake.Context) error {
		return ctx.Remove("build")
	})

	bake.Execute()
}
