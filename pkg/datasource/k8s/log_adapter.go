package k8s

import (
	"strings"

	"github.com/alibaba/sentinel-golang/logging"
	"github.com/go-logr/logr"
)

// noopInfoLogger is a logr.InfoLogger that's always disabled, and does nothing.
type noopInfoLogger struct{}

func (l *noopInfoLogger) Enabled() bool                   { return false }
func (l *noopInfoLogger) Info(_ string, _ ...interface{}) {}

var disabledInfoLogger = &noopInfoLogger{}

type k8SLogger struct {
	l             logging.Logger
	level         logging.Level
	names         []string
	keysAndValues []interface{}
}

func (k *k8SLogger) buildNames() string {
	size := len(k.names)
	if size == 0 {
		return ""
	}
	sb := strings.Builder{}
	for i, name := range k.names {
		sb.WriteString(name)
		if i == size-1 {
			continue
		}
		sb.WriteString(".")
	}
	sb.WriteString(" ")
	return sb.String()
}

func (k *k8SLogger) Info(msg string, keysAndValues ...interface{}) {
	keysAndValues = append(keysAndValues, k.keysAndValues...)
	switch k.level {
	case logging.WarnLevel:
		k.l.Warn(k.buildNames()+msg, keysAndValues...)
	case logging.InfoLevel:
		k.l.Info(k.buildNames()+msg, keysAndValues...)
	case logging.DebugLevel:
		k.l.Debug(k.buildNames()+msg, keysAndValues...)
	default:
		k.l.Info(k.buildNames()+msg, keysAndValues...)
	}
}

func (k *k8SLogger) Enabled() bool {
	return true
}

func (k *k8SLogger) Error(err error, msg string, keysAndValues ...interface{}) {
	keysAndValues = append(keysAndValues, k.keysAndValues...)
	k.l.Error(err, k.buildNames()+msg, keysAndValues...)
}

func (k *k8SLogger) V(level int) logr.InfoLogger {
	if k.Enabled() {
		names := make([]string, len(k.names))
		copy(names, k.names)
		kvs := make([]interface{}, len(k.keysAndValues))
		copy(kvs, k.keysAndValues)
		return &k8SLogger{
			l:             k.l,
			level:         logging.Level(level),
			names:         names,
			keysAndValues: kvs,
		}
	}
	return disabledInfoLogger
}

func (k *k8SLogger) WithValues(keysAndValues ...interface{}) logr.Logger {
	names := make([]string, len(k.names))
	copy(names, k.names)
	kvs := make([]interface{}, len(k.keysAndValues))
	copy(kvs, k.keysAndValues)
	kvs = append(kvs, keysAndValues...)
	return &k8SLogger{
		l:             k.l,
		level:         k.level,
		names:         names,
		keysAndValues: kvs,
	}
}

func (k *k8SLogger) WithName(name string) logr.Logger {
	names := make([]string, len(k.names))
	copy(names, k.names)
	names = append(names, name)
	kvs := make([]interface{}, len(k.keysAndValues))
	copy(kvs, k.keysAndValues)
	return &k8SLogger{
		l:             k.l,
		level:         k.level,
		names:         names,
		keysAndValues: kvs,
	}
}
