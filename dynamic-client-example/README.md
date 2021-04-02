* Follow this [tutorial](https://github.com/kubernetes/client-go/tree/master/examples/dynamic-create-update-delete-deployment)

# CRUD (Create, Read, Update, Delete) of k8s object Deployment resources using Dynamic Package

- With this example we will demonstrate the basic CRUD operations on `Deployment` resources using client-go's `dynamic` package, which is also knows as `dynamic-client`

## Typed

- The code is this directory is based on similar [client-go example](https://github.com/kubernetes/client-go/tree/master/examples/create-update-delete-deployment)
- The typed client sets make it simple to communicate with the API server using pre-generated local API objects to achieve an RPC-like programming experience. 
- However, when using typed clients, programs are forced to be tightly coupled with the version and the types used.

## Dynamic

- The `dynamic` package on the other hand, uses a simple type, `unstructured.Unstructured`, to represent all object values from the API server.

