mistry
====================================
[![Build Status](https://api.travis-ci.org/skroutz/mistry.svg?branch=master)](https://travis-ci.org/skroutz/mistry)

mistry executes user-provided build jobs inside pre-defined, isolated
environments and makes the results available for later consumption.

It enables fast workflows by employing caching techniques and incremental
builds due to its copy-on-write snapshotting features.

Features include:

- running arbitrary commands inside isolated environments (provided as Docker images)
- build caching & incremental building (see [*"Build cache"*](https://github.com/skroutz/mistry/wiki/Build-cache))
- a simple JSON API for interacting with the server (scheduling jobs etc.)
- ([wip](https://github.com/skroutz/mistry/pull/17)) a web view for inspecting the progress and result of builds
- efficient use of disk space due to copy-on-write semantics (using [Btrfs snapshotting](https://en.wikipedia.org/wiki/Btrfs#Subvolumes_and_snapshots))

For more information take a look at the [wiki](https://github.com/skroutz/mistry/wiki).






Status
-------------------------------------------------
mistry project is still in alpha and is not yet recommended for use in
production environments until we reach the 1.x series.






Getting started
-------------------------------------------------
You can get the binaries from the
[latest releases](https://github.com/skroutz/mistry/releases).

Alternatively, if you have Go 1.9 or later installed you can get the
latest development version:

```shell
$ go get -u github.com/skroutz/mistry
```





Usage
--------------------------------------------------
To boot the server, a configuration file is needed. You can use [`config.sample.json`](config.sample.json)
as a starting point:

```shell
$ mistryd --config config.json
```

Use `mistryd --help` for more info.

The paths denoted by `projects_path` and `build_path` settings should already
be created and writable.






### Adding projects

The `projects_path` path should contain all the projects known to mistry.
These are the projects for which jobs can be built.

Refer to [File system layout - Projects directory](https://github.com/skroutz/mistry/wiki/File-system-layout#projects-directory) for more info.






### API

Interacting with mistry is done in two ways ; (1) using `mistry` or (2)
using directly the JSON API. We recommended using `mistry` whenever possible
(although it may occassionally lag behind the server in terms of
supported features).

**Scheduling a new job and fetching the artifacts:**
```shell
$ mistry --project foo
```
This will place the artifacts and the current working directory. See `mistry build -h` for more options.

**Scheduling a new job without fetching the artifacts**:
``` shell
$ curl -H 'Content-Type: application/json' -d '{"project": "foo"}' localhost:8462/jobs
```







Configuration
-------------------------------------------------
The following settings are currently supported:

| Setting        | Description           | Default  |
| ------------- |:-------------:| -----:|
| `projects_path` (string)      | The path where project folders are located | "" |
| `build_path` (string)      | The root path where artifacts will be placed       |   "" |
| `mounts` (object{string:string}) |  The paths from the host machine that should be mounted inside the execution containers     |    {} |

For a sample configuration file refer to [`config.sample.json`](config.sample.json).




Credits
-------------------------------------------------
mistry is released under the GNU General Public License version 3. See [COPYING](COPYING).
