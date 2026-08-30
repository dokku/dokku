package scheduler_k3s

import (
	"errors"
	"fmt"
	"regexp"

	"helm.sh/helm/v3/pkg/chartutil"
)

// helmMaxReleaseNameLength mirrors the release name cap chartutil.ValidateReleaseName
// enforces, which helm does not export
const helmMaxReleaseNameLength = 53

// maxNodeProfileNameLength is the longest node profile name that still derives a helm
// release name for its node sysctls chart that helm will accept
const maxNodeProfileNameLength = helmMaxReleaseNameLength - len(nodeSysctlsProfileReleasePrefix)

// legacyMaxNodeProfileNameLength is the length limit node profiles were created under
// before their names were validated against the helm release name they derive
const legacyMaxNodeProfileNameLength = 32

// nodeProfileNamePattern matches the node profile names dokku can derive every downstream
// name from. It is lowercase-only because a helm release name is, and dokku appends a
// profile name to a release name prefix verbatim.
var nodeProfileNamePattern = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]*[a-z0-9])?$`)

// legacyNodeProfileNamePattern matches the node profile names dokku accepted before it
// validated them against the helm release name they derive
var legacyNodeProfileNamePattern = regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])?$`)

// validateNodeProfileName rejects node profile names dokku cannot derive a valid helm
// release name from. The node sysctls release is the tightest constraint a profile name
// feeds, so the name is checked against that rather than on its own: a name accepted here
// but rejected by helm would create a profile that only fails later, on commands with no
// obvious relationship to it.
func validateNodeProfileName(profileName string) error {
	if profileName == "" {
		return errors.New("Missing profile name")
	}

	if !nodeProfileNamePattern.MatchString(profileName) {
		return fmt.Errorf("Invalid profile name, must only contain lowercase alphanumeric characters and dashes and cannot start or end with a dash: %s", profileName)
	}

	if len(profileName) > maxNodeProfileNameLength {
		return fmt.Errorf("Profile name is too long, must be at most %d characters: %s", maxNodeProfileNameLength, profileName)
	}

	if err := chartutil.ValidateReleaseName(getNodeSysctlsReleaseName(profileName)); err != nil {
		return fmt.Errorf("Invalid profile name %s, unable to derive a node sysctls release name from it: %w", profileName, err)
	}

	return nil
}

// validateStoredNodeProfileName checks a node profile name against the rules it may have
// been stored under. It stays looser than validateNodeProfileName so a profile created
// before names were validated against their helm release name is still removable by the
// command that created it.
func validateStoredNodeProfileName(profileName string) error {
	if profileName == "" {
		return errors.New("Missing profile name")
	}

	if !legacyNodeProfileNamePattern.MatchString(profileName) {
		return fmt.Errorf("Invalid profile name, must only contain alphanumeric characters and dashes and cannot start or end with a dash: %s", profileName)
	}

	if len(profileName) > legacyMaxNodeProfileNameLength {
		return fmt.Errorf("Profile name is too long, must be at most %d characters: %s", legacyMaxNodeProfileNameLength, profileName)
	}

	return nil
}
