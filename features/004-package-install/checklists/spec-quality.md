# Spec Quality Checklist: Package Install (Dist Strategy)

**Purpose**: Comprehensive requirements quality validation across all 23 FRs, user stories, edge cases, assumptions, and success criteria. Tests whether requirements are complete, clear, consistent, and ready for implementation planning.
**Created**: 2026-03-09
**Feature**: [spec.md](../spec.md)

## Requirement Consistency

- [ ] CHK001 FR-013 references `.aipkg/agent-instructions.md` as a merge target file. The project initialization spec (003) defines `.aipkg/agent-instructions/` as a categorized directory. Are the directory and the merged file two separate concepts, or does one replace the other? [Conflict, Spec FR-013 vs 003 spec]
- [ ] CHK002 FR-009 validates archive structure (valid zip, single top-level directory, contains `aipkg.json`) but does not mention validating the `artifacts` array. FR-013 depends on reading that array. What happens if `aipkg.json` exists but has no `artifacts` field? [Consistency, Spec FR-009 vs FR-013]
- [ ] CHK003 The archive format spec (002) describes extracting to a single directory. This spec extracts individual artifacts to multiple categorized directories. Is the extraction model in FR-013 described as building on the archive format, or does it redefine extraction? [Consistency, Spec FR-013 vs Assumptions]
- [ ] CHK004 SC-006 says "completes without re-downloading" but the skip mechanism depends on the `require` field (FR-015), while FR-012 independently skips cached downloads. Are the two skip mechanisms (project-file-based and cache-based) clearly distinguished in the spec? [Consistency, SC-006 vs FR-012 vs FR-015]

## Requirement Clarity

- [ ] CHK005 FR-012 says archives "MUST be named using a scheme that includes scope, package name, and version" but does not define the exact naming scheme. Is the cache filename format specified sufficiently for two implementers to produce the same result? [Clarity, Spec FR-012]
- [ ] CHK006 FR-005 says "choose the highest semver key from the `versions` map." Are semver comparison rules specified or referenced? For example, is `1.0.0-beta.1` considered lower than `1.0.0`? [Clarity, Spec FR-005]
- [ ] CHK007 FR-004 accepts `AIPKG_REGISTRY` as a URL override but does not define URL normalization rules. Is behavior defined for trailing slashes, query parameters, or path components? [Clarity, Spec FR-004]
- [ ] CHK008 FR-013 mentions "keyed by the artifact's original name" for merged files but does not define the merge format. Is the structure of `.aipkg/mcp.json` (e.g., object with artifact name keys mapping to server configs) specified? [Clarity, Spec FR-013]
- [ ] CHK009 FR-014 says "the package's contributions to merged files are removed by key" but does not define how contributions are identified within merged files (markers, keys, delimiters). Is the de-merge strategy specified? [Clarity, Spec FR-014]
- [ ] CHK010 Is the `@scope/name@version` parsing rule defined? Both the scope and version use `@` as a prefix. How does the CLI distinguish between the scope separator and the version separator? [Clarity, Gap]

## Requirement Completeness

- [ ] CHK011 FR-016 says categorized subdirectories are created "on demand" but does not enumerate which directories. Is the full list of categorized directories specified, or is it implied from the artifact type table? [Completeness, Spec FR-016]
- [ ] CHK012 Are error message requirements specified for artifact placement failures (e.g., merge conflict in mcp.json, write permission denied on categorized directory)? [Completeness, Gap]
- [ ] CHK013 FR-022 says the CLI must "preserve existing content" when writing `aipkg-project.json`. Does this include preserving fields not known to the current CLI version (forward compatibility)? [Completeness, Spec FR-022]
- [ ] CHK014 FR-013 specifies collision detection across packages but does not define how the CLI discovers which artifacts belong to already-installed packages. Is the mechanism for enumerating installed packages' artifacts specified (e.g., reading all cached manifests for packages listed in `require`)? [Completeness, Spec FR-013]

## Scenario Coverage

- [ ] CHK015 FR-015 determines "already installed" from the `require` field in `aipkg-project.json`. If artifacts exist on disk from a failed previous install but the `require` field was never updated, could `aipkg install` skip a package that's partially installed? Is the relationship between on-disk state and project file state defined? [Coverage, Spec FR-015 vs FR-019]
- [ ] CHK016 FR-012 stores archives in a global cache. If two projects use different registries (one HTTPS production, one HTTP local test), could a cached archive from the test registry be used for a production project? Is cache isolation per registry defined? [Coverage, Spec FR-012 vs FR-010]
- [ ] CHK017 Edge case for partial failure (FR-019) says "packages that were successfully installed before the failure remain installed." If a failure occurs mid-placement (e.g., after placing 2 of 4 artifacts from one package), what is the expected state? Is atomicity of single-package installation defined? [Coverage, Spec FR-019]
- [ ] CHK018 Are requirements defined for what happens when `aipkg require` is run outside of any project directory (no `aipkg-project.json` in any parent)? FR-020 checks the current directory only. Is parent directory traversal explicitly excluded? [Coverage, Edge Cases vs FR-020]
- [ ] CHK019 Are requirements defined for concurrent access? If two terminals run `aipkg require` simultaneously on the same project, what happens to `aipkg-project.json` and the `.aipkg/` directory? [Coverage, Gap]
- [ ] CHK020 FR-014 identifies previous artifacts by reading the cached archive's manifest. What happens if the cached archive for the previous version has been manually deleted from `~/.aipkg/cache/`? Is there a fallback mechanism, or does the upgrade fail? [Coverage, Spec FR-014 vs FR-012]

## Acceptance Criteria Quality

- [ ] CHK021 US1 AS2 says "mergeable artifacts (mcp-server, agent-instructions) are merged into their respective files" but does not define what the merged output looks like or how to verify correctness. Is the merge result specified as an acceptance criterion? [Measurability, Spec US1-AS2]
- [ ] CHK022 US1 AS4 says "the old artifacts are removed, the new version's artifacts are placed." Is there a way to verify completeness of removal (i.e., no orphaned artifacts from the old version remain)? [Measurability, Spec US1-AS4]
- [ ] CHK023 SC-001 says "the package is immediately available on disk." "Available" means artifacts in categorized directories. Is "immediately available" defined precisely enough to test? [Measurability, SC-001]

## Dependencies & Assumptions

- [ ] CHK024 The spec assumes `aipkg-project.json` has `specVersion` and `require` fields (003 spec). Has the 003 implementation been verified to produce this exact format, including an empty `require` object (not absent)? [Assumption, Spec Assumptions]
- [ ] CHK025 The spec assumes "Archives contain a single top-level directory, and extraction strips it." FR-013's artifact placement reads the `artifacts` array and places each artifact individually. Does the archive format spec (002) guarantee that the `artifacts` array paths correspond to the archive's internal structure? [Assumption, Spec Assumptions vs FR-013]
- [ ] CHK026 FR-012 introduces `~/.aipkg/cache/` as a global directory. Are cross-platform path considerations addressed (e.g., Windows `%USERPROFILE%`, XDG_CACHE_HOME on Linux)? [Assumption, Spec FR-012]

## Notes

- The existing `requirements.md` checklist in this directory is outdated (references `.aipkg/packages/` and bundled dependencies from the pre-revision spec).
- CHK014 and CHK020 are new items added after dropping three-segment naming. Without prefix-based identification, artifact ownership depends entirely on cached manifests. The reliability of that mechanism under edge cases (deleted cache, corrupted cache) needs consideration.
