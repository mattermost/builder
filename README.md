# `mmbuild`: The Mattermost Build System

The `mmbuild` CLI is a general purpose (ie not Mattermost-specific)
that inserts a layer between a builder (for example `make`) and the
CI/CD system that actually runs it (for example GitLab which powers the mattermost builds) to achieve modern Supply Chain Security features
even when the underlying CI/CD engine does not support them.

This repository contains the CLI for the Mattermost build system. This is a thin
repo, which leverages the code currently hosted in
[mattermost/cicd-sdk](https://github.com/mattermost/cicd-sdk) to do its job.

## Purpose

The main goal of `mmbuild` is to enable any CI/CD system to secure its build
with the latest Supply Chain Security technologies like provenance attestations,
digital signatures and SBOMs.

`mmbuild` controls the inputs and outputs of a build. It allows a project
to have predicable (ie reproducible) outcomes by feeding specific artifacts
and by controlling the environment where the builder runs.

Builds can be chained together in a pipeline by passing common environment
variables between them and by feeding the artifacts from a build as the
build materials from the next without relying on the artifacts of the 
CI/CD engine running the job.

Any build can be replayed by consuming the provenance attestation from
produced after every successful run.

## Features

The following are some of the features of the `mmbuild` system. It is not comprehensive and hopefully its up to date:

* Builds are defined from a YAML configuration file that lives in the project
repository. 

* Provenance attestations: Runs generate (and consume as inputs)
[in-toto](https://in-toto.io) attestations with [SLSA](https://slsa.dev)
v0.2 predicates.

* Support for loading environment variables from different sources: 1:1 from
the running environment, hardcoded in the config file or from a secrets backend.

* Extendable handlers that can be used to add more builders to `mmbuild`. The
initial version of mmbuild has a handler for running `make`.

* Pluggable [storage backend system](docs/storage.md) supporting build materials
stored in a variety of backend. Currently s3, http, local filesystems and git are 
supported. More builders can be added if fetching artifacts from other systems
is needed.

* Native git repository cloning, avoids having a separate clone step in the 
pipeline.

* Smarter text replacements. Before running a build, `mmbuild` can search
and replace in files flags to specified values which can be read from secrets.

### Upcoming Features in the Roadmap

The following features have not yet implemented but are planned for 
future releases:

* Software Build of Materials (SBoM): Builds can generate SPDX SBOMs describing
the artifacts the produce.


* Digital Signatures: The most important part missing for now is
supporting digital signatures for artifacts and container images but
also for the attestations and SBOMs. Before accepting artifacts in
a build, their provenance should be cryptographically checked.

## Design

The Mattermost build system takes some ideas from the
[Secure Software Factory](https://docs.google.com/document/d/1FwyOIDramwCnivuvUxrMmHmCr02ARoA3jw76o1mGfGQ),
a secure supply chain reference implementation from the [OpenSSF](https://openssf.org/).

The build replay system uses [in-toto](https://in-toto.io/) attestations with
[SLSA](https://slsa.dev/) [v0.2 predicates](https://slsa.dev/provenance/v0.2) that allow
`mmbuilder` to execute a builder with the captured metadata of a previous run.