// Package sdk provides a reusable framework for building Docker plugin containers
// that expose both REST/HTTP and gRPC APIs.
//
// Usage:
//
//	server, _ := sdk.NewPlugin()
//	server.AddCapability(&contract.Capability{...})
//	server.HandleFunc("/api/v1/custom", myHandler)
//	server.Start(ctx)
package sdk
