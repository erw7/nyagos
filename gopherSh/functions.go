package gopherSh

import (
	"bytes"
	"context"
	"errors"
	"io/ioutil"
	"os"
	"strings"

	"github.com/yuin/gopher-lua"
	"github.com/zetamatta/nyagos/alias"
	"github.com/zetamatta/nyagos/shell"
)

type LuaBinaryChank struct {
	Chank *lua.LFunction
}

func (this LuaBinaryChank) String() string {
	return this.Chank.String()
}

func (this *LuaBinaryChank) Call(ctx context.Context, cmd *shell.Cmd) (int, error) {
	luawrapper, ok := cmd.Tag().(*luaWrapper)
	if !ok {
		return 255, errors.New("LuaBinaryChank.Call: Lua instance not found")
	}
	L := luawrapper.Lua
	ctx = context.WithValue(ctx, luaKey, L)
	L.Push(this.Chank)

	table := L.NewTable()
	for i, arg1 := range cmd.Args() {
		L.SetTable(table, lua.LNumber(i), lua.LString(arg1))
	}
	rawargs := L.NewTable()
	for i, arg1 := range cmd.RawArgs() {
		L.SetTable(rawargs, lua.LNumber(i), lua.LString(arg1))
	}
	L.SetField(table, "rawargs", rawargs)
	L.Push(table)

	errorlevel := 0
	err := callLua(ctx, &cmd.Shell, 1, 1)
	if err == nil {
		switch val := L.Get(-1).(type) {
		case *lua.LTable:
			size := val.Len()
			newargs := make([]string, size)
			for i := 0; i < size; i++ {
				newargs[i] = L.GetTable(val, lua.LNumber(i+1)).String()
			}
			sh := cmd.Command()
			sh.SetArgs(newargs)
			errorlevel, err = sh.Spawnvp(ctx)
			sh.Close()
		case lua.LNumber:
			errorlevel = int(val)
		case lua.LString:
			errorlevel, err = cmd.Interpret(ctx, string(val))
		}
		L.Pop(1)
	}
	return errorlevel, err
}

func cmdSetAlias(L Lua) int {
	key := strings.ToLower(L.ToString(-2))
	switch L.Get(-1).Type() {
	case lua.LTString:
		alias.Table[key] = alias.New(L.ToString(-1))
	case lua.LTFunction:
		alias.Table[key] = &LuaBinaryChank{Chank: L.ToFunction(-1)}
	case lua.LTNil:
		delete(alias.Table, key)
	}
	L.Push(lua.LTrue)
	return 1
}

func cmdGetAlias(L Lua) int {
	value, ok := alias.Table[strings.ToLower(L.ToString(-1))]
	if !ok {
		L.Push(lua.LNil)
		return 1
	}
	switch v := value.(type) {
	case *LuaBinaryChank:
		L.Push(v.Chank)
	default:
		L.Push(lua.LString(v.String()))
	}
	return 1
}

func cmdExec(L Lua) int {
	errorlevel := 0
	var err error
	table, ok := L.Get(1).(*lua.LTable)
	if ok {
		n := table.Len()
		args := make([]string, 0, n)
		for i := 1; i <= n; i++ {
			arg1 := L.GetTable(table, lua.LNumber(i)).String()
			args = append(args, arg1)
		}
		ctx, sh := getRegInt(L)
		if sh == nil {
			println("main/lua_cmd.go: cmdExec: not found interpreter object")
			sh = shell.New()
			newL, err := Clone(L)
			if err == nil && newL != nil {
				sh.SetTag(&luaWrapper{Lua: newL})
			}
			defer sh.Close()
		}
		cmd := sh.Command()
		defer cmd.Close()
		cmd.SetArgs(args)
		errorlevel, err = cmd.Spawnvp(ctx)
	} else {
		statement, ok := L.Get(1).(lua.LString)
		if !ok {
			return lerror(L, "nyagos.exec: the 1st argument is not a string")
		}
		ctx, sh := getRegInt(L)
		if ctx == nil {
			return lerror(L, "nyagos.exec: context not found")
		}
		if sh == nil {
			println("nyagos.exec: warning shell is not found.")
			sh = shell.New()
			sh.SetTag(&luaWrapper{L})
			defer sh.Close()
		}
		errorlevel, err = sh.Interpret(ctx, string(statement))
	}
	L.Push(lua.LNumber(errorlevel))
	if err != nil {
		L.Push(lua.LString(err.Error()))
	} else {
		L.Push(lua.LNil)
	}
	return 2
}

func cmdEval(L Lua) int {
	statement, ok := L.Get(1).(lua.LString)
	if !ok {
		L.Push(lua.LNil)
		L.Push(lua.LString("nyagos.eval: an argument is not string"))
		return 2
	}
	r, w, err := os.Pipe()
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}
	go func(statement string, w *os.File) {
		ctx, sh := getRegInt(L)
		if ctx == nil {
			ctx = context.Background()
			println("cmdEval: context not found.")
		}
		if sh == nil {
			sh = shell.New()
			println("cmdEval: shell not found.")
			defer sh.Close()
		}
		sh.SetTag(&luaWrapper{L})
		saveOut := sh.Stdout
		sh.Stdout = w
		sh.Interpret(ctx, statement)
		sh.Stdout = saveOut
		w.Close()
	}(string(statement), w)

	result, err := ioutil.ReadAll(r)
	r.Close()
	if err == nil {
		L.Push(lua.LString(string(bytes.Trim(result, "\r\n\t "))))
	} else {
		L.Push(lua.LNil)
	}
	return 1
}
