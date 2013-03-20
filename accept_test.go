// Copyright 2013 Ryan Rogers. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package accept

import (
	"testing"
)

func TestParse(t *testing.T) {
	type parseTest struct {
		input  string
		output AcceptSlice
	}

	parseTests := []parseTest{
		{ // 0
			// Empty/not sent header signals that everything is accepted.
			input: "",
			output: AcceptSlice{
				{ // 0
					Type:       "*",
					Subtype:    "*",
					Q:          1,
					Extensions: map[string]string{},
				},
			},
		},
		{ // 1
			// Chrome is currently sending this.
			input: "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			output: AcceptSlice{
				{ // 0
					Type:       "text",
					Subtype:    "html",
					Q:          1,
					Extensions: map[string]string{},
				},
				{ // 1
					Type:       "application",
					Subtype:    "xhtml+xml",
					Q:          1,
					Extensions: map[string]string{},
				},
				{ // 2
					Type:       "application",
					Subtype:    "xml",
					Q:          0.9,
					Extensions: map[string]string{},
				},
				{ // 3
					Type:       "*",
					Subtype:    "*",
					Q:          0.8,
					Extensions: map[string]string{},
				},
			},
		},
		{ // 2
			// Same as 1, except with crazy whitespacing.
			input: `text  /  html  ,	application	/	xhtml+xml	,
					application
					/
					xml
					;
					q
					=
					0.9
					,  *  /  *  ;  q  =  0.8`,
			output: AcceptSlice{
				{ // 0
					Type:       "text",
					Subtype:    "html",
					Q:          1,
					Extensions: map[string]string{},
				},
				{ // 1
					Type:       "application",
					Subtype:    "xhtml+xml",
					Q:          1,
					Extensions: map[string]string{},
				},
				{ // 2
					Type:       "application",
					Subtype:    "xml",
					Q:          0.9,
					Extensions: map[string]string{},
				},
				{ // 3
					Type:       "*",
					Subtype:    "*",
					Q:          0.8,
					Extensions: map[string]string{},
				},
			},
		},
		{ // 3
			// Same as 1, except with modified/invalid qvals.
			input: "text/html;q=1.05,application/xhtml+xml;q=-1.05,application/xml;q=1.0=0.5,*/*;q=INVALID",
			output: AcceptSlice{
				{ // 0
					Type:       "text",
					Subtype:    "html",
					Q:          1,
					Extensions: map[string]string{},
				},
			},
		},
		{ // 4
			// Complex ordering of preference.
			input: "*/*,*/*;a=1,*/*;a=1;b=1,text/*,text/*;a=1,text/*;a=1;b=1,*/plain,*/plain;a=1,*/plain;a=1;b=1,text/plain,text/plain;a=1,text/plain;a=1;b=1",
			output: AcceptSlice{
				{ // 0
					Type:    "text",
					Subtype: "plain",
					Q:       1,
					Extensions: map[string]string{
						"a": "1",
						"b": "1",
					},
				},
				{ // 1
					Type:    "text",
					Subtype: "plain",
					Q:       1,
					Extensions: map[string]string{
						"a": "1",
					},
				},
				{ // 2
					Type:       "text",
					Subtype:    "plain",
					Q:          1,
					Extensions: map[string]string{},
				},
				{ // 3
					Type:    "text",
					Subtype: "*",
					Q:       1,
					Extensions: map[string]string{
						"a": "1",
						"b": "1",
					},
				},
				{ // 4
					Type:    "text",
					Subtype: "*",
					Q:       1,
					Extensions: map[string]string{
						"a": "1",
					},
				},
				{ // 5
					Type:       "text",
					Subtype:    "*",
					Q:          1,
					Extensions: map[string]string{},
				},
				{ // 6
					Type:    "*",
					Subtype: "plain",
					Q:       1,
					Extensions: map[string]string{
						"a": "1",
						"b": "1",
					},
				},
				{ // 7
					Type:    "*",
					Subtype: "plain",
					Q:       1,
					Extensions: map[string]string{
						"a": "1",
					},
				},
				{ // 8
					Type:       "*",
					Subtype:    "plain",
					Q:          1,
					Extensions: map[string]string{},
				},
				{ // 9
					Type:    "*",
					Subtype: "*",
					Q:       1,
					Extensions: map[string]string{
						"a": "1",
						"b": "1",
					},
				},
				{ // 10
					Type:    "*",
					Subtype: "*",
					Q:       1,
					Extensions: map[string]string{
						"a": "1",
					},
				},
				{ // 11
					Type:       "*",
					Subtype:    "*",
					Q:          1,
					Extensions: map[string]string{},
				},
			},
		},
	}

	var accepted AcceptSlice
	for testPos, test := range parseTests {
		accepted = Parse(test.input)
		if len(accepted) != len(test.output) {
			t.Errorf("Parse (%d): expected %d elements, received %d.", testPos, len(test.output), len(accepted))
			continue
		}
		for i, a := range accepted {
			if a.Type != test.output[i].Type {
				t.Errorf("Parse (%d.%d): expected type '%v', received '%v'.", testPos, i, test.output[i].Type, a.Type)
			}
			if a.Subtype != test.output[i].Subtype {
				t.Errorf("Parse (%d.%d): expected subtype '%v', received '%v'.", testPos, i, test.output[i].Subtype, a.Subtype)
			}
			if a.Q != test.output[i].Q {
				t.Errorf("Parse (%d.%d): expected qval '%v', received '%v'.", testPos, i, test.output[i].Q, a.Q)
			}
			if !mapsAreSimilar(a.Extensions, test.output[i].Extensions) {
				t.Errorf("Parse (%d.%d): expected extensions '%v', received '%v'.", testPos, i, test.output[i].Extensions, a.Extensions)
			}
		}
	}
}

