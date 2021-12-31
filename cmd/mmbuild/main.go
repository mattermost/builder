package main

import (
	"path/filepath"

	"github.com/mattermost/cicd-sdk/pkg/build"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{}

var buildCmd = &cobra.Command{
	Use: "build",
	// SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runBuild(bOpts)
	},
}

var replayCmd = &cobra.Command{
	Use:           "replay",
	SilenceErrors: true,
	Args:          cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return replayBuild(args[0], rOpts)
	},
}

type buildOpts = struct {
	forceBuild bool
	SBOM       bool
	configFile string
	workDir    string
}

type replayOpts = struct {
	workDir string
}

var (
	bOpts = &buildOpts{}
	rOpts = &replayOpts{}
)

func init() {
	// Options for mmbuild build
	buildCmd.PersistentFlags().StringVar(
		&bOpts.configFile, "conf", "", "configuration file for the build",
	)
	buildCmd.PersistentFlags().StringVarP(
		&bOpts.workDir, "workdir", "w", ".", "working directory where the build will run",
	)
	buildCmd.PersistentFlags().BoolVarP(
		&bOpts.forceBuild, "force", "f", false, "execute the builder even if artifacts are found",
	)
	buildCmd.PersistentFlags().BoolVar(
		&bOpts.SBOM, "sbom", false, "write an sbom to the workind directory after building",
	)

	// Options for mmbuild replay
	replayCmd.PersistentFlags().StringVarP(
		&rOpts.workDir, "workdir", "w", ".", "working directory where the replay will run",
	)
	replayCmd.PersistentFlags().BoolVarP(
		&bOpts.forceBuild, "force", "f", false, "execute the builder even if artifacts are found",
	)

	rootCmd.AddCommand(buildCmd)
	rootCmd.AddCommand(replayCmd)
}

func runBuild(opts *buildOpts) (err error) {
	var b *build.Build
	if opts.configFile == "" {
		b, err = build.NewFromConfigFile(filepath.Join(opts.workDir, build.ConfigFileName))
	} else {
		b, err = build.NewFromConfigFile(opts.configFile)
	}
	if err != nil {
		return errors.Wrap(err, "creating new build")
	}

	b.Options().SBOM = opts.SBOM
	b.Options().Workdir = opts.workDir
	b.Options().ForceBuild = opts.forceBuild

	run := b.Run()
	return errors.Wrap(run.Execute(), "executing build run")
}

func replayBuild(attestationPath string, ropts *replayOpts) (err error) {
	opts := &build.Options{Workdir: ropts.workDir}
	b, err := build.NewFromAttestation(attestationPath, opts)
	if err != nil {
		return errors.Wrap(err, "creating build")
	}

	return errors.Wrap(b.RunAttestation(attestationPath), "executing replay run")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		logrus.Fatal(err)
	}
}
