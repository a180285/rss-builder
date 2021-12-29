package gin_middlewares

import (
	"bytes"
	"fmt"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"runtime"
	"strings"
)

var (
	dunno     = []byte("???")
	centerDot = []byte("·")
	dot       = []byte(".")
	slash     = []byte("/")
)

type ErrorResponse struct {
	ErrorMessage string      `json:"errorMessage"`
	Error        interface{} `json:"Error"`
	Stack        []string    `json:"stack"`
}

type ClientErrorResponse struct {
	ErrorMessage string   `json:"errorMessage"`
	Stack        []string `json:"stack"`
	ErrorCode    int      `json:"errorCode"`
}

// RecoveryWithWriter returns a middleware for a given writer that recovers from any panics and writes a 500 if there was one.
func MyGinRecovery(c *gin.Context) {
	defer func() {
		if err := recover(); err != nil {
			stack := GetStack(3)
			httpRequest, _ := httputil.DumpRequest(c.Request, false)

			errAndStack := fmt.Sprintf("%s\n%s", err, stack)
			fmt.Printf("[Panic recovered]:\n%s\n%s\n", string(httpRequest), errAndStack)
			stackList := strings.Split(string(stack), "\n")
			stackList = stackList[:len(stackList)-1]

			errorResponse := ErrorResponse{
				ErrorMessage: fmt.Sprintf("%s", err),
				Error:        err,
				Stack:        stackList,
			}
			c.AbortWithStatusJSON(http.StatusInternalServerError, errorResponse)
		}
	}()
	c.Next()
}

// stack returns a nicely formatted stack frame, skipping skip frames.
func GetStack(skip int) []byte {
	buf := new(bytes.Buffer) // the returned data
	// As we loop, we open files and read them. These variables record the currently
	// loaded file.
	var lines [][]byte
	var lastFile string
	for i := skip; ; i++ { // Skip the expected number of frames
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		// Print this much at least.  If we can't find the source, it won't show.
		fmt.Fprintf(buf, "%s:%d (0x%x)\n", file, line, pc)
		if file != lastFile {
			data, err := ioutil.ReadFile(file)
			if err != nil {
				continue
			}
			lines = bytes.Split(data, []byte{'\n'})
			lastFile = file
		}
		fmt.Fprintf(buf, "\t%s: %s\n", function(pc), source(lines, line))
	}
	return buf.Bytes()
}

// source returns a space-trimmed slice of the n'th line.
func source(lines [][]byte, n int) []byte {
	n-- // in stack trace, lines are 1-indexed but our array is 0-indexed
	if n < 0 || n >= len(lines) {
		return dunno
	}
	return bytes.TrimSpace(lines[n])
}

// function returns, if possible, the name of the function containing the PC.
func function(pc uintptr) []byte {
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return dunno
	}
	name := []byte(fn.Name())
	// The name includes the path name to the package, which is unnecessary
	// since the file name is already included.  Plus, it has center dots.
	// That is, we see
	//	runtime/debug.*T·ptrmethod
	// and want
	//	*T.ptrmethod
	// Also the package path might contains dot (e.g. code.google.com/...),
	// so first eliminate the path prefix
	if lastslash := bytes.LastIndex(name, slash); lastslash >= 0 {
		name = name[lastslash+1:]
	}
	if period := bytes.Index(name, dot); period >= 0 {
		name = name[period+1:]
	}
	name = bytes.Replace(name, centerDot, dot, -1)
	return name
}
