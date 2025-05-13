package internal

// ANSI color codes for terminal output
const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorPurple = "\033[35m"
	ColorCyan   = "\033[36m"
	ColorWhite  = "\033[37m"
	ColorGray   = "\033[90m"
	ColorBold   = "\033[1m"
)

// ColoredMessage returns string wrapped in ANSI color codes
func ColoredMessage(color string, message string) string {
	return color + message + ColorReset
}

// Success returns a green colored message
func Success(message string) string {
	return ColoredMessage(ColorGreen, message)
}

// Error returns a red colored message
func Error(message string) string {
	return ColoredMessage(ColorRed, message)
}

// Warning returns a yellow colored message
func Warning(message string) string {
	return ColoredMessage(ColorYellow, message)
}

// Info returns a cyan colored message
func Info(message string) string {
	return ColoredMessage(ColorCyan, message)
}

// Highlight returns a purple colored message
func Highlight(message string) string {
	return ColoredMessage(ColorPurple, message)
}

// Gray returns a gray colored message
func Gray(message string) string {
	return ColoredMessage(ColorGray, message)
}

// Bold returns a bold message
func Bold(message string) string {
	return ColoredMessage(ColorBold, message)
}

// PrintDivider prints a colorful divider to separate console output sections
func PrintDivider() {
	divider := "========================================================"
	println(ColoredMessage(ColorBlue, divider))
}

// PrintSectionDivider prints a section divider with a title
func PrintSectionDivider(title string) {
	divider := "========================================================"
	println("\n" + ColoredMessage(ColorBlue, divider))
	println(Bold(ColoredMessage(ColorBlue, "  " + title)))
	println(ColoredMessage(ColorBlue, divider))
} 