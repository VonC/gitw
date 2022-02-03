module gitw

go 1.16

require (
	github.com/charmbracelet/lipgloss v0.4.0
	github.com/erikgeiser/promptkit v0.6.0
	github.com/muesli/termenv v0.11.0
	github.com/romana/rlog v0.0.0-20171115192701-f018bc92e7d7
	golang.org/x/sys v0.0.0-20220128215802-99c3d69c2c27
)

replace github.com/erikgeiser/promptkit => ./deps/promptkit
