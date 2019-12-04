package tr

import (
	"context"
	"fmt"
	"os"
	"path"
	"runtime"
	"strings"
)

const (
	infoFormat = "%s \033[0;32mInfo\033[0m  (\033[0;33m%s\033[0m)"
	warnFormat = "%s \033[1;36mWarn\033[0m  (\033[0;33m%s\033[0m)"
	errFormat  = "%s \033[0;31mError\033[0m (\033[0;33m%s\033[0m)"
)

var (
	inChan   chan string
	ctxTr    context.Context
	cancelTr context.CancelFunc
)

func init() {
	inChan = make(chan string)
	ctxTr, cancelTr = context.WithCancel(context.Background())
}

func Init() {
	go func() {
		for {
			select {
			case <-ctxTr.Done():
				return
			case text := <-inChan:
				fmt.Fprint(os.Stdout, text+"\n")
			}
		}
	}()
}

func Cancel() {
	cancelTr()
}

func In() {
	inChan <- fmt.Sprintf("%s >>", location(2))
}

func Out() {
	inChan <- fmt.Sprintf("%s <<", location(2))
}

func Info(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	if len(args) == 0 {
		msg = format
	}
	inChan <- fmt.Sprintf(infoFormat, location(2), msg)
}

func Warning(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	if len(args) == 0 {
		msg = format
	}
	inChan <- fmt.Sprintf(warnFormat, location(2), msg)
}

func Error(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	if len(args) == 0 {
		msg = format
	}
	inChan <- fmt.Sprintf(errFormat, location(2), msg)
}

func IsOK(err error) bool {
	if err != nil {
		fmt.Println(err)
		return false
	}
	return true
}

func location(skip int) string {
	if pc, file, line, ok := runtime.Caller(skip); ok {
		if fn := runtime.FuncForPC(pc); fn != nil {
			name := fn.Name()
			if idx := strings.LastIndex(name, "."); idx != -1 {
				return fmt.Sprintf("%s:%d - %s", path.Base(file), line, name[idx+1:])
			}
		}
	}
	return ""
}
