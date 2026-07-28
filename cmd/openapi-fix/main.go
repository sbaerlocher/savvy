// Package main post-processes the spec that `swag init` writes into
// docs/openapi/.
//
// swag v2.0.0-rc5 emits every `@Param … body` as
//
//	oneOf: [{type: object}, {$ref: <DTO>}]
//
// which is unsatisfiable: the DTO schemas declare neither `required` nor
// `additionalProperties: false`, so any JSON object matches both branches
// while `oneOf` demands exactly one. No request body can validate, and code
// generators emit an anonymous union instead of the named DTO. This tool
// collapses the wrapper down to the `$ref` branch that was intended.
//
// Both edits are textual rather than a parse-and-reserialize. docs.go carries
// Go template actions (`{{ marshal .Schemes }}`) inside its JSON string, so it
// is not valid JSON to begin with; and re-encoding the YAML reflows every
// untouched line (string wrapping and sequence indentation differ from swag's
// own writer), which would bury five real changes under a thousand cosmetic
// ones.
//
// Run via `make openapi`, never by hand — it rewrites generated files.
package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

const (
	yamlPath = "docs/openapi/openapi.yaml"
	docsPath = "docs/openapi/docs.go"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "openapi-fix: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	yamlFixes, err := rewrite(yamlPath, collapseYAML)
	if err != nil {
		return err
	}

	docsFixes, err := rewrite(docsPath, collapseDocsGo)
	if err != nil {
		return err
	}

	// The two artifacts describe the same API, so a divergence means one of the
	// patterns stopped matching — most likely because a swag upgrade changed
	// the emitted shape. Failing here beats shipping a half-fixed spec.
	if yamlFixes != docsFixes {
		return fmt.Errorf("collapsed %d oneOf wrappers in %s but %d in %s: the two artifacts disagree",
			yamlFixes, yamlPath, docsFixes, docsPath)
	}

	if yamlFixes == 0 {
		return fmt.Errorf("no oneOf request-body wrappers found: did swag stop emitting them? drop this tool if so")
	}

	fmt.Printf("✓ collapsed %d oneOf request-body wrappers\n", yamlFixes)

	return nil
}

// rewrite applies fn to the file at path, writing it back only when fn changed
// something. Returns the number of replacements fn reported.
func rewrite(path string, fn func(string) (string, int)) (int, error) {
	raw, err := os.ReadFile(path) //nolint:gosec // fixed generated path
	if err != nil {
		return 0, err
	}

	out, count := fn(string(raw))
	if count == 0 {
		return 0, nil
	}

	if err := os.WriteFile(path, []byte(out), 0o600); err != nil {
		return 0, err
	}

	return count, nil
}

// collapseYAML replaces every wrapper with its `$ref` branch, re-indented to
// where the `oneOf` key stood. swag's shape:
//
//	schema:
//	  oneOf:
//	  - type: object
//	  - $ref: '#/components/schemas/api.CardCreateRequest'
//	    description: Card to create
//	    summary: card
//
// The dashes sit at the same indentation as the `oneOf` key, so the branch body
// is two columns further in. This runs line-by-line rather than as one regex
// because the branch ends at the first line indented no deeper than its dash —
// a bound a regex cannot express, and getting it wrong swallows the following
// sibling keys into the branch.
func collapseYAML(src string) (string, int) {
	lines := strings.Split(src, "\n")
	out := make([]string, 0, len(lines))
	count := 0

	for i := 0; i < len(lines); i++ {
		branch, next, ok := matchWrapper(lines, i)
		if !ok {
			out = append(out, lines[i])

			continue
		}

		out = append(out, branch...)
		count++
		i = next - 1
	}

	return strings.Join(out, "\n"), count
}

// wrapperHead matches the `oneOf:` key and the `- type: object` branch that
// together mark swag's request-body wrapper. Anchoring on that bare first
// branch keeps a hand-written oneOf elsewhere in the spec untouched.
var (
	wrapperHead = regexp.MustCompile(`^([ \t]*)oneOf:$`)
	bareObject  = regexp.MustCompile(`^([ \t]*)- type: object$`)
	refBranch   = regexp.MustCompile(`^([ \t]*)- (\$ref:.*)$`)
)

// matchWrapper reports whether a wrapper starts at lines[i]. On a match it
// returns the collapsed branch re-indented to the `oneOf` key, plus the index
// of the first line after the wrapper.
func matchWrapper(lines []string, i int) (branch []string, next int, ok bool) {
	head := wrapperHead.FindStringSubmatch(lines[i])
	if head == nil || i+2 >= len(lines) {
		return nil, 0, false
	}

	if !bareObject.MatchString(lines[i+1]) {
		return nil, 0, false
	}

	ref := refBranch.FindStringSubmatch(lines[i+2])
	if ref == nil {
		return nil, 0, false
	}

	indent := head[1]
	branch = []string{indent + ref[2]}

	// Continuation lines belong to the branch only while they are indented
	// deeper than the dash that introduced it.
	dashIndent := len(ref[1])

	next = i + 3
	for ; next < len(lines); next++ {
		trimmed := strings.TrimLeft(lines[next], " \t")
		if trimmed == "" {
			break
		}

		if len(lines[next])-len(trimmed) <= dashIndent {
			break
		}

		branch = append(branch, indent+trimmed)
	}

	return branch, next, true
}

// docsGoWrapper matches the generated JSON form of the same wrapper, capturing
// the indentation of the `"oneOf"` key and the body of the `$ref` branch.
// `(?s)` lets `.` span the newlines between the branches; the JSON braces bound
// the branch, so no indentation-aware scan is needed here.
var docsGoWrapper = regexp.MustCompile(
	`(?s)([ \t]*)"oneOf": \[\s*\{\s*"type": "object"\s*\},\s*\{(.*?)\}\s*\]`)

// collapseDocsGo applies the same collapse to the JSON embedded in docs.go,
// re-indenting the branch to where the `"oneOf"` key stood.
func collapseDocsGo(src string) (string, int) {
	count := 0

	out := docsGoWrapper.ReplaceAllStringFunc(src, func(match string) string {
		groups := docsGoWrapper.FindStringSubmatch(match)
		indent, inner := groups[1], groups[2]

		if !strings.Contains(inner, `"$ref"`) {
			return match
		}

		count++

		body := make([]string, 0, 3)

		for _, line := range strings.Split(inner, "\n") {
			trimmed := strings.TrimSpace(line)
			if trimmed == "" {
				continue
			}

			body = append(body, indent+trimmed)
		}

		return strings.Join(body, "\n")
	})

	return out, count
}
