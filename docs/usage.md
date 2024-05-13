# Base URL

When the metric reporting daemon is running on a Concertim appliance available at, say, `concertim.alces-flight.com`, its URL will be `https://concertim.alces-flight.com/mrd`.  When running locally during development the URL will be `http://localhost:3000`.

All URL paths in this document are relative to this URL.

# Reporting a metric

Requests to report metrics should be authenticated with a JWT token, see the Authentication section below for more details.

A metric is reported to the URL `/:device_id/metrics` where `:device_id` is the ID of a device already known to Concertim, e.g., `1`.

The body is a JSON document containing the keys, `name`, `value`, `units`, `type`, `slope` and `ttl`.

`name`
: the metric's name, a dotted prefix can be used such as `ct.ipmi`, `ct.snmp` or `ct.user`.

`type`
: the type of the metric's value.  This must be one of `string`, `int8`, `uint8`, `int16`, `uint16`, `int32`, `uint32`, `float` or `double`.

`value`
: the metric's value.  This must be a JSON value consistent with the given `type`.  E.g., if the `type` is `string` then a JSON string must be given, otherwise a JSON number must be given.

`units`
: the units for the metric.

`slope`
: an indication of how the metrics value can change over time.  Valid values are `zero`, `positive`, `negative`, `both` or `derivative`.  `zero` is for values that will not change; `postitive` for values that will only increase; `negative` for values that will only decrease; `both` for values that may increase or decrease.

`ttl`
: the time in seconds until the metric should be considered stale.  Once metrics are considered to be stale they are removed from the metric stream.

E.g.,

```
PUT /:device/metrics
Content-Type: application/json
Authorization: Bearer <TOKEN>
{
  "name": "my-metric",
  "value": 12,
  "units": "my-units",
  "type": "uint32",
  "slope": "both",
  "ttl": 180
}
```

# Errors

