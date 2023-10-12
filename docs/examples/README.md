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

Report a single string metric for a device.  If `DEVICE_ID` is not given it
will default to `1`, `METRIC_NAME` defaults to `caffeine.more` and
`VALUE` to `yes`.

```
./string-metric.sh [DEVICE_ID [METRIC_NAME [VALUE]]]
```

Report a single int32 metric for a device. If `DEVICE_ID` is not given it
will default to `1`, `METRIC_NAME` defaults to `caffeine.level`,
`VALUE` to a random integer between 12 and 24 and `UNIT` to `""`.

```
./int32-metric.sh [DEVICE_ID [METRIC_NAME [VALUE [UNIT]]]]
```

Report a single constant valued int32 metric for a device. If `DEVICE_ID` is
not given it will default to `1`, `METRIC_NAME` defaults to
`caffeine.capacity`, `VALUE` is hardcoded to 10 and `UNIT` to `""`.  The metric
is treated as constant due to the script setting a `slope` of `zero`.

```
./int32-metric.sh [DEVICE_ID [METRIC_NAME [UNIT]]]
```

Report a single double metric for a device. If `DEVICE_ID` is not given it
will default to `1`, `METRIC_NAME` defaults to `caffeine.max` and
`VALUE` to a random double between 0 and 10.

```
./double-metric.sh [DEVICE_ID [METRIC_NAME [VALUE [UNIT]]]]
```

Report multiple int32 metrics for a device.  If `DEVICE_ID` is not given it
will default to `1`.

```
./multiple-int32-metrics.sh [DEVICE_ID]
```

## Metric querying API usage

Get a list of current metrics that were processed in the most recent processing run.

```
./list-current-metrics.sh
```

Get a list of historic metrics.

```
./list-historic-metrics.sh
```

Get the current values for all devices that reported a value for the given
metric in the most recent processing run.  `METRIC_NAME` defaults to
`caffeine.level`.

```
./list-current-metric-values.sh [METRIC_NAME]
```

Get the historic metric values between the given times for all devices that
have reported a value for the given metric.  `START` and `END` must be
formatted as an integer number of seconds since the UNIX epoch. `METRIC_NAME`
defaults to `caffeine.level`.  If Ruby is installed, `START` defaults to one
hour ago and `END` defaults to now.  If Ruby is not installed, arbitrary
hardcoded defaults are used.

```
./list-historic-metric-values.sh [METRIC_NAME] [START] [END]
```

Get a list of current metrics for a single device.  `DEVICE_ID` defaults to `1`.

```
./list-current-device-metrics.sh [DEVICE_ID]
```

Get a list of historic metrics for a single device.  `DEVICE_ID` defaults to `1`.

```
./list-historic-device-metrics.sh [DEVICE_ID]
```

Get the historic metric values for the given device between the given times.
`START` and `END` must be formatted as an integer number of seconds since the
UNIX epoch.  `DEVICE_ID` defaults to `1`. `METRIC_NAME` defaults to
`caffeine.level`.  If Ruby is installed, `START` defaults to one hour ago and
`END` defaults to now.  If Ruby is not installed, arbitrary hardcoded defaults
are used.

```
./list-historic-device-metric-values.sh [DEVICE_ID] [METRIC_NAME] [START] [END]
```
