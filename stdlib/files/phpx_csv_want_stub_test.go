//go:build !phpx

package files_test

// phpxCSVForeignEnclosure is what a script prints when it names an enclosure
// other than the double quote: upstream refuses both calls.
const phpxCSVForeignEnclosure = "refused|refused"
