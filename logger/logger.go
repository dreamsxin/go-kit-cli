package logger

import (
	"fmt"
	"log"
)

func init() {
	log.SetFlags(log.Llongfile)
}

const (
	ERROR = iota // 0
	WARN         // 1
	INFO         // 2
	TRACE        // 3
	DEBUG        // 4
)

var Logger = &LevelLogger{}

type LevelLogger struct {
	Level int
}

func (l *LevelLogger) Log(lvl int, a ...interface{}) {
	if lvl <= l.Level {
		s := fmt.Sprint(a...)
		log.Output(2, s)
	}
}

func (l *LevelLogger) Logf(lvl int, format string, a ...interface{}) {
	if lvl <= l.Level {
		s := fmt.Sprint(a...)
		log.Output(2, s)
	}
}

func (l *LevelLogger) Logln(lvl int, a ...interface{}) {
	if lvl <= l.Level {
		s := fmt.Sprint(a...)
		log.Output(2, s)
	}
}
