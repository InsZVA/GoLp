package goLp

import (
	"bufio"
	"fmt"
	"github.com/pkg/errors"
	"os"
	"os/exec"
)

var UnExpectedResult error = errors.New("unexpected result file format")

type Solver interface {
	Solve(lpFile string, status *string) (map[string]float64, error)
}

type CBCSolver struct {
	cbcPath string
}

func (solver *CBCSolver) Solve(lpFile string, status *string) (map[string]float64, error) {
	resFile := lpFile+"res.txt"
	cmd := exec.Command(solver.cbcPath, lpFile, "solve", "solu", resFile)
	err := cmd.Run()
	if err != nil {
		return nil, err
	}
	f, err := os.Open(resFile)
	if err != nil {
		return nil, err
	}
	reader := bufio.NewReader(f)
	n, err := fmt.Fscanf(reader, "%s", status)
	if err != nil {
		return nil, err
	}
	if n != 1 {
		return nil, UnExpectedResult
	}

	var skip string
	for i := 1; i <= 4; i++ {
		n, err = fmt.Fscanf(reader, "%s", &skip)
		if err != nil {
			return nil, err
		}
		if n != 1 {
			return nil, UnExpectedResult
		}
	}
	var id int
	var name string
	var val, addition float64
	res := make(map[string]float64)
	for n, err = fmt.Fscanf(reader, "\n%d%s%f%f", &id, &name, &val, &addition); err == nil && n == 4;
	n, err = fmt.Fscanf(reader, "\n%d%s%f%f", &id, &name, &val, &addition) {
		res[name] = val
	}
	err = f.Close()
	if err != nil {
		return nil, err
	}
	return res, nil
}