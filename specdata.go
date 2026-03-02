package aipkg

import _ "embed"

// PackageSchemaJSON is the compiled package manifest JSON Schema.
//
//go:embed spec/schema/package.json
var PackageSchemaJSON []byte

// ReservedScopesText is the list of reserved scope names.
//
//go:embed spec/reserved-scopes.txt
var ReservedScopesText []byte
