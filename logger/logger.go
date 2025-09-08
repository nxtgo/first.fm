package logger

import (
	"os"
	"time"

	"github.com/nxtgo/zlog"
)

var Log *zlog.Logger

func init() {
	Log = zlog.New()
	Log.SetOutput(os.Stdout)
	Log.SetTimeFormat(time.Kitchen)
	Log.EnableColors(true)
	Log.ShowCaller(true)
	if os.Getenv("MODE") == "" {
		Log.SetLevel(zlog.LevelDebug)
	} else {
		Log.SetLevel(zlog.LevelInfo)
	}
}
