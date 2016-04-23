marathon-netscaler
------------------

Interact with [NetScaler](https://www.citrix.com/products/netscaler-application-delivery-controller/overview.html) on behalf of [Marathon](https://mesosphere.github.io/marathon) applications.

## running marathon-netscaler

```
make
bin/marathon-netscaler bin/marathon-netscaler -log.level=debug -marathon.uri=http://marathon:8080 -netscaler.uri=http://netscaler
```

## using marathon-netscaler

Deploy your app to Marathon with a custom label called `netscaler.service_group`.

```
{
  "id": "/example",
  "labels": {
    "netscaler.service_group": "example"
  }
}
```

## todo

This is a work in progress.

- Update NetScaler
- Honor a Marathon label that specifies which app port index to use; currently assumes the first port
- Add a cli argument that supplies a custom label prefix; i.e. `dev_netscaler.service_group` instead of `netscaler.service_group`
- One or more metrics collection thingies
- A cool name
