package main

import (
	"fmt"
	"strconv"
	"strings"
)

type ActionType int

const (
	BuildSite ActionType = iota
	DevServer
)

type BuildSiteOptions struct {
	buildDir    string
	buildDrafts bool
}

type DevServerOptions struct {
	port        int
	buildDrafts bool
}

type Action struct {
	typ     ActionType
	siteDir string

	buildSiteOpts BuildSiteOptions
	devServerOpts DevServerOptions
}

const (
	DefaultServerPort     = 4200
	DefaultBuildDirectory = "ssg-build"
)

// TODO: help menu

func sprintUsage() string {
	return "usage: ssg [dev / build] [directory]"
}

// ssg dev [directory] --port= -D --drafts
// ssg build [directory] --build-dir= -D --drafts
func parseArgs(args []string) (Action, error) {
	if len(args) < 2 {
		return Action{}, fmt.Errorf("%s", sprintUsage())
	}

	action := Action{}
	action.siteDir = args[1]

	var err error

	switch args[0] {
	case "build":
		action.typ = BuildSite

		if len(args) > 2 {
			action.buildSiteOpts, err = parseBuildSiteOptions(args[2:])
			if err != nil {
				return Action{}, fmt.Errorf("failed to parse options: %w", err)
			}
		} else {
			action.buildSiteOpts = defaultBuildSiteOptions()
		}

	case "dev":
		action.typ = DevServer

		if len(args) > 2 {
			action.devServerOpts, err = parseDevServerOptions(args[2:])
			if err != nil {
				return Action{}, fmt.Errorf("failed to parse options: %w", err)
			}
		} else {
			action.devServerOpts = defaultDevServerOpts()
		}

	default:
		return Action{}, fmt.Errorf(
			"invalid command %s. expected build or dev",
			args[0],
		)
	}

	return action, nil
}

func defaultBuildSiteOptions() BuildSiteOptions {
	return BuildSiteOptions{DefaultBuildDirectory, false}
}

func defaultDevServerOpts() DevServerOptions {
	return DevServerOptions{DefaultServerPort, false}
}

func parseBuildSiteOptions(opts []string) (BuildSiteOptions, error) {
	bso := defaultBuildSiteOptions()

	foundDraftOpt := false
	foundBuildDirOpt := false

	for _, opt := range opts {
		switch opt {
		case "-D", "--draft", "--draft=true":
			if foundDraftOpt {
				return BuildSiteOptions{}, fmt.Errorf(
					"multiple options given for --draft",
				)
			}
			bso.buildDrafts = true
			foundDraftOpt = true
		case "--draft=false":
			if foundDraftOpt {
				return BuildSiteOptions{}, fmt.Errorf(
					"multiple options given for --draft",
				)
			}
			foundDraftOpt = true
			bso.buildDrafts = false
		default:
			if strings.HasPrefix(opt, "--build-dir") {
				if foundBuildDirOpt {
					return BuildSiteOptions{}, fmt.Errorf(
						"multiple options given for --build-dir",
					)
				}

				if !strings.Contains(opt, "=") {
					return BuildSiteOptions{}, fmt.Errorf(
						"expected a directory for --build-dir. example --build-dir=ssg-build",
					)
				}

				foundBuildDirOpt = true
				bso.buildDir = strings.Split(opt, "=")[1]
				continue
			}
			return BuildSiteOptions{}, fmt.Errorf(
				"unrecognized option: %s",
				opt,
			)
		}
	}

	return bso, nil
}

func parseDevServerOptions(opts []string) (DevServerOptions, error) {
	dso := defaultDevServerOpts()

	foundDraftOpt := false
	foundPortOpt := false

	for _, opt := range opts {
		switch opt {
		case "-D", "--draft", "--draft=true":
			if foundDraftOpt {
				return DevServerOptions{}, fmt.Errorf(
					"multiple options given for --draft",
				)
			}
			dso.buildDrafts = true
			foundDraftOpt = true

		case "--draft=false":
			if foundDraftOpt {
				return DevServerOptions{}, fmt.Errorf(
					"multiple options given for --draft",
				)
			}
			foundDraftOpt = true
			dso.buildDrafts = false
		default:
			if strings.HasPrefix(opt, "--port") {
				if foundPortOpt {
					return DevServerOptions{}, fmt.Errorf(
						"multiple options given for --port",
					)
				}
				foundPortOpt = true

				if !strings.Contains(opt, "=") {
					return DevServerOptions{}, fmt.Errorf(
						"expected a number to be given for port. ex: --port=4200",
					)
				}

				num := strings.Split(opt, "=")[1]
				port, err := strconv.ParseInt(
					num,
					10,
					16,
				)
				if err != nil {
					return DevServerOptions{}, fmt.Errorf(
						"failed to parse number %s",
						num,
					)
				}
				dso.port = int(port)
				continue
			}

			return DevServerOptions{}, fmt.Errorf(
				"unrecognized option: %s",
				opt,
			)
		}
	}

	return dso, nil

}
