marathon-netscaler
------------------

Interact with [NetScaler](https://www.citrix.com/products/netscaler-application-delivery-controller/overview.html) on behalf of [Marathon](https://mesosphere.github.io/marathon) applications.

## running marathon-netscaler

```
make
bin/marathon-netscaler bin/marathon-netscaler -marathon.uri=http://marathon:8080 -netscaler.uri=http://netscaler
```

## using marathon-netscaler

Deploy your app to Marathon with a custom label called `netscaler.service_group`.

```
{
  "id": "/example",
  "ports": [9090],
  "labels": {
    "netscaler.service_group": "example"
  }
}
```

## todo

This is a work in progress.

- Update NetScaler
- Add a cli argument that supplies a custom label prefix; i.e. `dev_netscaler.service_group` instead of `netscaler.service_group`
- Add a cli argument that supplies a custom NetScaler server-name prefix for Mesos agents; i.e. `dev_marathon-agent.hostname.com` instead of `marathon-agent.hostname.com`
- One or more metrics collection thingies
- A cool name
