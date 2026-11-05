package runner

// NamespaceLate names the installers RegisterFS runs after the filesystem
// functions are rooted: RegisterFS registers them again against the script's
// directory and would replace a binding that overrode one earlier.
const NamespaceLate = NamespacePX + `\Late`
