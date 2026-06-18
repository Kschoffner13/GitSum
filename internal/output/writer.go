package output

import "io"

// Write writes the summary to w.
func Write(w io.Writer, summary string) error {
	_, err := io.WriteString(w, summary)
	return err
}
