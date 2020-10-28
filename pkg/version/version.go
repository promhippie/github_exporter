package version

import (
	"runtime"
)

var (
	// String gets defined by the build system.
	String = "0.0.0-dev"

	// Revision indicates the commit this binary was built from.
	Revision string

	// Date indicates the date this binary was built.
	Date string

	// Go running this binary.
	Go = runtime.Version()
)
