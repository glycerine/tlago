package tlc

import "strings"

// ExtractDistributedStartupProperties consumes the properties which Java's
// launcher applies before entering a distributed main method. Validate the
// entire argument list before changing process-wide startup settings.
func ExtractDistributedStartupProperties(args []string) ([]string, error) {
	remaining := make([]string, 0, len(args))
	type property struct{ name, value string }
	var properties []property
	for _, arg := range args {
		if !strings.HasPrefix(arg, "-D") {
			remaining = append(remaining, arg)
			continue
		}
		name, value, err := parseTLCSystemPropertyOption(arg)
		if err != nil {
			return nil, err
		}
		properties = append(properties, property{name, value})
	}
	for _, p := range properties {
		tlcSetStartupSystemProperty(p.name, p.value)
	}
	return remaining, nil
}

// DistributedSystemProperty reads the same property store used by the original
// distributed option and startup ports, including environment fallbacks.
func DistributedSystemProperty(name, fallback string) string {
	if value, ok := tlcLookupSystemProperty(name); ok {
		return value
	}
	return fallback
}
