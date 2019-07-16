package cpython

import (
	"errors"
	"github.com/sbinet/go-python"
	"github.com/sirupsen/logrus"
)

//var pyModule, ut, exec, compile *python.PyObject
var pyModule *python.PyObject
var pyThreadState *python.PyThreadState
var PyStr = python.PyString_FromString

//var GoStr = python.PyString_AS_STRING

func init() {
	err := python.Initialize()
	if err != nil {
		panic(err.Error())
	}

	pyModule = python.PyImport_ImportModule("sys")
	//ut = pyModule.GetAttrString("ut")
	//exec = pyModule.GetAttrString("exec_code")
	//compile = pyModule.GetAttrString("compile")
	pyThreadState = python.PyEval_SaveThread()
}

// ImportModule will import cpython module from given directory
func ImportModule(dir, name string) *python.PyObject {
	sysModule := python.PyImport_ImportModule("sys") // import sys
	path := sysModule.GetAttrString("path")          // path = sys.path

	if err := python.PyList_Insert(path, 0, PyStr(dir)); err != nil {
		logrus.Error("ImportModule error")
	}

	return python.PyImport_ImportModule(name) // return __import__(name)
}

func SendEmail(smtp, sendAccount, sendPassword, toAccount, content string) error {
	python.PyEval_RestoreThread(pyThreadState)

	m := ImportModule("/data/backend_svr/tools", "password_email")
	if m == nil {
		return errors.New("import password_email error")
	}
	sendEmail := m.GetAttrString("send_email")
	if sendEmail == nil {
		return errors.New("get sendEmail error")
	}

	out := sendEmail.CallFunction(python.PyString_FromString(smtp),
		python.PyString_FromString(sendAccount), python.PyString_FromString(sendPassword),
		python.PyString_FromString(toAccount), python.PyString_FromString(content))

	pyThreadState = python.PyEval_SaveThread()

	if out == nil {
		return errors.New("call sendEmail error")
	}

	return nil
}