The error format is loosely based on the [JSON:API error](https://jsonapi.org/format/#errors) format.

If the device name is not known to the Concertim appliance a 404 error response will be given with a body of:

```
{
  "errors": [
    {
      "title": "Host Not Found",
      "detail": "Unknown host: <device>"
    }
  ],
  "status": 404
}
```

If the JSON body is not parsable a 400 error response is given with a body of:

```
{
  "errors": [
    {
      "title": "error parsing JSON body",
      "detail": "<detail of parse error>"
    }
  ],
  "status": 400
}
```

If the JSON body has the wrong structure a 400 error response is given with a body of:

```
{
  "errors": [
    {
      "title": "error parsing JSON body",
      "detail": "<partially helpful error message>"
    }
  ],
  "status": 400
}
```

If the JSON body has the correct structure, but the values are not valid, a 422 error response is given with a body of:


```
{
  "errors": [
    {
      "status": 422,
      "title": "<name of failed validation>",
      "detail": "<details about this failure>",
      "source": "<field that the validation is for>"
    }
  ],
  "status": 422
}
```

e.g., 

```
{
  "errors": [
    {
      "status": 422,
      "title": "required",
      "detail": "units is a required field",
      "source": "units"
    },
    {
      "status": 422,
      "title": "oneof",
      "detail": "slope must be one of [zero positive negative both derivative]",
      "source": "slope"
    },
    {
      "status": 422,
      "title": "min",
      "detail": "ttl must be 1 or greater",
      "source": "ttl"
    }
  ],
  "status": 422
}
```

# Retrieving metrics

## `GET /metrics/current`  List current metrics

Lists all current metrics found in the most recent processing run.  If a metric
is reported by multiple devices it will appear in the output just once.

### Response Codes

* `200 - OK`  Request was successful.
* `500 - Internal Server Error`  An unexpected error occurred.  This should not
  happen.
* `503 - Service Unavailable`  A processing run has not taken place yet.

### Response Parameters

* `id` : `string` : A unique identifier for this metric.
* `name` : `string` : The name of the metric.
* `units` : `string` : The units for this metric.  This is optional and could also be the empty string.
* `nature` : `string` : The nature of the metric.  One of `volatile`, `string_and_time` or `constant`.
* `min` : `any` : The minimum value reported for this metric in the last processing run across all processed devices.
* `max` : `any` : The maximum value reported for this metric in the last processing run across all processed devices.

### Response Example

```
[
  {
    "id": "caffeine.level",
    "name": "caffeine.level",
    "units": "",
    "nature": "volatile",
    "min": 0,
    "max": 99
  },
  {
    "id": "caffeine.consumption",
    "name": "caffeine.consumption",
    "units": "mugs",
    "nature": "volatile",
    "min": 1,
    "max": 4
  }
]
```

## `GET /metrics/<metric_name>/current`  List metric value for all devices reporting that metric

Returns a list containing the reported metric value for all devices that
reported the given metric in the most recent processing run.

### Response Codes

* `200 - OK`  Request was successful.
* `404 - Not Found`  The metric was not present in the last processing run or a processing run has not taken place yet.
* `500 - Internal Server Error`  An unexpected error occurred.  This should not
  happen.
* `503 - Service Unavailable`  A processing run has not taken place yet.

### Request Parameters

* `metric_name` : `string` : The name of the metric for which values should be returned.

### Response Parameters

* `id` : `string` : The identifier for the device.
* `value` : `any` : The value of this metric for this device.

### Response Example

```
[
  {
    "id": "1",
    "value": 12
  },
  {
    "id": "2",
    "value": 24
  }
]
```

## `GET /metrics/<metric_name>/historic/last/<duration>`  List historic metric values for all devices for the last hour, day or quarter

Returns a list containing the reported metric values in the last duration,
where duration is one of hour, day or quarter. Every device that has ever
reported this metric is included even if did not report any values in the given
duration.  If a device has never reported this metric, it is not included.  If
a device did not report a value at some points between the start and end times
the value will be returned as `null`.

### Response Codes

* `200 - OK`  Request was successful.
* `500 - Internal Server Error`  An unexpected error occurred.  This should not
  happen.

### Request Parameters

* `metric_name` : `string` : The name of the metric for which values should be returned.
* `duration` : `string` : The duration to consider.  One of `hour`, `day` or
`quarter`.  Only metric values reported in the last `duration` are returned.

### Response Parameters

* `id` : `string` : The identifier for the device.
* `values` : `array` : An array of historic values reported by this device for this metric.
* `values.value` : `any` : The value of the metric recorded at the
  corresponding timestamp, or `null` if no value was reported at that time stamp.
* `values.timestamp` : `timestamp` : The time the corresponding value was
  recorded as an integer number of seconds since the epoch (1970-01-01:00:00:00).

### Response Example

```
[
  {
    "id": "1",
    "values": [
      {"timestamp": 1696420533, "value": 12},
      {"timestamp": 1696420548, "value": 9},
      {"timestamp": 1696420518, "value": null},
      {"timestamp": 1696420503, "value": 10}
    ]
  },
  {
    "id": "2",
    "values": [
      {"timestamp": 1696420533, "value": 7},
      {"timestamp": 1696420548, "value": 5},
      {"timestamp": 1696420518, "value": 9},
      {"timestamp": 1696420503, "value": 12}
    ]
  }
]
```


## `GET /metrics/<metric_name>/historic/<start_time>/<end_time>`  List historic metric values for all devices between the given start and end times

Returns a list containing the reported metric values between the given start
time and end time.  Every device that has ever reported this metric is included
even if did not report any values between the given times.  If a device has
never reported this metric, it is not included.  If a device did not report a
value at some points between the start and end times the value will be returned
as `null`.

### Response Codes

* `200 - OK`  Request was successful.
* `500 - Internal Server Error`  An unexpected error occurred.  This should not
  happen.

### Request Parameters

* `metric_name` : `string` : The name of the metric for which values should be returned.
* `start_time` : `timestamp` : The start of the time range formatted as an
  integer number of seconds since the epoch (1970-01-01:00:00:00).
* `end_time` : `timestamp` : Optional, defaults to the current time. The end of
  the time range formatted as an integer number of seconds since the epoch
  (1970-01-01:00:00:00).

### Response Parameters

* `id` : `string` : The identifier for the device.
* `values` : `array` : An array of historic values reported by this device for this metric.
* `values.value` : `any` : The value of the metric recorded at the
  corresponding timestamp, or `null` if no value was reported at that time stamp.
* `values.timestamp` : `timestamp` : The time the corresponding value was
  recorded as an integer number of seconds since the epoch (1970-01-01:00:00:00).

### Response Example

```
[
  {
    "id": "1",
    "values": [
      {"timestamp": 1696420533, "value": 12},
      {"timestamp": 1696420548, "value": 9},
      {"timestamp": 1696420518, "value": null},
      {"timestamp": 1696420503, "value": 10}
    ]
  },
  {
    "id": "2",
    "values": [
      {"timestamp": 1696420533, "value": 7},
      {"timestamp": 1696420548, "value": 5},
      {"timestamp": 1696420518, "value": 9},
      {"timestamp": 1696420503, "value": 12}
    ]
  }
]
```

## `GET /devices/<device_id>/metrics/current`  List all current metrics for the given device

Returns a list containing all metrics value for the given device reported in
the most recent processing run.

### Response Codes

* `200 - OK`  Request was successful.
* `404 - Not Found`  The metric was not present in the last processing run or a processing run has not taken place yet.
* `500 - Internal Server Error`  An unexpected error occurred.  This should not
  happen.
* `503 - Service Unavailable`  A metric run has not yet taken place.

### Request Parameters

* `device_id` : `string` : The concertim ID of the device for which metrics should be returned.

### Response Parameters

* `id` : `string` : The identifier for this metric.
* `name` : `string` : The name of the metric.
* `units` : `string` : The units for this metric.  This is optional and could also be the empty string.
* `nature` : `string` : The nature of the metric.  One of `volatile`, `string_and_time` or `constant`.
* `value` : `any` : The value of this metric.

### Response Example

```
[
  {
    "id": "caffeine.level",
    "name": "caffeine.level",
    "units": "",
    "nature": "volatile",
    "value": 20
  },
  {
    "id": "caffeine.consumption",
    "name": "caffeine.consumption",
    "units": "mugs",
    "nature": "volatile",
    "value": 1.6228
  }
]
```

## `GET /devices/<device_id>/metrics/<metric_name>/historic/last/<duration>`  List historic metric values for a single device and metric for the last hour, day or quarter

Returns a list containing the reported metric values in the last duration,
where duration is one of hour, day or quarter. If the device has never reported
this metric, a 404 response is returned.  If the device did not report a value
at some points in the given duration the value will be returned as `null`.

### Response Codes

* `200 - OK`  Request was successful.
* `404 - Not Found`  The device has never reported this metric.
* `500 - Internal Server Error`  An unexpected error occurred.  This should not
  happen.

### Request Parameters

* `device_id` : `string` : The concertim ID of the device for which metrics should be returned.
* `metric_name` : `string` : The name of the metric for which values should be returned.
* `duration` : `string` : The duration to consider.  One of `hour`, `day` or
`quarter`.  Only metric values reported in the last `duration` are returned.

### Response Parameters

* `value` : `any` : The value of the metric recorded at the corresponding
  timestamp, or `null` if no value was reported at that time stamp.
* `timestamp` : `timestamp` : The time the corresponding value was recorded as
  an integer number of seconds since the epoch (1970-01-01:00:00:00).

### Response Example

```
[
  {"timestamp": 1696420533, "value": 12},
  {"timestamp": 1696420548, "value": 9},
  {"timestamp": 1696420518, "value": null},
  {"timestamp": 1696420503, "value": 10}
]
```


## `GET /devices/<device_id>/metrics/<metric_name>/historic/<start_time>/<end_time>`  List historic metric values for a single device and metric between the given start and end times

Returns a list containing the reported metric values between the given start
time and end time for the specified device.  If the device has never reported
this metric, a 404 response is returned.  If the device did not report a value
at some points between the start and end times the value will be returned as
`null`.

### Response Codes

* `200 - OK`  Request was successful.
* `404 - Not Found`  The device has never reported this metric.
* `500 - Internal Server Error`  An unexpected error occurred.  This should not
  happen.

### Request Parameters

* `device_id` : `string` : The concertim ID of the device for which metrics should be returned.
* `metric_name` : `string` : The name of the metric for which values should be returned.
* `start_time` : `timestamp` : The start of the time range formatted as an
  integer number of seconds since the epoch (1970-01-01:00:00:00).
* `end_time` : `timestamp` : Optional, defaults to the current time. The end of
  the time range formatted as an integer number of seconds since the epoch
  (1970-01-01:00:00:00).

### Response Parameters

* `value` : `any` : The value of the metric recorded at the corresponding
  timestamp, or `null` if no value was reported at that time stamp.
* `timestamp` : `timestamp` : The time the corresponding value was recorded as
  an integer number of seconds since the epoch (1970-01-01:00:00:00).

### Response Example

```
[
  {"timestamp": 1696420533, "value": 12},
  {"timestamp": 1696420548, "value": 9},
  {"timestamp": 1696420518, "value": null},
  {"timestamp": 1696420503, "value": 10}
]
```

# Authentication

Requests requiring authentication should set the `Authorization` header using the `Bearer` authentication strategy.  The token should be a JWT token, which can be created as described below.

When deployed as part of the Concertim appliance, an API token can be created by using the ct-visualisation-app [get-auth-token.sh example script](https://github.com/openflighthpc/concertim-ct-visualisation-app/blob/main/docs/api/examples/get-auth-token.sh), which will print the auth token to standard output.

When running locally, an API token can be created by running, `go run cmd/create-auth-token/main.go`.  The auth token will be printed to standard output.

```bash
$ go run cmd/create-auth-token/main.go --config-file ./config/config.canned.yml
eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3MDEzNjg4ODl9.xfwqABmzcjpRIWU8cvjP8daIgFmF5dWkR2m1CSOLkL4
```

When running in docker container you can use the following:

```bash
$ docker run --rm --volume .:/app concertim-metric-reporting-daemon go run ./cmd/create-auth-token/main.go --config-file ./config/config.canned.yml 
eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3MDEzNzAxNDJ9.HXSKYVXKDnA2-hRHBtpesrpIYVxav7l1M4U2IOl1Gc8
```
