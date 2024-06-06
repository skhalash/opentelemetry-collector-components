//go:build tools

package tools // import "github.com/kyma-project/opentelemetry-collector-components/internal/tools"

// This file follows the recommendation at
// https://go.dev/wiki/Modules#how-can-i-track-tool-dependencies-for-a-module
// on how to pin tooling dependencies to a go.mod file.
// This ensures that all systems use the same version of tools in addition to regular dependencies.
import (
	_ "github.com/golangci/golangci-lint/cmd/golangci-lint"
	_ "go.opentelemetry.io/build-tools/crosslink"
	_ "go.opentelemetry.io/collector/cmd/mdatagen"
)