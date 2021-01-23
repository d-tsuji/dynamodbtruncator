package testonly

import (
	"os/exec"
	"strings"
	"testing"
)

func CmdsExec(t *testing.T, cmds []string) {
	t.Helper()
	for _, cmd := range cmds {
		CmdExec(t, cmd)
	}
}

func CmdExec(t *testing.T, cmd string) {
	t.Helper()
	args := strings.Split(cmd, " ")
	if err := exec.Command(args[0], args[1:]...).Run(); err != nil {
		t.Errorf("cmd (%s) exec: %v", cmd, err)
	}
}

func CmdExecCombinedOutput(t *testing.T, cmd string) []byte {
	t.Helper()
	args := strings.Split(cmd, " ")
	b, err := exec.Command(args[0], args[1:]...).CombinedOutput()
	if err != nil {
		t.Errorf("cmd (%s) exec: %v", cmd, err)
	}
	return b
}
