package commands

import (
	"context"
	"io"
	"os"

	"github.com/zetamatta/nyagos/dos"
)

func getwd_() string {
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	return dos.NetDriveToUNC(wd)
}

func clone_(action string, out io.Writer) (int, error) {
	wd := getwd_()
	var err error
	var me string
	me, err = os.Executable()
	if err != nil {
		return 1, err
	}
	err = dos.ShellExecute(action, me, "", wd)
	if err != nil {
		err = dos.ShellExecute(action, dos.TruePath(me), "", wd)
	}
	if err != nil {
		err2 := dos.ShellExecute(action, "CMD.EXE", "/c \""+me+"\"", wd)
		if err2 != nil {
			return 1, err // return original error
		}
	}
	return 0, nil
}

func cmdClone(ctx context.Context, cmd Param) (int, error) {
	return clone_("open", cmd.Err())
}

func cmdSu(ctx context.Context, cmd Param) (int, error) {
	return clone_("runas", cmd.Err())
}
