# Building Projects With mmbuild

This guide explains how to run a build using `mmbuild`. 

<!-- toc -->
- [Basic flow](#basic-flow)
- [Configuration File](#configuration-file)
  - [runner](#runner)
  - [artifacts](#artifacts)
    - [Provenance Metadata](#provenance-metadata)
  - [Environment](#environment)
  - [Materials](#materials)
    - [Main Source Repository](#main-source-repository)
- [Running the Build](#running-the-build)
  - [Command Line Interface](#command-line-interface)
  - [Invoking mmbuild:](#invoking-mmbuild)
  - [Exit Codes](#exit-codes)
  - [Output and logs](#output-and-logs)
- [Running mmbuild in GitLab](#running-mmbuild-in-gitlab)
  - [Sharing the Environment](#sharing-the-environment)
- [Staging Path](#staging-path)
- [Concepts](#concepts)
<!-- /toc -->

## Basic flow

A build managed by `mmbuild` in its most basic form will go through the following flow:

1. Read a build configuration file
2. Pull all build materials 
3. Set up a controlled environment as defined in the build configuration
4. Invoke the selected runner
5. Ensure all artifacts are produced 
6. Store the build outputs and its provenance attestation

## Configuration File

The `mmbuild` configuration file will usually live inside the main source
repository of the project. When the builder runs, the first thing it will do is try
to look for the file in the specified directory.

Here is an example, basic configuration file:

```yaml
---
runner:
    id: make
    params: ["compile"]
provenance: /tmp/slsa/
artifacts:
  destination: s3://mattermost-development-test/build-test/
  files: ["binary"]
env:
  - var: SHA_COMMIT
  - var: SHA_COMMIT_WITH_VAL
    value: b739074e0260def700eb13b2aa6091cae9366327
materials:
  - uri: s3://mattermost-development-test/materials/add-on.tar.gz
```

Let's go step by step thhrough what this configuration does:

### runner

`mmbuild` supports pluggable runners. For now only one is implemented: `make`.
This is a simple runner that calls make passing it the defined parameters. In the 
example above, there is one parameter, "compile". This is equivalent to 
running this in your shell:

```bash
make compile
```

The difference here from running it manually is that the environment is controlled 
by mmbuild and all materials are made available beforehand to the runner.

### artifacts

When the example above runs, it expects the runner to produce one artifact: a file
called simply `binary`. The path of the artifact is relative to the working directory
passed with the `-w` flag to `mmbuild`.

The other key in the artifacts section of the file is `destination:`. This value points
to a storage location where all artifacts will be stored after the build runs. For more
info about the backends see the [storage document](storage.md). All artifacts will be 
stored inside of a subdirectory in the destination URI formed by appending the 
`$MMBUILD_STAGING_PATH` value (see [Staging Path](#staging-path) below).

#### Provenance Metadata

The artifacts section also determines where the provenance metadata will be written.
`mmbuild` will always write the provenance attestation in the artifacts destination, under
the `MMBUILD_STAGING_PATH`. The Mattermost build system makes available the full URL
in the `MMBUILD_STAGING_URL` variable, which means that you can retrieve the 
provenance attestation from this location:

```
    ${MMBUILD_STAGING_URL}/provenance.json
```

### Environment

In this example, the builder will feed two environment variables to the runner. The way
these two are defined has a slight difference:

```yaml
env:
  - var: SHA_COMMIT
  - var: SHA_COMMIT_WITH_VAL
    value: b739074e0260def700eb13b2aa6091cae9366327
```

The first variable `SHA_COMMIT` will be read from `mmbuild`'s running environment and 
passed through to the runner. The second variable will be defined to the runner with
the fixed value from the config file `b73907...`. 

### Materials

The last part of the file lists the building materials. In the exaple, the builder will
download a single file and make it available for the runner inside the materials directory.
This directory can be used byt the runner and its scripts by reading the `MMBUILD_MATERIALS_DIR`
environment variable. 

In this example, the build material entry:

```
s3://mattermost-development-test/materials/add-on.tar.gz
```

Will be made available to the make runner in:

```
${MMBUILD_MATERIALS_DIR}/add-on.tar.gz
```

Note that the current mmbuild version will only handle single files as materials, recursive 
directory copying is a feature to be implemented in the future. 

#### Main Source Repository

It is not required to list the main source repository in the build materials section.

When running a build, `mmbuild` will detect if the working directory is a git repository
and record it as the first build material in the provenance attestation. This behavior
is currently hardcoded.  

## Running the Build

### Command Line Interface

The build subcommand in Version 0.2 of `mmbuild` has a simple CLI interface with
just a few flags for now:

```
]$ mmbuild build --help

Usage:
   build [flags]

Flags:
      --conf string      configuration file for the build
  -h, --help             help for build
  -w, --workdir string   working directory where the build will run (default ".")

```

`-w` or `--workdir` points to the working directory where the main source code lives.
By default, mmbuild will look in this directory for a file called `matterbuild.yaml` at
the root of the repo and use it as its configuration file. It defaults to the current
working directory.

The `--conf` flag allows to specify a different configuration file. 

### Invoking mmbuild:

To start a build, simply execute the `mmbuild` binary and point it to the root of the
repository:

```
]$ mmbuild build -w Projects/repository/
INFO[0000] Loading build configuration from Projects/repository/matterbuild.yaml 
INFO[0000] Replacing 1 configuration variables in YAML code ([PLUGIN_STORE_URL]) 
INFO[0000] > YAML conf variable SHA_COMMIT_WITH_VAL set to value
           'b739074e0260def700eb13b2aa6091cae9366327' from predefined environment 
INFO[0000] No configuration variables found in YAML code 
INFO[0000] Build conf:
.... (rest of output trimmed)
```

### Exit Codes

mmbuild supports only exit code 0 for success and non-zero for failures.

### Output and logs

All output is written to STDERR. In addition, the output of the runner is captured to a
logfile. Which is currently written to the system temporary directory (this should change
in future versions).

## Running mmbuild in GitLab

`mmbuild` is designed to sit between CICD system (ie GitLab) and the underlying runner
(ie `make`). This way it can be started by the CI and control the I/O of the runner.

As of this writing, the current approach to run `mmbuild` in a pipeline involves downloading
the binary from the latest GitHub release and executing it in the host container[^1]. As
releases are more solid this approach will change and binary verification should be performed.
The following `.gitlab-ci.yml` file has an example of building the mattermost webapp repo:

[^1]: Current plans are to make mmbuild available in a container image. 

```yaml
workflow:
  rules:
    - if: '$CI_PIPELINE_SOURCE != "push"'

stages:
  - build

clone:
  stage: build
  image: $HOST_IMAGE
  script:
    - git clone https://github.com/mattermost/mattermost-webapp.git --depth=1
  artifacts:
    paths:
      - mattermost-webapp
mmbuild:
  stage: build
  image: $HOST_IMAGE
  needs:
    - clone
  script: 
    - cd mattermost-webapp
    - curl -Lo mmbuild https://github.com/mattermost/builder/releases/download/$MMBUILDER_VERSION/mmbuilder_linux_amd64
    - chmod 0755 mmbuild
    - ./mmbuilder build
  artifacts:
    reports:
      dotenv: mattermost-webapp/build.env
variables:
  HOST_IMAGE: $CI_REGISTRY/mattermost/ci/images/mattermost-build-webapp:20210524_node-16
  MMBUILDER_VERSION: v0.0.2
```

### Sharing the Environment

At some point, a more complicated pipeline will have to chaing several builds together. In order to pass
values among them, `mmbuild` sets a few environment variables after each run. In the example above, note
under the `artifacts:` section how the configuration is exporting `mmbuild`'s environment by reading 
the `build.env` dotenv report. This gives the next job in the pipeline a way to consume values from
the `mmbuild` execution.

Currently, the two variables that are exported are:

* `MMBUILD_STAGING_PATH`: The computed hash to content-address the build
* `MMBUILD_STAGING_URL`: The full URI pointing to where the artifacts are staged

## Staging Path

Each build run can be content-addressed by its `STAGING_PATH`. The staging path of a build
is a sha256 hash constructed by computing the build's materials source and variables[^2].

[^2]: As of this writing, the STAGING_PATH computation is not taking into account the
environment, see [issue #19 in mattermost/cicd-sdk](https://github.com/mattermost/cicd-sdk/issues/19)
for more info.

The idea behind the value is to have a way to address a build by its inputs. If you
want to check if artifacts exist for a specific build, just check the artifacts directory
under the `STAGING_PATH` and you know.

## Concepts
Some of the docs and concepts in mmbuild use the [SLSA](https://slsa.dev/) 
vocabulary. Some terms have simple 1:1 equivalents to common DevOps
terminology and some are new but are important to understand:

* **Provenance:** A piece of metadata that provides information about the origin
of a build's artifacts. The provenance metadata describes what _ingredients_
went into a build, what process was applied to transform the inputs and 
finally what came out of the build. 

* **Attestation:** A file making a statement or verification about certain subjects.
It will be often found as "provenance attestation", referring a file that _attests_ to
the origin of a bunch of artifacts.

* **Subjects:** Subjects are the artifacts described by an attestation. They can be
the artifacts of a build, but not necesarily. Also, listing artifacts in a 
provenance attestation does not imply a complete listing of every file a build produced. 

* **Materials:** A build inputs. When a build runs, it will frequently pull packages
it needs to assmble the project. These external files are called materials. Everything
that goes into a project to assemble the final product is a material. 

* **Staging Path:** A sha256 hash computed from a builds inputs, source and environment.
It allows consumers of the build to content-address the run and its artifacts.

* **Runner:** mmbuild works by controlling a _runner_'s environment and inputs. A runner
is a program executed by the build system that does the actual work of building the
project. Most promiment example: `make`.