func TestNegotiate(t *testing.T) {
	type negotiateTest struct {
		header   string
		types    []string
		expected string
	}

	negotiateTests := []negotiateTest{
		{ // 0
			// FIXME: I'm not sure if this behavior makes sense.
			// Can't negotiate when given zero types.
			header:   "application/xml,application/xhtml+xml,text/html;q=0.9,text/plain;q=0.8,image/png,*/*;q=0.5",
			types:    []string{},
			expected: "",
		},
		{ // 1
			// When given an empty header, the first type will match.
			header: "",
			types: []string{
				"application/octet-stream",
				"image/jpeg",
			},
			expected: "application/octet-stream",
		},
		{ // 2
			// application/xml is negotiated due to its position in header.
			header: "application/xml,application/xhtml+xml,text/html;q=0.9,text/plain;q=0.8,image/png,*/*;q=0.5",
			types: []string{
				"text/plain",
				"text/html",
				"application/xhtml+xml",
				"application/xml",
			},
			expected: "application/xml",
		},
		{ // 3
			// text/plain is negotiated due to its position in types.
			header: "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			types: []string{
				"text/plain",
				"image/png",
			},
			expected: "text/plain",
		},
		{ // 4
			// When a type or subtype is omitted, it is negotiated as "*".
			// The expected type is returned exactly as it was passed in.
			header: "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			types: []string{
				"text/",
				"/xml",
			},
			expected: "text/",
		},
		{ // 5
			// */* will always negotiate to itself.
			header: "text/html;q=0.9,text/plain,application/xhtml+xml,application/xml;q=0.9",
			types: []string{
				"*/*",
				"text/plain",
			},
			expected: "*/*",
		},
		{ // 6
			header: "text/html;q=0.9,text/plain,application/xhtml+xml,application/xml;q=0.9",
			types: []string{
				"text/*",
				"text/plain",
			},
			expected: "text/*",
		},
		{ // 7
			header: "text/html;q=0.9,text/plain,application/xhtml+xml,application/xml;q=0.9",
			types: []string{
				"*/xhtml+xml",
				"application/xhtml+xml",
			},
			expected: "*/xhtml+xml",
		},
	}

	for i, test := range negotiateTests {
		result, err := Negotiate(test.header, test.types...)
		if err != nil {
			t.Errorf("Negotiate (%d): expected no error, received '%v'.", i, err)
			continue
		}
		if result != test.expected {
			t.Errorf("Negotiate (%d): expected type '%v', received '%v'.", i, test.expected, result)
		}
	}
}

func TestAccepts(t *testing.T) {
	type acceptsTest struct {
		header string
		types  []string
	}

	acceptsTests_true := []acceptsTest{
		{ // 0
			header: "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			types: []string{
				"text/html",
				"application/xhtml+xml",
				"application/xml",
				"text",
				"image",
				"text/*",
				"image/*",
				"*/html",
				"*/xml",
			},
		},
	}
	acceptsTests_false := []acceptsTest{
		{ // 0
			header: "text/html,application/xhtml+xml,application/xml;q=0.9",
			types: []string{
				"",
				"text/plain",
				"application/octet-stream",
			},
		},
	}
	var accepted AcceptSlice

	for testPos, test := range acceptsTests_true {
		accepted = Parse(test.header)
		for i, ctype := range test.types {
			if !accepted.Accepts(ctype) {
				t.Errorf("Accepts (%d.%d): expected '%v' to be accepted.", testPos, i, ctype)
			}
		}
	}

	for testPos, test := range acceptsTests_false {
		accepted = Parse(test.header)
		for i, ctype := range test.types {
			if accepted.Accepts(ctype) {
				t.Errorf("Accepts (%d.%d): expected '%v' to not be accepted.", testPos, i, ctype)
			}
		}
	}
}

//
// Utility functions
//

func mapsAreSimilar(a, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}
	for aKey, aVal := range a {
		if bVal, exists := b[aKey]; !exists || aVal != bVal {
			return false
		}
	}
	return true
}
