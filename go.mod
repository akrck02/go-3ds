module github.com/akrck02/go-3ds

go 1.25.6

require (
	github.com/akrck02/go-renderer v0.0.0
	github.com/akrck02/go-ui v0.0.0-20261019173702-0c53a25e3cf4
)

require (
	golang.org/x/image v0.45.0 // indirect
	golang.org/x/sys v0.47.0 // indirect
	golang.org/x/text v0.41.0 // indirect
)

// go-ui asks for go-renderer v0.0.0 (it builds against a local copy); this is go-renderer's
// dev branch, which has ctr and everything waiting for a release.
replace github.com/akrck02/go-renderer => github.com/akrck02/go-renderer v0.0.0-20261019173110-5a5d4410828b
