module github.com/maccavelli/prepare-commit-msg

go 1.26.5

require golang.org/x/term v0.43.0

require (
	github.com/maccavelli/mcplib v0.2.0
	golang.org/x/sys v0.45.0 // indirect
)

replace github.com/maccavelli/mcplib => ../mcplib
