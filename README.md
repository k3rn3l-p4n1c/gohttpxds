# gohttpxds

An Http Client for Go that support xDS. This project is inspired by xDS implementation in gRPC and enable proxyless service mesh with Istio and TrafficDirector.

## Why gohttpxds?

- Save a lot of resources by removing proxies like envoy from you service mesh.
- If you are already using proxyless gRPC load balancing but you also have http based APIs, this gohttpxds can homogenize your architecture.

### What is xDS?

The xDS protocol is an open-source data streaming protocol used by service meshes and other cloud-native technologies. It is used to stream configuration data and telemetry data between different components of the system. xDS is designed to be extensible and support various types of data, such as configuration and telemetry. It enables dynamic configuration of the service mesh and helps to maintain the desired state of the system by providing real-time feedback about the system’s state. The xDS protocol is used by service mesh projects such as Istio.

### Inspired by gRPC

Proxyless gRPC load balancing is a technique for distributing load across multiple instances of a gRPC service without using an external proxy. It is an efficient way to scale out gRPC services, as it eliminates the need for extra components such as a service mesh or an API gateway. With proxyless gRPC load balancing, the client will send requests directly to the server and will be able to handle responses from multiple instances of the service. The load balancing is done using the xDS protocol, which provides an efficient and extensible way to configure and manage the load balancing strategy. By using this protocol, the client can dynamically adjust the load balancing strategy based on the current state of the system.

## How to use

``` Go
package main

import (
    "log"
    "net/http"

    "github.com/k3rn3l-p4n1c/gohttpxds"

    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
)

func main() {
    gohttpxds.Register("<host:port to you xDS server like Istio or TrafficDirector>", grpc.WithTransportCredentials(insecure.NewCredentials()), "<node Id>")

    resp, err := http.Get("xds://service/path")
    if err != nil {
        log.Fatal(err.Error())
    }
    // ...
}

```
