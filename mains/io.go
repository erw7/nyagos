package mains

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"github.com/yuin/gopher-lua"
)

type fileHandleT struct {
	scanner *bufio.Scanner
	closer  io.Closer
}

func ioLinesIter(L *lua.LState) int {
	ud, ok := L.Get(1).(*lua.LUserData)
	if !ok {
		L.Push(lua.LNil)
		return 1
	}
	fh, ok := ud.Value.(*fileHandleT)
	if !ok {
		L.Push(lua.LNil)
		return 1
	}
	if fh.scanner.Scan() {
		L.Push(lua.LString(fh.scanner.Text()))
		return 1
	}
	L.Push(lua.LNil)
	if fh.closer != nil {
		fh.closer.Close()
	}
	return 1
}

func ioLines(L *lua.LState) int {
	ud := L.NewUserData()
	_, sh := getRegInt(L)
	if L.GetTop() >= 1 {
		if filename, ok := L.Get(1).(lua.LString); ok {
			if fd, err := os.Open(string(filename)); err == nil {
				ud.Value = &fileHandleT{
					scanner: bufio.NewScanner(fd),
					closer:  fd,
				}
			} else {
				L.Push(lua.LNil)
				L.Push(lua.LString(fmt.Sprintf("%s: can not open", filename)))
				return 2
			}
		} else {
			L.Push(lua.LNil)
			L.Push(lua.LString("io.lines: not a string"))
			return 2
		}
	} else if sh != nil {
		ud.Value = &fileHandleT{
			scanner: bufio.NewScanner(sh.In()),
			closer:  nil,
		}
	} else {
		ud.Value = &fileHandleT{
			scanner: bufio.NewScanner(os.Stdin),
			closer:  nil,
		}
	}
	L.Push(L.NewFunction(ioLinesIter))
	L.Push(ud)
	L.Push(lua.LNil)
	return 3
}

func ioWrite(L *lua.LState) int {
	_, sh := getRegInt(L)
	out := sh.Out()
	for i, end := 1, L.GetTop(); i <= end; i++ {
		fmt.Fprint(out, L.Get(i).String())
	}
	return 0
}