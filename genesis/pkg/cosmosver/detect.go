package cosmosver

import (
	"github.com/interchained/genesis/genesis/pkg/gomodule"
	"golang.org/x/mod/semver"
)

const (
	cosmosModulePath            = "github.com/interchained/cosmos-sdk"
	cosmosModuleMaxLaunchpadTag = "v0.44.1"
	cosmosModuleStargateTag     = "v0.43.1"
)

// Detect dedects major version of Cosmos.
func Detect(appPath string) (Version, error) {
	parsed, err := gomodule.ParseAt(appPath)
	if err != nil {
		return 0, err
	}
	for _, r := range parsed.Require {
		v := r.Mod
		if v.Path == cosmosModulePath {
			switch {
			case semver.Compare(v.Version, cosmosModuleStargateTag) >= 0:
				return StargateZeroFourtyAndAbove, nil

			case semver.Compare(v.Version, cosmosModuleMaxLaunchpadTag) <= 0:
				return LaunchpadAny, nil

			default:
				return StargateBelowZeroFourty, nil
			}
		}
	}
	return 0, nil
}
