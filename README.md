### MILP Solver for Go

#### TODO List

[x] basic wrapper for large problem

[ ] custom cbc-solver commandline

[ ] other solver not only CBCSolver

[ ] high-level API calculate complex operators like ()

#### Usage

```go
cbc := &CBCSolver{cbcPath: "/usr/local/bin/cbc"}
p := NewProblem("a")
x1 := p.NewNormalVar("x1", 0, 40)
x2 := p.NewNormalVar("x2", math.Inf(-1), math.Inf(1))
x3 := p.NewNormalVar("x3", math.Inf(-1), math.Inf(1))
x4 := p.NewIntegerVar("x4", 2, 3)
p.SetObj(NewObjBuilder().Add(1, x1).Add(1, x2).Add(3, x3).Add(1, x4).Maximize())
p.AddST(NewSTBuilder("c1").Add(-1, x1).Add(1, x2).Add(1, x3).Add(10, x4).Leq().AddConst(40).Done())
p.AddST(NewSTBuilder("c2").Add(1, x1).Add(-3, x2).Add(1, x3).Leq().AddConst(30).Done())
p.AddST(NewSTBuilder("c3").Add(1, x2).Add(-3.5, x4).Eq().AddConst(0).Done())

err := p.Solve(cbc)
if err != nil {
    panic(err)
}
fmt.Println(x1.getVal()) // get the val of variable x1
```