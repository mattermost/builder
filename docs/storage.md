# Storage Backends

`mmbuild` can read and write objects (files) to different storage
backends which can be used as build materials (inputs) or build
subjects (artifacts).

## Storage URLs

Objects are referenced to a location in the backend system using
[SPDX Locator](https://spdx.github.io/spdx-spec/package-information/#77-package-download-location-field) URLs. These locators are compatible
with standard URLs:

```
https://example.com/file.txt
```

But can also reference an object in time, stored in a VCS system with
a specific transport:

```
git+https://git.myproject.org/MyProject#file.go

git+ssh://git.myproject.org/MyProject.git@da39a3ee5e6b4b0d3255bfef95601890afd80709#file.go

```

Not all backend systems suport all I/O operations. Some of them are not
yet implemented, and some of them don't make much sense. For example, storing
objects in HTTP servers.

## Supported Backends:

As of this writing, the supported backends include the following:

| URI Prefix | Backend | 
| --- | --- |
| file:// | Local filesystem |
| s3:// | AWS S3 |
| git://<br>git+https://<br>git+ssh:// | git repository |
| http://<br>https:// | HTTP server |

Backends are configured via environment variables. For example, the AWS S3 backend
gets its credentials from env vars which are read by the s3 client.

## Extending the Storage System

To add a new backend, add a new golang file with a struct implementing
the [`object.Backend` interface](https://github.com/mattermost/cicd-sdk/blob/c9a662396e1ec40dea34ea4fb7c5770c133746ec/pkg/object/backends/backends.go#L10-L16).
At the time of this writing the interface
exposes these methods:

```golang
type Backend interface {
	URLPrefix() string
	CopyObject(srcURL, destURL string) error
	Prefixes() []string
	PathExists(string) (bool, error)
	GetObjectHash(string) (map[string]string, error)
}
```

The functions should be self explainatory from their names.

Once the interface implementation is complete, simply add the new backend
to the [list of supported backends](https://github.com/mattermost/cicd-sdk/blob/c9a662396e1ec40dea34ea4fb7c5770c133746ec/pkg/object/object.go#L29-L34)
in the `object` package:

```golang
    // Add the implemented backends
	om.Backends = append(om.Backends,
		backends.NewFilesystemWithOptions(&backends.Options{}),
		backends.NewS3WithOptions(&backends.Options{}),
		backends.NewGitWithOptions(&backends.Options{}),
		backends.NewHTTPWithOptions(&backends.Options{}),
	)
```