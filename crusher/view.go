package crusher

import (
	"fmt"
	"strings"
)

// HexDumpColored returns a tview-colored hex dump string of the data.
// It limits the output to maxLines to prevent UI freezing on large files.
// headerOffset is used to highlight the locked header part.
func HexDumpColored(data []byte, headerOffset int, maxLines int) string {
	var builder strings.Builder

	bytesPerLine := 16
	lines := len(data) / bytesPerLine
	if len(data)%bytesPerLine != 0 {
		lines++
	}

	if lines > maxLines {
		lines = maxLines
	}

	for line := 0; line < lines; line++ {
		start := line * bytesPerLine
		end := start + bytesPerLine
		if end > len(data) {
			end = len(data)
		}

		// Write address
		builder.WriteString(fmt.Sprintf("[yellow]%08x[white]  ", start))

		// Write hex values
		for i := start; i < start+bytesPerLine; i++ {
			if i < len(data) {
				color := "[green]" // default body color
				if i < headerOffset {
					color = "[gray]" // locked header color
				}
				builder.WriteString(fmt.Sprintf("%s%02x ", color, data[i]))
			} else {
				builder.WriteString("   ")
			}
			if (i-start)%bytesPerLine == 7 {
				builder.WriteString(" ")
			}
		}

		builder.WriteString(" [white]|")

		// Write ascii values
		for i := start; i < end; i++ {
			c := data[i]
			color := "[green]"
			if i < headerOffset {
				color = "[gray]"
			}
			if c >= 32 && c <= 126 {
				builder.WriteString(fmt.Sprintf("%s%c", color, c))
			} else {
				builder.WriteString(fmt.Sprintf("%s.", color))
			}
		}
		builder.WriteString("[white]|\n")
	}

	if len(data)/bytesPerLine > maxLines {
		builder.WriteString(fmt.Sprintf("\n... [red]%d bytes hidden[white] ...\n", len(data)-(maxLines*bytesPerLine)))
	}

	return builder.String()
}
