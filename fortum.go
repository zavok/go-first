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
	
	progstart = 100
	rstackstart = 5
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
	const builtins string = "halt : immediate"
	var first = new(First)
	*first = First{
		NewStack(),
		NewMem(),
		make([]string, 0),
		0, 0,
		true,
		nil,
	}
	first.mem.Store(dep, progstart)
	first.mem.Store(rsp, rstackstart)
	first.in = bufio.NewReader(strings.NewReader(builtins))
	for i := 0; i < 3; i++ {
		first.define()
		first.mem.Store(first.mem.data[dep] - 1, -3)
		first.Compile(i, 10)
	}
	first.in = bufio.NewReader(strings.NewReader(
		"_read @ ! - * / <0 exit echo key _pick _loop"))
	for i:=3 ; i < 14; i++ {
		first.define()
		first.mem.Store(first.mem.data[dep] - 1, i)
	}

	first.define();
	first.mem.Store(first.mem.data[dep] -1, -3)
	first.Compile(3, first.lwp + 3)
	first.pc = first.lwp + 3
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
	if F.mem.data[rsp] >= progstart {
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

func (F *First) define() error {
	var s string
	nwp := F.mem.data[dep]
	if _, err := fmt.Fscan(F.in, &s); err != nil {
		return err
	}
	F.in.ReadRune()
	F.strings = append(F.strings, s)
	id := len(F.strings) - 1
	F.Compile(F.lwp, id, -2)
	F.lwp = nwp
	return nil
}

func (F *First) _read() error {
	var (
		s string
		err error
		num, wp int
	)
	if _, err = fmt.Fscan(F.in, &s); err != nil {
		return err
	}
	F.in.ReadRune()
	switch s {
	case "S":
		fmt.Println(F.stack.data)
		return nil
	case "R":
		fmt.Println(F.mem.data[4:F.mem.data[rsp]])
		return nil
	case "M":
		fmt.Println(F.mem.data[progstart:])
		return nil
	}
	wp = F.findWord(s)
	if wp != 0 {
		wp += 2
		inst, err := F.mem.Fetch(wp)
		switch {
		case inst == -3: // "run me"
			err = F.rpush(F.pc)
			F.pc = wp + 1
		case inst == -2: // "compile me"
			F.Compile(wp + 1)
		case (inst >=0) && (inst <= 13):
			F.Compile(inst)
		default:
			err = fmt.Errorf("invalid code pointer in word %s, %d", s, inst)
		}
		return err
	}
	if num, err = strconv.Atoi(s); err != nil {
		return err
	}
	F.Compile(-1, num)
	return nil
}

func (F *First) Run(input *bufio.Reader) error {
	var (
		err error
		inst, x, y int
	)
	F.in = input
	for err == nil {
		inst, err = F.mem.Fetch(F.pc)
		// fmt.Print(" ", inst)
		F.pc++
		switch inst {
		case -1: // internal builtin "pushint"
			x, err = F.mem.Fetch(F.pc)
			F.stack.Push(x)
			// fmt.Print("=", x)
			F.pc++
		case 0: // builtin "halt"
			fmt.Println(F.pc - 1, "halt")
			F.pc--
			return err
		case 1: // builtin "define", ":"
			err = F.define()
		case 2: // builtin "immediate"
			F.mem.Store(F.mem.data[dep] - 1, -3)
		case 3: // builtin "_read"
			err = F._read()
		case 4: // builtin "fetch", "@"
			x, err = F.stack.Pop()
			y, err = F.mem.Fetch(x)
			F.stack.Push(y)
		case 5: // builtin "store", "!"
			x, err = F.stack.Pop()
			y, err = F.stack.Pop()
			F.mem.Store(x, y)
		case 6: // builtin "minus", "-"
			x, err = F.stack.Pop()
			y, err = F.stack.Pop()
			F.stack.Push(y - x)
		case 7: // builtin "mulitply", "*"
			x, err = F.stack.Pop()
			y, err = F.stack.Pop()
			F.stack.Push(y * x)
		case 8: // builtin "divide", "/"
			x, err = F.stack.Pop()
			y, err = F.stack.Pop()
			F.stack.Push(y / x)
		case 9: // builtin "less than 0", "<0"
			x, err = F.stack.Pop()
			if (x < 0) {
				F.stack.Push(1)
			} else {
				F.stack.Push(0)
			}
		case 10: // builtin "exit"
			F.pc, err = F.rpop()
		case 11: // builtin "echo"
			x, err = F.stack.Pop()
			if err == nil {
				fmt.Printf("%c", x)
			}
		case 12: // builtin "key"
			var r rune
			//k := bufio.NewReader(F.in)
			r, _, err = F.in.ReadRune()
			// fmt.Printf("[%c][%d]\n", r, r)
			F.stack.Push(int(r))
		case 13: // builtin "_pick"
			x, err = F.stack.Pop()
			y = F.stack.data[len(F.stack.data) - x - 1]
			F.stack.Push(y)
		default:
			err = F.rpush(F.pc)
			F.pc = inst
		}
	}
	if err == io.EOF {
		F.pc--
		return nil
	}
	return err
}

func debugout(first *First) {
	fmt.Println("Words:", first.strings)
	fmt.Println("Stack:", first.stack.data)
	fmt.Println("Rstack:", first.mem.data[rstackstart:first.mem.data[rsp]])
	memdump, _ := os.Create("memdump")
	out := bufio.NewWriter(memdump)
	for _, m := range first.mem.data {
		fmt.Fprintln(out, m)
	}
	out.Flush()
	memdump.Close()
}

func main() {
	first, err := NewFirst()

	defer debugout(first)

	if err != nil {
		fmt.Println(err)
		return
	}
	err = first.Run(bufio.NewReader(strings.NewReader(third)))
	fmt.Println("---")
	if err != nil {
		fmt.Println(err)
	} else if err = first.Run(bufio.NewReader(os.Stdin)); err != nil {
			fmt.Println(err)
	}
}
