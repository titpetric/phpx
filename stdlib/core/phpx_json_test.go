//go:build phpx

package core_test

import "testing"

// TestPhpxJSONDecodeAnswersStdClass covers PHP's default shape: an object in
// the document decodes into an stdClass instance and reads with `->name`, at
// any depth, and only `$associative = true` asks for an array instead. The
// registration lives in phpx_json.go and replaces upstream's decoder, which
// builds an array and refuses false.
func TestPhpxJSONDecodeAnswersStdClass(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		src  string
		want string
	}{
		{
			name: "default",
			src:  `<?php $o = json_decode('{"a":1}'); echo get_class($o), " ", $o->a;`,
			want: "stdClass 1",
		},
		{
			name: "false",
			src:  `<?php $o = json_decode('{"a":1}', false); echo get_class($o), " ", $o->a;`,
			want: "stdClass 1",
		},
		{
			name: "null",
			src:  `<?php $o = json_decode('{"a":1}', null); echo get_class($o), " ", $o->a;`,
			want: "stdClass 1",
		},
		{
			name: "associative",
			src:  `<?php $a = json_decode('{"a":1}', true); echo gettype($a), " ", $a["a"];`,
			want: "array 1",
		},
		{
			name: "nested",
			src:  `<?php $o = json_decode('{"a":{"b":2}}'); echo get_class($o->a), " ", $o->a->b;`,
			want: "stdClass 2",
		},
		{
			name: "nested associative",
			src:  `<?php $a = json_decode('{"a":{"b":2}}', true); echo gettype($a["a"]), " ", $a["a"]["b"];`,
			want: "array 2",
		},
		{
			name: "list of objects",
			src:  `<?php $rows = json_decode('[{"c":3}]'); echo get_class($rows[0]), " ", $rows[0]->c;`,
			want: "stdClass 3",
		},
		{
			name: "empty object",
			src:  `<?php $o = json_decode('{}'); echo get_class($o), " ", count(get_object_vars($o));`,
			want: "stdClass 0",
		},
		{
			name: "document order",
			src:  `<?php echo json_encode(json_decode('{"z":1,"a":{"q":2,"b":3}}'));`,
			want: `{"z":1,"a":{"q":2,"b":3}}`,
		},
		{
			name: "scalars",
			src:  `<?php var_dump(json_decode('true'), json_decode('null'), json_decode('7'), json_decode('"s"'));`,
			want: "bool(true)\nNULL\nint(7)\nstring(1) \"s\"\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := runPHP(t, tt.src); got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}
