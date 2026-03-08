package aipkg

import _ "embed"

// PackageSchemaJSON is the compiled package manifest JSON Schema.
//
//go:embed spec/schema/aipkg.json
var PackageSchemaJSON []byte

// ProjectSchemaJSON is the compiled project file JSON Schema.
//
//go:embed spec/schema/project.json
var ProjectSchemaJSON []byte

// ReservedScopesText is the list of reserved scope names.
//
//go:embed spec/reserved-scopes.txt
var ReservedScopesText []byte
