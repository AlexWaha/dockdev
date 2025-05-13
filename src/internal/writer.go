package internal

import (
	"bufio"
	"bytes"
	"io"
)

// FilterFunc is a function that determines whether a line should be written
type FilterFunc func(line string) bool

// FilteredWriter is a writer that can filter lines based on their content
type FilteredWriter struct {
	writer io.Writer
	filter FilterFunc
	buffer bytes.Buffer
}

// NewFilteredWriter creates a new filtered writer
func NewFilteredWriter(writer io.Writer, filter FilterFunc) *FilteredWriter {
	return &FilteredWriter{
		writer: writer,
		filter: filter,
	}
}

// Write implements the io.Writer interface
func (fw *FilteredWriter) Write(p []byte) (n int, err error) {
	// Add the new bytes to our buffer
	fw.buffer.Write(p)
	
	// Check if we have complete lines in the buffer
	scanner := bufio.NewScanner(bytes.NewReader(fw.buffer.Bytes()))
	var lines []string
	var remainingData []byte
	
	// Find complete lines
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	
	// If the buffer doesn't end with a newline, save the incomplete line
	if len(p) > 0 && p[len(p)-1] != '\n' {
		if len(lines) > 0 {
			// Remove the last line from complete lines
			remainingData = []byte(lines[len(lines)-1])
			lines = lines[:len(lines)-1]
		} else {
			// The entire buffer is an incomplete line
			remainingData = fw.buffer.Bytes()
		}
	}
	
	// Reset the buffer and add back any incomplete line
	fw.buffer.Reset()
	if len(remainingData) > 0 {
		fw.buffer.Write(remainingData)
	}
	
	// Write filtered lines to the underlying writer
	for _, line := range lines {
		if fw.filter(line) {
			// Line passed the filter, write it plus a newline
			fw.writer.Write([]byte(line + "\n"))
		}
	}
	
	// Return the original length, since we've "consumed" all the bytes,
	// even if we didn't write them all
	return len(p), nil
} 