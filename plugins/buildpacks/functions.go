package buildpacks

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

func validBuildpackURL(buildpack string) (string, error) {
	if buildpack == "" {
		return buildpack, errors.New("Must specify a buildpack to add")
	}

	reHerokuValue := regexp.MustCompile(`(?m)^([\w]+\/[\w]+)$`)
	if found := reHerokuValue.Find([]byte(buildpack)); found != nil {
		parts := strings.SplitN(buildpack, "/", 2)
		return fmt.Sprintf("https://github.com/%s/heroku-buildpack-%s.git", parts[0], parts[1]), nil
	}

	reString := regexp.MustCompile(`(?m)^(http|https|git)(:\/\/|@)([^\/:]+)[\/:]([^\/:]+)\/(.+)(.git(#derp)?)?$`)
	if found := reString.Find([]byte(buildpack)); found != nil {
		return buildpack, nil
	}

	return buildpack, fmt.Errorf("Invalid buildpack specified: %v", buildpack)
}
