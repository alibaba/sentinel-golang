package k8s

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/alibaba/sentinel-golang/logging"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func Test_k8sLogger_Info(t *testing.T) {
	t.Run("Test_k8sLogger_Info", func(t *testing.T) {
		fileName := "k8s-logger-adapter-test1.log"
		tmpDir := os.TempDir()
		if !strings.HasSuffix(tmpDir, string(os.PathSeparator)) {
			tmpDir = tmpDir + string(os.PathSeparator)
		}
		logger, err := logging.NewSimpleFileLogger(tmpDir + fileName)
		assert.NoError(t, err)
		time.Sleep(time.Second * 1)
		defer func() {
			if err := os.Remove(tmpDir + fileName); err != nil {
				t.Fatal(err)
			}
		}()

		logging.ResetGlobalLoggerLevel(logging.DebugLevel)
		defer logging.ResetGlobalLoggerLevel(logging.InfoLevel)
		var l logr.Logger = &k8SLogger{
			l:             logger,
			level:         logging.GetGlobalLoggerLevel(),
			names:         make([]string, 0),
			keysAndValues: make([]interface{}, 0),
		}
		l.V(int(logging.InfoLevel)).Info("info test msg", "k1", "v1")
		l.V(int(logging.WarnLevel)).Info("warn test msg", "k1", "v1")
		l.V(int(logging.DebugLevel)).Info("debug test msg", "k1", "v1")
		l.Error(errors.New("error test"), "error test msg", "k1", "v1")
		logFile, err := os.OpenFile(tmpDir+fileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777)
		if err != nil {
			t.Fatal(err)
		}
		defer func() {
			if e := logFile.Close(); e != nil {
				t.Fatal(e)
			}
		}()
		reader := bufio.NewReader(logFile)
		l1, _, err := reader.ReadLine()
		if err != nil {
			t.Fatal(err)
		}
		assert.True(t, strings.Contains(string(l1), `"logLevel":"INFO","msg":"info test msg","k1":"v1"}`))

		l2, _, err := reader.ReadLine()
		if err != nil {
			t.Fatal(err)
		}
		assert.True(t, strings.Contains(string(l2), `"logLevel":"WARNING","msg":"warn test msg","k1":"v1"}`))

		l3, _, err := reader.ReadLine()
		if err != nil {
			t.Fatal(err)
		}
		assert.True(t, strings.Contains(string(l3), `"logLevel":"DEBUG","msg":"debug test msg","k1":"v1"}`))

		l4, _, err := reader.ReadLine()
		if err != nil {
			t.Fatal(err)
		}
		assert.True(t, strings.Contains(string(l4), `"logLevel":"ERROR","msg":"error test msg","k1":"v1"}`))
	})

	t.Run("Test_k8sLogger_Info2", func(t *testing.T) {
		fileName := "k8s-logger-adapter-test2.log"
		tmpDir := os.TempDir()
		if !strings.HasSuffix(tmpDir, string(os.PathSeparator)) {
			tmpDir = tmpDir + string(os.PathSeparator)
		}
		logger, err := logging.NewSimpleFileLogger(tmpDir + fileName)
		assert.NoError(t, err)
		time.Sleep(time.Second * 1)
		defer func() {
			if err := os.Remove(tmpDir + fileName); err != nil {
				t.Fatal(err)
			}
		}()

		logging.ResetGlobalLoggerLevel(logging.DebugLevel)
		defer logging.ResetGlobalLoggerLevel(logging.InfoLevel)
		var l logr.Logger = &k8SLogger{
			l:             logger,
			level:         logging.GetGlobalLoggerLevel(),
			names:         make([]string, 0),
			keysAndValues: make([]interface{}, 0),
		}
		l = l.WithName("k8s-logger")
		l = l.WithValues("k2", "v2")
		l.V(int(logging.InfoLevel)).Info("info test msg", "k1", "v1")
		logFile, err := os.OpenFile(tmpDir+fileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777)
		if err != nil {
			t.Fatal(err)
		}
		defer func() {
			if e := logFile.Close(); e != nil {
				t.Fatal(e)
			}
		}()

		reader := bufio.NewReader(logFile)
		l1, _, err := reader.ReadLine()
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println(string(l1))
		assert.True(t, strings.Contains(string(l1), `"logLevel":"INFO","msg":"k8s-logger info test msg","k1":"v1","k2":"v2"}`))
	})
}
