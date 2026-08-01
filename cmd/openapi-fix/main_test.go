package main

import "testing"

func TestCollapseYAML(t *testing.T) {
	// Indentation and key order copied from swag's actual output.
	src := `paths:
  /cards:
    post:
      requestBody:
        content:
          application/json:
            schema:
              oneOf:
              - type: object
              - $ref: '#/components/schemas/api.CardCreateRequest'
                description: Card to create
                summary: card
        description: Card to create
        required: true
`
	want := `paths:
  /cards:
    post:
      requestBody:
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/api.CardCreateRequest'
              description: Card to create
              summary: card
        description: Card to create
        required: true
`

	got, count := collapseYAML(src)
	if count != 1 {
		t.Fatalf("count = %d, want 1", count)
	}

	if got != want {
		t.Errorf("collapseYAML mismatch\n got:\n%s\nwant:\n%s", got, want)
	}
}

// A oneOf that is not swag's request-body wrapper must survive untouched,
// otherwise the tool would silently destroy a real union schema.
func TestCollapseYAMLLeavesGenuineUnionAlone(t *testing.T) {
	src := `schema:
  oneOf:
  - $ref: '#/components/schemas/api.CardDTO'
  - $ref: '#/components/schemas/api.VoucherDTO'
`

	got, count := collapseYAML(src)
	if count != 0 {
		t.Fatalf("count = %d, want 0", count)
	}

	if got != src {
		t.Errorf("input was modified:\n%s", got)
	}
}

func TestCollapseDocsGo(t *testing.T) {
	src := `                        "schema": {
                            "oneOf": [
                                {
                                    "type": "object"
                                },
                                {
                                    "$ref": "#/components/schemas/api.CardCreateRequest",
                                    "summary": "card"
                                }
                            ]
                        }`
	want := `                        "schema": {
                            "$ref": "#/components/schemas/api.CardCreateRequest",
                            "summary": "card"
                        }`

	got, count := collapseDocsGo(src)
	if count != 1 {
		t.Fatalf("count = %d, want 1", count)
	}

	if got != want {
		t.Errorf("collapseDocsGo mismatch\n got:\n%s\nwant:\n%s", got, want)
	}
}
