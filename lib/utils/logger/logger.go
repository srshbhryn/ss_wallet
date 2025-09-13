package logger

import (
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"wallet/lib/config"

	"github.com/google/uuid"
	"github.com/natefinch/lumberjack"
)

var GitCommit = "dev" // overwritten by -ldflags

var logger *slog.Logger

func init() {
	initLogger(os.Args[len(os.Args)-1])
}

func Get() *slog.Logger {
	pkg, fn := getCallerPackageAndFunc()
	return logger.With("package", pkg, "function", fn)
}

func initLogger(appName string) {
	logDir := config.LogDir
	var level slog.Level
	if config.Env.Get() == config.PROD {
		level = slog.LevelInfo
	} else {
		level = slog.LevelDebug
	}
	logger = slog.New(slog.NewJSONHandler(&lumberjack.Logger{
		Filename: filepath.Join(logDir, appName+".log"),
		MaxSize:  1024,
	}, &slog.HandlerOptions{
		Level: level,
	})).With("app", appName, "instance", uuid.New().String(), "git_commit", GitCommit)
	logger.Info("logger started")
}

// getCallerPackageAndFunc returns the clean package path and function name with receiver
func getCallerPackageAndFunc() (string, string) {
	pc, _, _, ok := runtime.Caller(2)
	if !ok {
		return "unknown", "unknown"
	}
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return "unknown", "unknown"
	}

	fullName := fn.Name() // e.g. "github.com/me/project/pkg.(*MyType).DoSomething"

	// Separate package and function
	lastSlash := strings.LastIndex(fullName, "/")
	lastDot := strings.LastIndex(fullName, ".")
	if lastDot == -1 || lastDot < lastSlash {
		return fullName, "unknown"
	}

	pkg := fullName[:lastDot]
	function := fullName[lastDot+1:]

	// Remove package path from function if it contains the receiver
	if openParen := strings.Index(function, "("); openParen != -1 {
		// The receiver exists, keep it with the function
		// pkg is everything before the receiver
		pkg = fullName[:lastSlash]
	}

	return pkg, function
}
