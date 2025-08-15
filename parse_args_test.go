package main

import (
	"fmt"
	"strings"
	"testing"
)

func TestParseArgs(t *testing.T) {
	tests := []struct {
		args           string
		expectedAction Action
		expectedErr    error
	}{
		{
			"build site",
			Action{
				typ:           BuildSite,
				siteDir:       "site",
				buildSiteOpts: defaultBuildSiteOptions(),
			},
			nil,
		},
		{
			"build",
			Action{},
			fmt.Errorf("usage: ssg [dev / build] [directory]"),
		},
		{
			"build site --draft",
			Action{
				typ:           BuildSite,
				siteDir:       "site",
				buildSiteOpts: BuildSiteOptions{DefaultBuildDirectory, true},
			},
			nil,
		},
		{
			"build site --draft=true",
			Action{
				typ:           BuildSite,
				siteDir:       "site",
				buildSiteOpts: BuildSiteOptions{DefaultBuildDirectory, true},
			},
			nil,
		},
		{
			"build site --draft=false",
			Action{
				typ:           BuildSite,
				siteDir:       "site",
				buildSiteOpts: BuildSiteOptions{DefaultBuildDirectory, false},
			},
			nil,
		},
		{
			"build site -D",
			Action{
				typ:           BuildSite,
				siteDir:       "site",
				buildSiteOpts: BuildSiteOptions{DefaultBuildDirectory, true},
			},
			nil,
		},
		{
			"build site --build-dir=out",
			Action{
				typ:           BuildSite,
				siteDir:       "site",
				buildSiteOpts: BuildSiteOptions{"out", false},
			},
			nil,
		},
		{
			"build site --build-dir=build",
			Action{
				typ:           BuildSite,
				siteDir:       "site",
				buildSiteOpts: BuildSiteOptions{"build", false},
			},
			nil,
		},
		{
			"build site --build-dir=out -D",
			Action{
				typ:           BuildSite,
				siteDir:       "site",
				buildSiteOpts: BuildSiteOptions{"out", true},
			},
			nil,
		},
		{
			"build site --build-dir=out --draft",
			Action{
				typ:           BuildSite,
				siteDir:       "site",
				buildSiteOpts: BuildSiteOptions{"out", true},
			},
			nil,
		},
		{
			"build site --build-dir=out --draft=false",
			Action{
				typ:           BuildSite,
				siteDir:       "site",
				buildSiteOpts: BuildSiteOptions{"out", false},
			},
			nil,
		},
		{
			"build site --draft=true --build-dir=out",
			Action{
				typ:           BuildSite,
				siteDir:       "site",
				buildSiteOpts: BuildSiteOptions{"out", true},
			},
			nil,
		},
		{
			"build site -D --build-dir=out",
			Action{
				typ:           BuildSite,
				siteDir:       "site",
				buildSiteOpts: BuildSiteOptions{"out", true},
			},
			nil,
		},
		{
			"build site -D -D",
			Action{},
			fmt.Errorf(
				"failed to parse options: multiple options given for --draft",
			),
		},
		{
			"build site --draft -D",
			Action{},
			fmt.Errorf(
				"failed to parse options: multiple options given for --draft",
			),
		},
		{
			"build site --draft --draft",
			Action{},
			fmt.Errorf(
				"failed to parse options: multiple options given for --draft",
			),
		},
		{
			"build site --draft=false --draft",
			Action{},
			fmt.Errorf(
				"failed to parse options: multiple options given for --draft",
			),
		},
		{
			"build site --draft=false -D",
			Action{},
			fmt.Errorf(
				"failed to parse options: multiple options given for --draft",
			),
		},
		{
			"build site --draft=false --draft=true",
			Action{},
			fmt.Errorf(
				"failed to parse options: multiple options given for --draft",
			),
		},
		{
			"build site --bad=true --draft",
			Action{},
			fmt.Errorf(
				"failed to parse options: unrecognized option: --bad=true",
			),
		},
		{
			"build site --draft --draft --bad=true",
			Action{},
			fmt.Errorf(
				"failed to parse options: multiple options given for --draft",
			),
		},
		{
			"build site --build-dir",
			Action{},
			fmt.Errorf(
				"failed to parse options: expected a directory for --build-dir. example --build-dir=ssg-build",
			),
		},
		{
			"build site --build-dir=ssg-build --build-dir=bad",
			Action{},
			fmt.Errorf(
				"failed to parse options: multiple options given for --build-dir",
			),
		},
		{
			"build site --build-dir=ssg-build --draft --build-dir=bad",
			Action{},
			fmt.Errorf(
				"failed to parse options: multiple options given for --build-dir",
			),
		},
		{
			"build site --draft --build-dir=ssg-build --draft --build-dir=bad",
			Action{},
			fmt.Errorf(
				"failed to parse options: multiple options given for --draft",
			),
		},

		{
			"dne oops",
			Action{},
			fmt.Errorf("invalid command dne. expected build or dev"),
		},

		{
			"dev site",
			Action{
				typ:           DevServer,
				siteDir:       "site",
				devServerOpts: defaultDevServerOpts(),
			},
			nil,
		},
		{
			"dev",
			Action{},
			fmt.Errorf("usage: ssg [dev / build] [directory]"),
		},
		{
			"dev site --port=80",
			Action{
				typ:           DevServer,
				siteDir:       "site",
				devServerOpts: DevServerOptions{80, false},
			},
			nil,
		},
		{
			"dev site --port=8080",
			Action{
				typ:           DevServer,
				siteDir:       "site",
				devServerOpts: DevServerOptions{8080, false},
			},
			nil,
		},
		{
			"dev site --port=1000000",
			Action{},
			fmt.Errorf(
				"failed to parse options: failed to parse number 1000000",
			),
		},
		{
			"dev site --bad=true",
			Action{},
			fmt.Errorf(
				"failed to parse options: unrecognized option: --bad=true",
			),
		},
		{
			"dev site --draft",
			Action{
				typ:           DevServer,
				siteDir:       "site",
				devServerOpts: DevServerOptions{DefaultServerPort, true},
			},
			nil,
		},
		{
			"dev site --draft=true",
			Action{
				typ:           DevServer,
				siteDir:       "site",
				devServerOpts: DevServerOptions{DefaultServerPort, true},
			},
			nil,
		},
		{
			"dev site -D",
			Action{
				typ:           DevServer,
				siteDir:       "site",
				devServerOpts: DevServerOptions{DefaultServerPort, true},
			},
			nil,
		},
		{
			"dev site --draft=false",
			Action{
				typ:           DevServer,
				siteDir:       "site",
				devServerOpts: DevServerOptions{DefaultServerPort, false},
			},
			nil,
		},
		{
			"dev site -D -D",
			Action{},
			fmt.Errorf(
				"failed to parse options: multiple options given for --draft",
			),
		},
		{
			"dev site --draft -D",
			Action{},
			fmt.Errorf(
				"failed to parse options: multiple options given for --draft",
			),
		},
		{
			"dev site --draft=false -D",
			Action{},
			fmt.Errorf(
				"failed to parse options: multiple options given for --draft",
			),
		},
		{
			"dev site --draft --draft=false",
			Action{},
			fmt.Errorf(
				"failed to parse options: multiple options given for --draft",
			),
		},
		{
			"dev site --port=20 --port=10",
			Action{},
			fmt.Errorf(
				"failed to parse options: multiple options given for --port",
			),
		},
		{
			"dev site --port=20 --draft",
			Action{
				typ:           DevServer,
				siteDir:       "site",
				devServerOpts: DevServerOptions{20, true},
			},
			nil,
		},
		{
			"dev site --port",
			Action{},
			fmt.Errorf(
				"failed to parse options: expected a number to be given for port. ex: --port=4200",
			),
		},
	}

	for _, tt := range tests {
		t.Run(
			fmt.Sprintf("%s", tt.args),
			func(t *testing.T) {
				args := strings.Split(tt.args, " ")
				action, err := parseArgs(args)
				if !errEqual(err, tt.expectedErr) {
					t.Fatalf(
						"wrong err. expected=%v. got=%v",
						tt.expectedErr,
						err,
					)
				}
				if action != tt.expectedAction {
					t.Errorf(
						"wrong action.\nexpected=%+v.\n got=%+v",
						tt.expectedAction,
						action,
					)
				}
			},
		)
	}
}

func errEqual(err1, err2 error) bool {
	if err1 == nil && err2 == nil {
		return true
	}
	if err1 == nil || err2 == nil {
		return false
	}
	return err1.Error() == err2.Error()
}
