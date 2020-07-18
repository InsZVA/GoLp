package goLp

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
)

const (
	objectiveMinimize = 1
	objectiveMaximize = 2

	opNone = 0
	opGeq  = 2
	opLeq  = 4
	opEq   = 5

	varNormal  = 0
	varGeneral = 1
	varBinary  = 2
)

type InvalidST struct {
	error
}

func NewInvalidSTError(stName string) InvalidST {
	return InvalidST{error: fmt.Errorf("invalid subject to: %s", stName)}
}

var BuilderDoneErr = fmt.Errorf("builder has done")
var DuplicateVariableErr = fmt.Errorf("dunplicate variable")
var NoObjErr = fmt.Errorf("no objective error")
var VariableNameTooLongErr = fmt.Errorf("variable name too long(>200)")

type ST struct {
	name string
	err error
	statement map[string]float64
	op int64
	c float64
}

type Obj struct {
	objective int64
	objFunc map[string]float64
	c float64
}

type Problem struct {
	name string
	status string
	err error
	opt *Obj
	sts []ST
	vars map[string]*Variable
}

func NewProblem(name string) *Problem {
	return &Problem{name: name, vars: make(map[string]*Variable), status: "Not solved"}
}

type Variable struct {
	name string
	lowerLimit float64
	upperLimit float64
	tp int64
	val float64
}

// returns 0 before problem solved
// need to cast if not normal
func (v *Variable) getVal() float64 {
	return v.val
}

func (p *Problem) NewNormalVar(name string, lowerLimit float64, upperLimit float64) *Variable {
	return p.newVariable(name, lowerLimit, upperLimit, varNormal)
}

func (p *Problem) NewIntegerVar(name string, lowerLimit float64, upperLimit float64) *Variable {
	return p.newVariable(name, lowerLimit, upperLimit, varGeneral)
}

func (p *Problem) NewBinaryVar(name string, lowerLimit float64, upperLimit float64) *Variable {
	return p.newVariable(name, lowerLimit, upperLimit, varBinary)
}

func (p *Problem) newVariable(name string, lowerLimit float64, upperLimit float64, tp int64) *Variable {
	if len(name) > 200 {
		p.err = VariableNameTooLongErr
	}
	strings.Replace(name, " ", "_", -1)
	if v, ok := p.vars[name]; ok {
		p.err = DuplicateVariableErr
		return v
	}
	p.vars[name] = &Variable{
		name: name,
		lowerLimit: lowerLimit,
		upperLimit: upperLimit,
		tp: tp,
	}
	return p.vars[name]
}

func (p *Problem) AddST(st ST) *Problem {
	p.sts = append(p.sts, st)
	return p
}

func (p *Problem) SetObj(obj *Obj) {
	p.opt = obj
}

func (p *Problem) Solve(solver Solver) error {
	if p.err != nil {
		return p.err
	}
	if p.opt == nil {
		return NoObjErr
	}
	for _, st := range p.sts {
		if st.err != nil {
			return st.err
		}
	}
	lpFile, err := p.generateLPFile()
	if err != nil {
		return err
	}
	res, err := solver.Solve(lpFile, &p.status)
	if err != nil {
		return err
	}
	for name, val := range res {
		if v, ok := p.vars[name]; ok {
			v.val = val
		}
	}
	return nil
}

func (p *Problem) Status() string {
	return p.status
}

type Generator struct {
	writer io.Writer
	line []byte
}

func NewGenerator(w io.Writer) *Generator {
	return &Generator{writer: w, line: make([]byte, 0)}
}

func (g *Generator) NextLine() error {
	if len(g.line) == 0 {
		return nil
	}
	g.line = append(g.line, '\n')
	_, err := g.writer.Write(g.line)
	if err != nil {
		return err
	}
	g.line = make([]byte, 0)
	return nil
}

func (g *Generator) Append(word string) error {
	if len(g.line) + len(word) > 250 {
		err := g.NextLine()
		if err != nil {
			return err
		}
		return g.Append(word)
	}
	g.line = append(g.line, word...)
	return nil
}

func (g *Generator) Flush() error {
	err := g.NextLine()
	if err != nil {
		return err
	}
	return nil
}

func (p *Problem) generateObj(writer *Generator) error {
	var toWrite string
	var err error
	if p.opt.objective == objectiveMaximize {
		toWrite = "Maximize"
	} else {
		toWrite = "Minimize"
	}
	err = writer.Append(toWrite)
	if err != nil {
		return err
	}
	err = writer.NextLine()
	if err != nil {
		return err
	}
	err = writer.Append("obj: ")
	if err != nil {
		return err
	}

	first := true
	for name, co := range p.opt.objFunc {
		if first {
			err = writer.Append(fmt.Sprintf("%f %s", co, name))
			if err != nil {
				return err
			}
			first = false
		} else {
			if co >= 0 {
				err = writer.Append(fmt.Sprintf(" + %f %s", co, name))
			} else {
				err = writer.Append(fmt.Sprintf(" %f %s", co, name))
			}
			if err != nil {
				return err
			}
		}
	}
	if p.opt.c >= 0 {
		err = writer.Append(fmt.Sprintf(" + %f", p.opt.c))
	} else {
		err = writer.Append(fmt.Sprintf(" %f", p.opt.c))
	}
	if err != nil {
		return err
	}
	err = writer.NextLine()
	if err != nil {
		return err
	}
	return nil
}

