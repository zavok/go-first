package main

import (
	"fmt"
	"bufio"
	"os"
	"io"
	"strconv"
	"strings"
)

const (
	rsp int = 1
	dep int = 0
	
	rstacksize  = 512
	rstackstart = 32
)

var laswor string

type Mem struct {
	data []int
}

func NewMem() *Mem {
	var mem = new(Mem)
	mem.data = make([]int, 0)
	return mem
}

func (m *Mem) Fetch(addr int) (int, error) {
	if (addr >= len(m.data)) || (addr < 0) {
		return 0, fmt.Errorf("memfetch: addr %d out of bonds", addr)
	}
	return m.data[addr], nil
}

func (m *Mem) Store(addr, val int) {
	if addr >= len(m.data) {
		buf := make([]int, addr - len(m.data) + 1)
		m.data = append(m.data, buf...)
	}
	m.data[addr] = val
}

type Stack struct {
	data []int
}

func NewStack() *Stack {
	stack := new(Stack)
	stack.data = make([]int, 1)
	return stack
}

func (s *Stack) Push(val int) {
	s.data = append(s.data, val)
}

func (s *Stack) Pop() (int, error) {
	if s.data == nil {
		return 0, fmt.Errorf("stack is empty")
	}
	x := s.data[len(s.data) - 1]
	s.data = s.data[:len(s.data) - 1]
	return x, nil
}

func (s *Stack) TOS() *int {
	return &s.data[len(s.data) - 1]
}

type First struct {
	stack *Stack
	mem *Mem
	strings []string
	pc, lwp int
	run bool
	in *bufio.Reader
}

func NewFirst() (*First, error) {
	const builtins string = ": immediate _read @ ! - * / <0 exit echo key _pick"
	first := new(First)
	*first = First{
		NewStack(),
		NewMem(),
		make([]string, 0),
		0, 0,
		true,
		nil,
	}
	first.mem.Store(dep, rstackstart + rstacksize)
	first.mem.Store(rsp, rstackstart)
	first.in = bufio.NewReader(strings.NewReader(builtins))

	first.define(3)
	first.define(4)
	first.define(1)
	loopword := first.mem.data[dep]
	first.Compile(5, 2)
	first.pc = first.mem.data[dep]
	first.Compile(loopword, first.pc - 1)
	for i := 6; i < 16; i++ {
		first.define(1)
		first.Compile(i)
	}
	return first, nil
}

func (F *First) findWord(s string) int {
	var err error
	for wp := F.lwp; wp != 0; wp, err = F.mem.Fetch(wp) {
		if err != nil {
			return 0
		}
		id, err := F.mem.Fetch(wp + 1)
		if err != nil {
			return 0
		}
		if F.strings[id] == s {
			return wp
		}
	}
	return 0
}

func (F *First) rpush(val int) (error) {
	if F.mem.data[rsp] >= rstacksize + rstackstart {
		return fmt.Errorf("rstack is full")
	}
	F.mem.data[rsp]++
	F.mem.Store(F.mem.data[rsp], val)
	return nil
}

func (F *First) rpop() (int, error) {
	if F.mem.data[rsp] <= rstackstart {
		return 0, fmt.Errorf("rstack is empty")
	}
	x, err := F.mem.Fetch(F.mem.data[rsp])
	F.mem.data[rsp]--
	return x, err
}

func (F *First) Compile(vals ...int) {
	for _, val := range vals {
		F.mem.Store(F.mem.data[dep], val)
		F.mem.data[dep]++
	}
}

func (F *First) define(code int) error {
	var s string
	F.Compile(F.lwp)
	F.lwp = F.mem.data[dep] - 1
	F.Compile(len(F.strings), code)
	if _, err := fmt.Fscan(F.in, &s); err != nil {
		return err
	}
	F.strings = append(F.strings, s)
	return nil
}

func (F *First) _read() error {
	var (
		s string
		num, wp int
		err error
	)
	if _, err = fmt.Fscan(F.in, &s); err != nil {
		return err
	}
	wp = F.findWord(s)
	if wp != 0 {
		return F.Step(wp + 2)
	}
	if num, err = strconv.Atoi(s); err != nil {
		return err
	}
	F.Compile(2, num)
	return nil
}

func (F *First) Run(input *bufio.Reader) error {
	F.in = input
	var (
		addr int
		err error
	)
	for err == nil {
		addr, err = F.mem.Fetch(F.pc)
		F.pc++
		if err != nil {
			break
		}
		err = F.Step(addr)
	} 
	if err == io.EOF {
		return nil
	}
	return err
}

func (F *First) Step(addr int) error {
	var x, y int
	inst, err := F.mem.Fetch(addr)
	addr++
	switch inst {
	case 0: // internal builtin "pushint"
		x, err = F.mem.Fetch(F.pc)
		F.pc++
		F.stack.Push(x)
	case 1: // compile
		F.Compile(addr)
	case 2: // run
		F.rpush(F.pc)
		F.pc = addr
	case 3: // builtin "define", ":"
		err = F.define(1)
		F.Compile(2)
	case 4: // builtin "immediate"
		F.mem.data[dep] -= 2
		F.Compile(2)
	case 5: // builtin "_read"
		err = F._read()
	case 6: // builtin "fetch", "@"
		x, err = F.stack.Pop()
		y, err = F.mem.Fetch(x)
		F.stack.Push(y)
	case 7: // builtin "store", "!"
		x, err = F.stack.Pop()
		y, err = F.stack.Pop()
		F.mem.Store(x, y)
	case 8: // builtin "minus", "-"
		x, err = F.stack.Pop()
		y, err = F.stack.Pop()
		F.stack.Push(y - x)
	case 9: // builtin "mulitply", "*"
		x, err = F.stack.Pop()
		y, err = F.stack.Pop()
		F.stack.Push(y * x)
	case 10: // builtin "divide", "/"
		x, err = F.stack.Pop()
		y, err = F.stack.Pop()
		F.stack.Push(y / x)
	case 11: // builtin "less than 0", "<0"
		x, err = F.stack.Pop()
		if (x < 0) {
			F.stack.Push(1)
		} else {
			F.stack.Push(0)
		}
	case 12: // builtin "exit"
		F.pc, err = F.rpop()
	case 13: // builtin "echo"
		x, err = F.stack.Pop()
		if err == nil {
			fmt.Printf("%c", x)
		}
	case 14: // builtin "key"
		var r rune
		r, _, err = F.in.ReadRune()
		F.stack.Push(int(r))
	case 15: // builtin "_pick"
		x, err = F.stack.Pop()
		y = F.stack.data[len(F.stack.data) - x - 1]
		F.stack.Push(y)
	default:
		err = fmt.Errorf("unexpected instruction %d", inst)
	}
	return err
}

func main() {
	first, err := NewFirst()
	if err != nil {
		fmt.Println(err)
	} else if err = first.Run(bufio.NewReader(os.Stdin));
	  err != nil {
		fmt.Println(err)
	}
}
