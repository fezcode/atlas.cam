//go:build gobake
package bake_recipe

import (
	"fmt"
	"runtime"

	"github.com/fezcode/gobake"
)

func Run(bake *gobake.Engine) error {
	if err := bake.LoadRecipeInfo("recipe.piml"); err != nil {
		return err
	}

	// Helper function to build a specific target
	buildTarget := func(ctx *gobake.Context, osName, arch string) error {
		output := fmt.Sprintf("build/%s-%s-%s", bake.Info.Name, osName, arch)
		if osName == "windows" {
			output += ".exe"
		}

		// Configure CGO
		// Windows: Requires CGO for MediaFoundation (needs MinGW on host)
		if osName == "windows" {
			ctx.Env = []string{"CGO_ENABLED=1"}
		} else if osName == "darwin" {
			// macOS: Requires CGO for AVFoundation
			// Only enable if we are actually building ON macOS
			if runtime.GOOS == "darwin" {
				ctx.Env = []string{"CGO_ENABLED=1"}
			} else {
				ctx.Env = []string{"CGO_ENABLED=0"}
			}
		} else {
			// Linux/Other: Default to 0 for static binary portability
			ctx.Env = []string{"CGO_ENABLED=0"}
		}

		ctx.Log("Baking %s/%s -> %s", osName, arch, output)
		if err := ctx.BakeBinary(osName, arch, output); err != nil {
			ctx.Log("Warning: Failed to build for %s/%s: %v", osName, arch, err)
			// We return nil to allow other targets in a batch to proceed,
			// but we logged the warning.
			return nil
		}
		return nil
	}

	bake.Task("build:linux", "Builds for Linux (amd64, arm64)", func(ctx *gobake.Context) error {
		ctx.Mkdir("build")
		buildTarget(ctx, "linux", "amd64")
		buildTarget(ctx, "linux", "arm64")
		return nil
	})

	bake.Task("build:windows", "Builds for Windows (amd64, arm64)", func(ctx *gobake.Context) error {
		ctx.Mkdir("build")
		buildTarget(ctx, "windows", "amd64")
		buildTarget(ctx, "windows", "arm64")
		return nil
	})

	bake.Task("build:darwin", "Builds for macOS (amd64, arm64)", func(ctx *gobake.Context) error {
		ctx.Mkdir("build")
		buildTarget(ctx, "darwin", "amd64")
		buildTarget(ctx, "darwin", "arm64")
		return nil
	})

	bake.Task("build", "Builds for all supported platforms", func(ctx *gobake.Context) error {
		ctx.Log("Building %s v%s for ALL platforms...", bake.Info.Name, bake.Info.Version)
		ctx.Mkdir("build")

		// Run all build steps
		buildTarget(ctx, "linux", "amd64")
		buildTarget(ctx, "linux", "arm64")
		buildTarget(ctx, "windows", "amd64")
		buildTarget(ctx, "windows", "arm64")
		buildTarget(ctx, "darwin", "amd64")
		buildTarget(ctx, "darwin", "arm64")

		return nil
	})

	bake.Task("clean", "Removes build artifacts", func(ctx *gobake.Context) error {
		return ctx.Remove("build")
	})

	return nil
}
