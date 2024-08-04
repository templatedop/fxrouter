package fxrouter

import (
	"fmt"
	logger"github.com/templatedop/fxlogger"
	
	//logger"gotemplate/logger"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"

	"runtime"

	"github.com/gin-gonic/gin"
)

type fake struct{}

var (
	StackTraceMaxDepth = 6

	packageName = reflect.TypeOf(fake{}).PkgPath()
)

func Exception(logger *logger.Logger) gin.HandlerFunc {

	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {

				NewStacktrace := newStacktrace()
				printedstack := PrintStackTrace(NewStacktrace.frames)
				stackerr, ok := r.(error)
				if !ok {
					stackerr = fmt.Errorf("%v", r)
				}

				bodyBytes, err := io.ReadAll(c.Request.Body)
				if err != nil {
					bodyBytes = []byte("Error reading body")
				}
				bodyString := string(bodyBytes)
				logger.ToZerolog().
					Error().
					Err(fmt.Errorf("%v", printedstack)).
					Str("method", c.Request.Method).
					Str("url", c.Request.URL.String()).
					Str("context:", bodyString).
					Str("", stackerr.Error()).Msg("Panic recovered")

				c.AbortWithStatusJSON(500, gin.H{"error": "Internal Server Error"})

			}

		}()
		c.Next()

	}
}

type oopsStacktrace struct {
	//span   string
	frames []oopsStacktraceFrame
}
type oopsStacktraceFrame struct {
	//pc       uintptr
	file     string
	function string
	line     int
}

func PrintStackTrace(frames []oopsStacktraceFrame) string {
	var builder strings.Builder
	for _, frame := range frames {

		builder.WriteString(fmt.Sprintf("%v:%v %v() \n", frame.file, frame.line, frame.function))
	}
	return builder.String()
}
func (frame *oopsStacktraceFrame) String() string {
	currentFrame := fmt.Sprintf("%v:%v", frame.file, frame.line)
	if frame.function != "" {
		currentFrame = fmt.Sprintf("%v:%v %v() \n\n", frame.file, frame.line, frame.function)
	}

	return currentFrame
}

func newStacktrace() *oopsStacktrace {

	frames := []oopsStacktraceFrame{}
	for i := 0; len(frames) < StackTraceMaxDepth; i++ {

		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		file = removeGoPath(file)

		f := runtime.FuncForPC(pc)
		if f == nil {
			break
		}
		function := shortFuncName(f)

		isGoPkg := len(runtime.GOROOT()) > 0 && strings.Contains(file, runtime.GOROOT())
		isOopsPkg := strings.Contains(file, packageName)
		isGoSrc := strings.Contains(file, "/src")
		isExceptionSrc := strings.Contains(file, "Exception.go")
		isGinContrib := strings.Contains(file, "gin-contrib")

		if isGoPkg || isOopsPkg || isGoSrc || isExceptionSrc || isGinContrib {
			continue
		}

		frames = append(frames, oopsStacktraceFrame{
			//pc:       pc,
			file:     file,
			function: function,
			line:     line,
		})
	}

	return &oopsStacktrace{
		//span:   stacks,
		frames: frames,
	}

}
func removeGoPath(path string) string {
	dirs := filepath.SplitList(os.Getenv("GOPATH"))

	sort.Stable(longestFirst(dirs))
	for _, dir := range dirs {
		srcdir := filepath.Join(dir, "src")
		rel, err := filepath.Rel(srcdir, path)

		if err == nil && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			return rel
		}
	}
	return path
}

func shortFuncName(f *runtime.Func) string {

	longName := f.Name()

	withoutPath := longName[strings.LastIndex(longName, "/")+1:]
	withoutPackage := withoutPath[strings.Index(withoutPath, ".")+1:]

	shortName := withoutPackage
	shortName = strings.Replace(shortName, "(", "", 1)
	shortName = strings.Replace(shortName, "*", "", 1)
	shortName = strings.Replace(shortName, ")", "", 1)

	return shortName
}

type longestFirst []string

func (strs longestFirst) Len() int           { return len(strs) }
func (strs longestFirst) Less(i, j int) bool { return len(strs[i]) > len(strs[j]) }
func (strs longestFirst) Swap(i, j int)      { strs[i], strs[j] = strs[j], strs[i] }
