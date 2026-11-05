package tests

import (
	"testing"
)

// TestPhpxBuildMatchesTag pins that the constant follows the build tag by
// checking that a fixture marked phpx runs only where the constant is true.
func TestPhpxBuildMatchesTag(t *testing.T) {
	t.Parallel()
	areas, err := embeddedFixtures()
	if err != nil {
		t.Fatal(err)
	}
	for _, area := range areas {
		for _, fx := range area.Fixtures {
			if fx.Phpx && !PhpxBuild {
				t.Errorf("%s needs phpx and loaded in a build without it", fx.Path)
			}
		}
	}
}
