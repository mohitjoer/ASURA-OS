// The control tools. The hardware detection libraries live in their own module
// under system/hardware because they are self-contained and separately
// testable; this module is what a user actually runs.
//
// The replace directive points at that module in-tree rather than resolving it
// from a registry. There is no published module to resolve, and a build that
// needed one could not be reproduced from a checkout.
module toren/desktop

go 1.26.0

require (
	github.com/BurntSushi/toml v1.6.0
	github.com/godbus/dbus/v5 v5.2.2
	golang.org/x/sys v0.49.0
	toren/hardware v0.0.0
)

require (
	github.com/dustin/go-humanize v1.1.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/mattn/go-isatty v0.0.24 // indirect
	github.com/ncruces/go-strftime v1.1.0 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20230129092748-24d4a6f8daec // indirect
	modernc.org/libc v1.80.1 // indirect
	modernc.org/mathutil v1.7.1 // indirect
	modernc.org/memory v1.12.1 // indirect
	modernc.org/sqlite v1.63.0 // indirect
)

replace toren/hardware => ../system/hardware
