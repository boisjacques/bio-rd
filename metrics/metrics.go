package metrics

var prefix = "bio"

// OverridePrefix overrides the prefix used by FullName to get unified metric names for one application (when using BIO as library)
func OverridePrefix(p string) {
	prefix = p
}

// FullName creates an fully qualified metric name
func FullName(name string) string {
	return prefix + "/" + name
}
