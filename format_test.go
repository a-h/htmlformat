package htmlformat

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestFormat(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:  "missing closing tags are inserted",
			input: `<li>`,
			expected: `<li>
</li>
`,
		},
		{
			name:  "html attribute escaping is normalized",
			input: `<ol> <li style="&amp;&#38;"> A </li> <li> B </li> </ol> `,
			expected: `<ol>
 <li style="&amp;&amp;">
  A
 </li>
 <li>
  B
 </li>
</ol>
`,
		},
		{
			name:  "bare ampersands are escaped",
			input: `<ol> <li style="&"> A </li> <li> B </li> </ol> `,
			expected: `<ol>
 <li style="&amp;">
  A
 </li>
 <li>
  B
 </li>
</ol>
`,
		},
		{
			name:  "html elements are indented",
			input: `<ol> <li class="name"> A </li> <li> B </li> </ol> `,
			expected: `<ol>
 <li class="name">
  A
 </li>
 <li>
  B
 </li>
</ol>
`,
		},
		{
			name:     "text fragments are supported",
			input:    `test 123`,
			expected: `test 123` + "\n",
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			r := strings.NewReader(test.input)
			w := new(strings.Builder)
			if err := Fragment(w, r); err != nil {
				t.Fatalf("failed to format: %v", err)
			}
			if diff := cmp.Diff(test.expected, w.String()); diff != "" {
				t.Error(diff)
			}
		})
	}
}
