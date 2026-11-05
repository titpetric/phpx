//go:build phpx

package parser

import "testing"

// TestAlternativeSyntax pins that the colon forms of if, while, for, foreach
// and switch parse.
func TestAlternativeSyntax(t *testing.T) {
	t.Parallel()
	src := `<?php
if ($a): echo 1; elseif ($b): echo 2; else: echo 3; endif;
while ($a): $a--; endwhile;
for ($i = 0; $i < 3; $i++): echo $i; endfor;
foreach ($xs as $x): echo $x; endforeach;
switch ($a): case 1: echo 1; break; default: echo 2; endswitch;
`
	if _, err := Parse(src); err != nil {
		t.Fatal(err)
	}
}
