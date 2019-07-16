package cpython

import (
	"errors"
	"github.com/sbinet/go-python"
	"github.com/sirupsen/logrus"
)

var pyModule, ut, exec, compile *python.PyObject
var pyThreadState *python.PyThreadState
var PyStr = python.PyString_FromString
var GoStr = python.PyString_AS_STRING

func init() {
	err := python.Initialize()
	if err != nil {
		panic(err.Error())
	}

	pyModule = python.PyImport_ImportModule("sys")
	ut = pyModule.GetAttrString("ut")
	exec = pyModule.GetAttrString("exec_code")
	compile = pyModule.GetAttrString("compile")
	pyThreadState = python.PyEval_SaveThread()
}

// ImportModule will import cpython module from given directory
func ImportModule(dir, name string) *python.PyObject {
	sysModule := python.PyImport_ImportModule("sys") // import sys
	path := sysModule.GetAttrString("path")          // path = sys.path
	python.PyList_Insert(path, 0, PyStr(dir))        // path.insert(0, dir)
	return python.PyImport_ImportModule(name)        // return __import__(name)
}

func SendEmail(smtp, sendAccount, sendPassword, toAccount, content string) error {
	python.PyEval_RestoreThread(pyThreadState)

	logrus.Debug("smtp:", smtp)
	logrus.Debug("sendAccount:", sendAccount)
	logrus.Debug("sendPassword:", sendPassword)
	logrus.Debug("toAccount:", toAccount)
	logrus.Debug("content:", content)

	m := python.PyImport_ImportModule("sys")
	if m == nil {
		return errors.New("import sys error")
	}
	path := m.GetAttrString("path")
	if path == nil {
		return errors.New("get path error")
	}
	logrus.Debug("test0")

	//加入当前目录，空串表示当前目录
	currentDir := python.PyString_FromString("/data/backend_svr/tools")
	if err := python.PyList_Insert(path, 0, currentDir); err != nil {
		return errors.New("get path error")
	}

	logrus.Debug("test1")

	m = python.PyImport_ImportModule("password_email")
	if m == nil {
		return errors.New("import password_email error")
	}
	sendEmail := m.GetAttrString("send_email")
	if sendEmail == nil {
		return errors.New("get sendEmail error")
	}

	logrus.Debug("test2")

	out := sendEmail.CallFunction(python.PyString_FromString(smtp),
		python.PyString_FromString(sendAccount), python.PyString_FromString(sendPassword),
		python.PyString_FromString(toAccount), python.PyString_FromString(content))

	pyThreadState = python.PyEval_SaveThread()

	if out == nil {
		return errors.New("call sendEmail error")
	}

	return nil
}
