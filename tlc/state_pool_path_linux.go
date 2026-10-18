//go:build linux

package tlc

func statePoolCanonicalPath(path string) (string, error) {
	return javaIOCanonicalPath(path)
}
