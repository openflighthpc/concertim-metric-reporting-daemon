# Example scripts for using the metric reporting API

These example scripts demonstrate usage of the metric reporting API.  Example
scripts for the rack and device API can be found at
https://github.com/alces-flight/concertim-ct-visualisation-app/tree/main/docs/api/examples

For the best experience you will want to ensure that you are viewing the
example scripts for the same release as the Concertim instance.

## Authentication and selecting the Concertim instance

The example scripts are created to make it easy to specify which Concertim
instance to communicate with.  This is done via setting the `CONCERTIM_HOST`
environment variable.  If this is not set it will default to
`command.concertim.alces-flight.com`.

It is currently left as an exercise for the API user to either ensure that that
the name `command.concertim.alces-flight.com` resolves to the correct IP
address or alternatively to specify `CONCERTIM_HOST`.

To use the API an authentication token needs to be obtained.  The example below
does so using a Concertim instance available at `10.151.15.150`.  The snippet
is intended to be ran from within the rack and device API example script
directory.

```
export CONCERTIM_HOST=10.151.15.150
export AUTH_TOKEN=$(LOGIN=admin PASSWORD=admin ./get-auth-token.sh)
```

With `AUTH_TOKEN` and `CONCERTIM_HOST` both exported the other API example
scripts can be used.

## Metric reporting API usage

### Eventual consistency

There is a delay of up to a minute after a device has been created through the
device API before metrics can be reported for it.

Once metrics have been reported, there is a delay of up to a minute before
those metrics appear on the interactive rack view.

### Example scripts

The snippets below assume that `CONCERTIM_HOST` and `AUTH_TOKEN` are set as described above.

Report a single string metric for a device.  If `DEVICE_NAME` is not given it
will default to `comp001`, `METRIC_NAME` defaults to `caffeine.more` and
`VALUE` to `yes`.

```
./string-metric.sh [DEVICE_NAME [METRIC_NAME [VALUE]]]
```

Report a single int32 metric for a device. If `DEVICE_NAME` is not given it
will default to `comp001`, `METRIC_NAME` defaults to `caffeine.level`,
`VALUE` to a random integer between 12 and 24 and `UNIT` to `""`.

```
./int32-metric.sh [DEVICE_NAME [METRIC_NAME [VALUE [UNIT]]]]
```

Report a single double metric for a device. If `DEVICE_NAME` is not given it
will default to `comp001`, `METRIC_NAME` defaults to `caffeine.max` and
`VALUE` to a random double between 0 and 10.

```
./double-metric.sh [DEVICE_NAME [METRIC_NAME [VALUE]]]
```

Report multiple int32 metrics for a device.  If `DEVICE_NAME` is not given it
will default to `comp001`.

```
./multiple-int32-metrics.sh [DEVICE_NAME]
```
