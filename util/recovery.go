package util

import (
	"fmt"
	"github.com/murang/potato/log"
	"runtime"
	"strings"
)

func Trace(msg string) string {
	var pcs [32]uintptr
	n := runtime.Callers(3, pcs[:])

	var str strings.Builder
	str.WriteString(msg + "\nPanic Trace =========================>")
	frames := runtime.CallersFrames(pcs[:n])
	for {
		frame, more := frames.Next()
		str.WriteString(fmt.Sprintf("\n\t%s:%d", frame.File, frame.Line))
		if !more {
			break
		}
	}
	return str.String()
}

func Recovery() {
	if r := recover(); r != nil {
		message := fmt.Sprintf("%s", r)
		log.Sugar.Errorf("%s\n\n", Trace(message))
	}
}

func GoSafe(fn func()) {
	go func() {
		defer Recovery()
		fn()
	}()
}
