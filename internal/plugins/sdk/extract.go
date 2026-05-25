//go:build extract

package sdk

// This file exists as a marker for the SDK extraction script.
// The build tag "extract" ensures this file is excluded from normal compilation.
// The extract-sdk.ps1 script copies all .go files from internal/plugins/sdk/
// to plugins/sdk/ and generates an independent go.mod for external plugin authors.
