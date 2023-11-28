# Development

To setup for development you will need to:

1. Build the docker image specifying the correct target.
2. Start the docker container with the correct volumes and user settings.

These are explained in more detail below.

## Build the docker image

The [Dockerfile](/Dockerfile) contains a `dev` target suitable for development.
To build an image from that target run the following:

```bash
docker build \
    --network=host \
    --target dev \
    --tag concertim-metric-reporting-daemon:latest \
    .
```

## Start the docker container

For development there are two contexts in which you might want to run MRD.

1. Standalone mode in which calls to `concertim-visualisation-app` are mocked.
2. Integrated mode which requires a running `concertim-visualisation-app` to be
   accessible.

To run MRD in *standalone* mode use the following:

```bash
docker run \
    --network=host \
    --volume .:/app \
    --volume ~/.cache/go-build:/.cache/go-build \
    --user "$(id -u):$(id -g)" \
    --env CONFIG_FILE=./config/config.canned.yml \
    concertim-metric-reporting-daemon
```

To run MRD in *integrated* mode, first start `concertim-visualisation-app` so
that it is exposed on the host's port `7000`, and then run the following:

```bash
docker run \
    --network=host \
    --volume .:/app \
    --volume ~/.cache/go-build:/.cache/go-build \
    --user "$(id -u):$(id -g)" \
    --env CONFIG_FILE=./config/config.dev.yml \
    --env JWT_SECRET=<JWT_SECRET> \
    concertim-metric-reporting-daemon
```

`<JWT_SECRET>` needs to be the same JWT secret that
`concertim-visualisation-app` has been configured with.


## Persistent RRD files

The commands given in "Start the docker container" will have the RRD files
created in a directory that will be destroyed when the container is destroyed.
If you want the RRD files to persist beyond the containers life add an
additional `--volume` to the command.

First decide where on the host machine the RRDs are going to be created and
create that directory.  Then start the docker container with an additional
`--volume` flag.  The example below has the RRDs save to `./tmp/rrds`.

```bash
docker run \
    <SNIP>
    --volume ./tmp/rrds:/var/lib/metric-reporting-daemon/rrds/
    <SNIP>
```
