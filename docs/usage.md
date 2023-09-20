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

## `GET /metrics/unique`  List unique metrics

Lists all unique metrics found in the most recent processing run.  The
uniqueness of a metric is determined on the metric's name, so if two devices
report a metric, say, `load.one`, that will result in a single unique metric.

### Response Codes

* `200 - OK`  Request was successful.
* `500 - Internal Server Error`  An unexpected error occurred.  This should not
  happen.

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

## `GET /metrics/<metric_name>/values`  List metric value for all devices reporting that metric

Returns a list containing the reported metric value for all devices that
reported the given metric in the most recent processing run.

### Response Codes

* `200 - OK`  Request was successful.
* `500 - Internal Server Error`  An unexpected error occurred.  This should not
  happen.

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

# Authentication

Requests requiring authentication should set the `Authorization` header using the `Bearer` authentication strategy.  The token should be a JWT token, which can be created as described below.

When deployed as part of the Concertim appliance, an API token can be created by using the ct-visualisation-app [get-auth-token.sh example script](https://github.com/alces-flight/concertim-ct-visualisation-app/blob/main/docs/api/examples/get-auth-token.sh), which will print the auth token to standard output.

When running locally, an API token can be created by running, `go run cmd/create-auth-token/main.go`.  The auth token will be printed to standard output.

```
$ curl -D - -k  -X POST http://localhost:3000/token -d '{}'
HTTP/1.1 200 OK
Content-Type: application/json; charset=utf-8
Date: Tue, 17 Jan 2023 18:09:30 GMT
Content-Length: 131

{"status":200,"token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2NzQwNjUzNzB9.mwoJLkw0eKZhRgyX2BX6RwzE1XDzb1b3VDWh-dW2AXk"}
```
