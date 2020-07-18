package goLp

import (
	"fmt"
	"math"
	"testing"
)


func NewOptimalProblem() *Problem {
	p := NewProblem("a")
	x1 := p.NewNormalVar("x1", 0, 40)
	x2 := p.NewNormalVar("x2", math.Inf(-1), math.Inf(1))
	x3 := p.NewNormalVar("x3", math.Inf(-1), math.Inf(1))
	x4 := p.NewIntegerVar("x4", 2, 3)
	p.SetObj(NewObjBuilder().Add(1, x1).Add(1, x2).Add(3, x3).Add(1, x4).Maximize())
	p.AddST(NewSTBuilder("c1").Add(-1, x1).Add(1, x2).Add(1, x3).Add(10, x4).Leq().AddConst(40).Done())
	p.AddST(NewSTBuilder("c2").Add(1, x1).Add(-3, x2).Add(1, x3).Leq().AddConst(30).Done())
	p.AddST(NewSTBuilder("c3").Add(1, x2).Add(-3.5, x4).Eq().AddConst(0).Done())
	return p
}

func TestOptimal(t *testing.T) {
	cbc := &CBCSolver{cbcPath: "/usr/local/bin/cbc"}
	p := NewOptimalProblem()
	x1 := p.vars["x1"]
	x2 := p.vars["x2"]
	x3 := p.vars["x3"]
	x4 := p.vars["x4"]
	err := p.Solve(cbc)
	if err != nil {
		t.Error(err)
	}
	if p.status != "Optimal" {
		t.Errorf("expect optimal result, actual: %s", p.status)
	}
	if x1.getVal() != 31 {
		t.Errorf("expect x1 == 31, actual: %f", x1.getVal())
	}
	if x2.getVal() != 10.5 {
		t.Errorf("expect x2 == 10.5, actual: %f", x1.getVal())
	}
	if x3.getVal() != 30.5 {
		t.Errorf("expect x3 == 30.5, actual: %f", x1.getVal())
	}
	if x4.getVal() != 3 {
		t.Errorf("expect x4 == 3, actual: %f", x1.getVal())
	}
}

func NewUnboundedProblem() *Problem {
	p := NewProblem("a")
	x1 := p.NewNormalVar("x1", 0, 40)
	x2 := p.NewNormalVar("x2", math.Inf(-1), math.Inf(1))
	p.SetObj(NewObjBuilder().Add(1, x1).Add(-1, x2).Maximize())
	p.AddST(NewSTBuilder("c1").Add(1, x1).Add(1, x2).Leq().AddConst(40).Done())
	return p
}

func TestUnbounded(t *testing.T) {
	cbc := &CBCSolver{cbcPath: "/usr/local/bin/cbc"}
	p := NewUnboundedProblem()
	err := p.Solve(cbc)
	if err != nil {
		t.Error(err)
	}
	if p.status != "Unbounded" {
		t.Errorf("expect Unbounded result, actual: %s", p.status)
	}
}

func NewInfeasibleProblem() *Problem {
	p := NewProblem("a")
	x1 := p.NewNormalVar("x1", 0, 40)
	x2 := p.NewNormalVar("x2", math.Inf(-1), math.Inf(1))
	p.SetObj(NewObjBuilder().Add(1, x1).Add(-1, x2).Minimize())
	p.AddST(NewSTBuilder("c1").Add(1, x1).Add(1, x2).Leq().AddConst(40).Done())
	p.AddST(NewSTBuilder("c2").Add(1, x1).Add(1, x2).Geq().AddConst(50).Done())
	return p
}

func TestInfeasible(t *testing.T) {
	cbc := &CBCSolver{cbcPath: "/usr/local/bin/cbc"}
	p := NewInfeasibleProblem()
	err := p.Solve(cbc)
	if err != nil {
		t.Error(err)
	}
	if p.status != "Infeasible" {
		t.Errorf("expect Infeasible result, actual: %s", p.status)
	}
}

func TestTooLongName(t *testing.T) {
	p := NewProblem("p1")
	var name string
	for i := 1; i <= 30; i++ {
		name += "xxxxxxxxxx"
	}
	p.NewNormalVar(name, math.Inf(-1), math.Inf(1))
	err := p.Solve(&CBCSolver{cbcPath: "/usr/local/bin/cbc"})
	if err != VariableNameTooLongErr {
		t.Errorf("expect too long name error, actual: %v", err)
	}
}


func TestLargeProblem(t *testing.T) {
	p := NewProblem("large")
	vars := make(map[string]*Variable)
	builder := NewObjBuilder()
	for i := 1; i <= 1000; i++ {
		name := fmt.Sprintf("x%d", i)
		vars[name] = p.NewNormalVar(name, 15, 25)
		builder.Add(1, vars[name])
	}
	p.SetObj(builder.Minimize())
	err := p.Solve(&CBCSolver{cbcPath: "/usr/local/bin/cbc"})
	if err != nil {
		t.Error(err)
	}
	if p.status != "Optimal" {
		t.Errorf("expect optimal result, actual: %s", p.status)
	}
	for i := 1; i <= 1000; i++ {
		name := fmt.Sprintf("x%d", i)
		if vars[name].getVal() != 15 {
			t.Errorf("expect x%d == 15, actual: %f", i, vars[name].getVal())
		}
	}
}