func (p *Problem) generateST(writer *Generator) error {
	var id int
	var err error

	err = writer.Append("Subject To")
	if err != nil {
		return err
	}
	err = writer.NextLine()
	if err != nil {
		return err
	}

	for _, st := range p.sts {
		err = writer.Append(fmt.Sprintf("c%d: ", id))
		if err != nil {
			return err
		}
		id++
		first := true
		for name, co := range st.statement {
			if first {
				err = writer.Append(fmt.Sprintf("%f %s", co, name))
				first = false
			} else {
				if co >= 0 {
					err = writer.Append(fmt.Sprintf(" + %f %s", co, name))
				} else {
					err = writer.Append(fmt.Sprintf(" %f %s", co, name))
				}
			}
			if err != nil {
				return err
			}
		}
		if st.op == opEq {
			err = writer.Append(" = ")
		} else if st.op == opGeq {
			err = writer.Append(" >= ")
		} else if st.op == opLeq {
			err = writer.Append(" <= ")
		}
		if err != nil {
			return err
		}
		err = writer.Append(fmt.Sprintf("%f", st.c))
		if err != nil {
			return err
		}
		err = writer.NextLine()
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *Problem) generateBounds(writer *Generator) error {
	var err error

	err = writer.Append("Bounds")
	if err != nil {
		return err
	}
	err = writer.NextLine()
	if err != nil {
		return err
	}

	for name, v := range p.vars {
		err = writer.Append(fmt.Sprintf("%f <= %s <= %f\n", v.lowerLimit, name, v.upperLimit))
		if err != nil {
			return err
		}
		err = writer.NextLine()
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *Problem) generateVars(writer *Generator) error {
	var err error

	normal := make([]*Variable, 0)
	general := make([]*Variable, 0)
	binary := make([]*Variable, 0)

	for _, v := range p.vars {
		if v.tp == varGeneral {
			general = append(general, v)
		} else if v.tp == varBinary {
			binary = append(binary, v)
		} else {
			normal = append(normal, v)
		}
	}

	if len(general) > 0 {
		err = writer.Append("General")
		if err != nil {
			return err
		}
		err = writer.NextLine()
		if err != nil {
			return err
		}

		for _, v := range general {
			err = writer.Append(" " + v.name)
			if err != nil {
				return err
			}
		}
		err = writer.NextLine()
		if err != nil {
			return err
		}
	}

	if len(binary) > 0 {
		err = writer.Append("Binary")
		if err != nil {
			return err
		}
		err = writer.NextLine()
		if err != nil {
			return err
		}

		for _, v := range general {
			err = writer.Append(" " + v.name)
			if err != nil {
				return err
			}
		}
		err = writer.NextLine()
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *Problem) generateLPFile() (string, error) {
	f, err := ioutil.TempFile("/tmp", p.name + "*.lp")
	if err != nil {
		return "", err
	}
	bufW := bufio.NewWriter(f)
	writer := NewGenerator(bufW)
	if err = p.generateObj(writer); err != nil {
		return "", err
	}
	if err = p.generateST(writer); err != nil {
		return "", err
	}
	if err = p.generateBounds(writer); err != nil {
		return "", err
	}
	if err = p.generateVars(writer); err != nil {
		return "", err
	}
	if err = writer.Append("End"); err != nil {
		return "", err
	}
	if err = writer.NextLine(); err != nil {
		return "", err
	}
	if err = writer.Flush(); err != nil {
		return "", err
	}
	if err = bufW.Flush(); err != nil {
		return "", err
	}
	name := f.Name()
	if err = f.Close(); err != nil {
		return "", err
	}
	return name, nil
}

type STBuilder struct {
	name string
	err error
	op int64
	c float64
	statement map[string]float64
}

func NewSTBuilder(name string) *STBuilder {
	return &STBuilder{name: name, statement: make(map[string]float64)}
}

func (builder *STBuilder) Add(coefficient float64, v *Variable) *STBuilder {
	if builder.op != opNone {
		coefficient = -coefficient
	}
	builder.statement[v.name] += coefficient
	return builder
}

func (builder *STBuilder) AddConst(val float64) *STBuilder {
	if builder.op != opNone {
		builder.c += val
	} else {
		builder.c -= val
	}
	return builder
}

func (builder *STBuilder) Geq() *STBuilder {
	if builder.op != opNone {
		builder.err = NewInvalidSTError(builder.name)
		return builder
	}
	builder.op = opGeq
	return builder
}

func (builder *STBuilder) Leq() *STBuilder {
	if builder.op != opNone {
		builder.err = NewInvalidSTError(builder.name)
		return builder
	}
	builder.op = opLeq
	return builder
}

func (builder *STBuilder) Eq() *STBuilder {
	if builder.op != opNone {
		builder.err = NewInvalidSTError(builder.name)
		return builder
	}
	builder.op = opEq
	return builder
}

func (builder *STBuilder) Done() ST {
	if builder.op == opNone {
		builder.err = NewInvalidSTError(builder.name)
	}
	res := ST{
		name: builder.name,
		err: builder.err,
		statement: builder.statement,
		c: builder.c,
		op: builder.op,
	}
	builder.err	= BuilderDoneErr
	return res
}

type ObjBuilder struct {
	objective int64
	objFunc map[string]float64
	c float64
}

func NewObjBuilder() *ObjBuilder {
	return &ObjBuilder{objFunc: make(map[string]float64)}
}

func (builder *ObjBuilder) Add(coefficient float64, v *Variable) *ObjBuilder {
	builder.objFunc[v.name] += coefficient
	return builder
}

func (builder *ObjBuilder) AddConst(val float64) *ObjBuilder {
	builder.c += val
	return builder
}

func (builder *ObjBuilder) Maximize() *Obj {
	return &Obj{
		objective: objectiveMaximize,
		objFunc: builder.objFunc,
		c: builder.c,
	}
}

func (builder *ObjBuilder) Minimize() *Obj {
	return &Obj{
		objective: objectiveMinimize,
		objFunc: builder.objFunc,
		c: builder.c,
	}
}
