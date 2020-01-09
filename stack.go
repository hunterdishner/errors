package errors

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"runtime"
	"strings"
)

const (
	sep       = "/"
	centerDot = "·"
	dot       = "."
)

var (
	unwantedFrames = []string{"gorilla", "runtime"}
	dunno          = []byte("?")
)

type stack struct {
	*runtime.Frames
	formatted []string
}

func callingLocations(depth, skip int) *stack {
	d := make([]uintptr, depth)
	n := runtime.Callers(skip, d)
	if n == 0 {
		return nil
	}

	d = d[:n]
	callFrames := runtime.CallersFrames(d)

	return &stack{Frames: callFrames}
}

func (st *stack) frames() []string {
	var (
		frames   []string
		lines    [][]byte
		lastFile string
	)

	for {
		frame, more := st.Next()
		if !more {
			break
		}

		var discardRemainder bool
		for _, f := range unwantedFrames {
			if strings.Contains(frame.File, f) {
				discardRemainder = true
				// If we got this far up the stack, we don't need the rest
				break
			}
		}

		if discardRemainder {
			break
		}

		buf := &bytes.Buffer{}
		fmt.Fprintf(buf, "%s:%d ", fileName(frame.Function, frame.File), frame.Line)

		if frame.File != lastFile {
			data, err := ioutil.ReadFile(frame.File)
			if err != nil {
				// Print the function name rather than source
				fmt.Fprintf(buf, "%s", function(frame.Function))
				frames = append(frames, buf.String())
				continue
			}
			lines = bytes.Split(data, []byte{'\n'})
			lastFile = frame.File
		}

		fmt.Fprintf(buf, "%s", source(lines, frame.Line-1))
		frames = append(frames, buf.String())
	}

	return frames
}

func (st *stack) Format(s fmt.State, verb rune) {
	if len(st.formatted) == 0 {
		st.formatted = st.frames()
	}

	f := "\n" + strings.Join(st.formatted, "\n")
	fmt.Fprint(s, f)
}

func (st *stack) MarshalJSON() ([]byte, error) {
	if len(st.formatted) == 0 {
		st.formatted = st.frames()
	}

	buf := bytes.NewBuffer([]byte("["))
	for i, frame := range st.formatted {
		if i > 0 {
			fmt.Fprint(buf, ",")
		}
		fmt.Fprintf(buf, "%q", frame)
	}
	fmt.Fprint(buf, "]")

	return buf.Bytes(), nil
}

// source returns a space-trimmed slice of the n'th line.
func source(lines [][]byte, n int) []byte {
	if n < 0 || n >= len(lines) {
		return dunno
	}
	return bytes.Trim(lines[n], " \t")
}

func function(name string) string {
	if period := strings.Index(name, dot); period >= 0 {
		name = name[period+1:]
	}
	name = strings.Replace(name, centerDot, dot, -1)
	return name
}

func fileName(name, file string) string {
	// Here we want to get the source file path relative to the compile time
	// GOPATH. As of Go 1.6.x there is no direct way to know the compiled
	// GOPATH at runtime, but we can infer the number of path segments in the
	// GOPATH. We note that fn.Name() returns the function name qualified by
	// the import path, which does not include the GOPATH. Thus we can trim
	// segments from the beginning of the file path until the number of path
	// separators remaining is one more than the number of path separators in
	// the function name. For example, given:
	//
	//    GOPATH     /home/user
	//    file       /home/user/src/pkg/sub/file.go
	//    fn.Name()  pkg/sub.Type.Method
	//
	// We want to produce:
	//
	//    pkg/sub/file.go
	//
	// From this we can easily see that fn.Name() has one less path separator
	// than our desired output. We count separators from the end of the file
	// path until it finds two more than in the function name and then move
	// one character forward to preserve the initial path segment without a
	// leading separator.
	goal := strings.Count(name, sep) + 2
	i := len(file)
	for n := 0; n < goal; n++ {
		i = strings.LastIndex(file[:i], sep)
		if i == -1 {
			// not enough separators found, set i so that the slice expression
			// below leaves file unmodified
			i = -len(sep)
			break
		}
	}
	// get back to 0 or trim the leading separator
	file = file[i+len(sep):]
	return file
}
