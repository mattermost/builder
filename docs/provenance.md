# Provenance Metdata

<!-- toc -->
- [Why Do We Need Provenance Metadata?](#why-do-we-need-provenance-metadata)
- [Metadata Structure](#metadata-structure)
- [Main Sections](#main-sections)
  - [Subjects](#subjects)
  - [Materials](#materials)
  - [Invocation](#invocation)
  - [Build Configuration File](#build-configuration-file)
<!-- /toc -->

## Why Do We Need Provenance Metadata?

Runs executed by `mmbuild` produce provanance attestations that capture the 
information that the system used to produce a given set of artifacts.
Theoretically, a controlled build can be rerun using the same inputs and
under the same environment and produce the same artifacts, verifiable bit
by bit.

When running the `mmbuild replay` subcommand, the Mattermost build system
can read the provenance metdata and execute the build again as it was run
when it was originally executed.

## Metadata Structure

The metadata files produced by mmbuild conform to the in-toto attestation
specification. The [in-toto](https://in-toto.io/) framework provides software
build systems a specification to describe and verify each stage of the build
as it runs. 

An attestation describes _subkects_ and a _predicate_. The predicate in the
`mmbuild` runs conforms to the [SLSA framework](https://slsa.dev/) 
(Supply chain Levels for Software Artifacts)
[version 0.2 provenance attestation](https://slsa.dev/provenance/v0.2). 

Here is an example of a build metadata file:

```json
{
  "_type": "https://in-toto.io/Statement/v0.1",
  "predicateType": "https://slsa.dev/provenance/v0.2",
  "subject": [
    {
      "name": "mattermost-webapp.tar.gz",
      "digest": {
        "sha256": "f0506779127b08be23e37c6022d1fd51418bb0b6481665fdeb51bd4bbed904d4",
        "sha512": "f56b6f3d60a59fce7675d5f7f564a34b5c9632401d2d00ac7a67bf6258458a51420283e1d99e18113668545c55e6ac95e5d5b1b19e5acaa886f9842c5f10080e"
      }
    }
  ],
  "predicate": {
    "builder": {
      "id": "MatterBuild/v0.1"
    },
    "buildType": "make",
    "invocation": {
      "configSource": {
        "uri": "matterbuild.yaml",
        "digest": {
          "sha1": "53868e12c770d6da4195955cf94a5b53b82a7c86"
        }
      },
      "parameters": [
        "package"
      ],
      "environment": {
          "PLUGIN_STORE_URL": "https://plugins-store.test.mattermost.com/release",
      }
    },
    "metadata": {
      "buildStartedOn": "2021-12-21T01:19:10.727607258Z",
      "buildFinishedOn": "2021-12-21T01:19:13.248116432Z",
      "completeness": {
        "parameters": true,
        "environment": true,
        "materials": true
      },
      "reproducible": true
    },
    "materials": [
      {
        "uri": "git+https://github.com/mattermost/mattermost-webapp.git",
        "digest": {
          "sha1": "53868e12c770d6da4195955cf94a5b53b82a7c86"
        }
      }
    ]
  }
}

```
## Main Sections

Let's go over some of the important parts of the example above.

### Subjects

The subjects in an attestation are the artifacts produced after a build.
In the example above, `mmbuild` ran and produced one file:
`mattermost-webapp.tar.gz`. One provenance file can attest to the origin
of any number of subjects (artifacts). The files are recorded along with
their sha256 and sha512 hashes.

### Materials

The build materials are all files consumed by the build to assemble its
artifacts. In the example above, only one build material is defined:
the project's source code, which was cloned from the `mattermost/mattermost-webapp`
repository. The build ran at commit  53868e12c770d6da4195955cf94a5b53b82a7c86.

### Invocation

The invocation is the actual command that `mmbuild` ran to execute the
builder. In this case the buildType is _make_. This part in the spec is arbitrary,
the value is written by the `mmbuild` `make` builder and the sole parameter is 
`package`. If replayed, the build system would use the make builder to run
`make module` in the repo.

The environment variables defined when the build ran are captured in the `environment`
key. There is one sole variable in the example: `PLUGIN_STORE_URL`.

### Build Configuration File

The provenance attestation registers the configuration file used to run the build.
In our case, this points to the `matterbuild.yaml` used by `mmbuild` to run. In the
example, the URI is just the filename (matterbuild.yaml) which means that the file
was read from the source repository at the commit shown[^1]

[^1]: The syntax to specify artifacts in git repos is abigous in the spec. As SLSA
is rapidly evolving, this will surely change in the upcoming versions.

