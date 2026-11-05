//go:build phpx

package files_test

// phpxCSVForeignEnclosure is what a script prints when it names an enclosure
// other than the double quote: phpx writes and reads it, as php does.
const phpxCSVForeignEnclosure = "a\nwritten|read"
