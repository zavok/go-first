package main

import (
	"fmt"
	"bufio"
	"os"
	"io"
	"strconv"
	"strings"
)

const rsp int = 1
const dep int = 0

type Mem struct {
	data []int
}

func NewMem() *Mem {
	var mem = new(Mem)
	mem.data = make([]int, 0)
	return mem
}

func (m *Mem) Fetch(addr int) int {
	if addr >= len(m.data) {
		return 0
	}
	return m.data[addr]
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
	stack.data = make([]int, 0)
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
	in io.Reader
}

func NewFirst() (*First, error) {
	var first = new(First)
	*first = First{
		NewStack(),
		NewMem(),
		make([]string, 0),
		0, 0,
		true,
		nil,
	}
	first.mem.Store(dep, 4096)
	first.mem.Store(rsp, 10)
	first.in = strings.NewReader(
		"halt : immediate")
	for i := 0; i < 3; i++ {
		first.define()
		first.mem.Store(first.mem.Fetch(dep) - 1, -3)
		first.Compile(i, 10)
	}
	first.in = strings.NewReader(
		"_read @ ! - * / <0 exit echo key _pick _loop")
	for i:=3 ; i < 14; i++ {
		first.define()
		first.mem.Store(first.mem.Fetch(dep) - 1, i)
	}

	// 4149 should be a pointer to exit as it is defined here

	first.define();
	first.mem.Store(first.mem.Fetch(dep) -1, -3)
	first.Compile(3, first.lwp + 3)
	first.pc = first.lwp + 3
	return first, nil
}

func (F *First) findWord(s string) int {
	for wp := F.lwp; wp != 0; wp = F.mem.Fetch(wp) {
		id := F.mem.Fetch(wp + 1)
		if F.strings[id] == s {
			return wp
		}
	}
	return 0
}

func (F *First) rpush(val int) (error) {
	if F.mem.Fetch(rsp) >= 4096 {
		return fmt.Errorf("rstack is full")
	}
	F.mem.data[rsp]++	
	F.mem.Store(F.mem.Fetch(rsp), val)
	return nil
}

func (F *First) rpop() (int, error) {
	if F.mem.Fetch(rsp) <= 10 {
		return 0, fmt.Errorf("rstack is empty")
	}
	x := F.mem.Fetch(F.mem.Fetch(rsp))
	F.mem.data[rsp]--
	return x, nil
}

func (F *First) Compile(vals ...int) {
	for _, val := range vals {
		F.mem.Store(F.mem.Fetch(dep), val)
		F.mem.data[dep]++
	}
}

func (F *First) define() error {
	var s string
	nwp := F.mem.Fetch(dep)
	_, err := fmt.Fscan(F.in, &s)
	if err != nil {
		return err
	}
	fmt.Printf("<%s> ", s)
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
	fmt.Printf("%s ", s)
	switch s {
	case "S":
		fmt.Println(F.stack.data)
		return nil
	case "R":
		fmt.Println(F.mem.data[4:F.mem.Fetch(rsp)])
		return nil
	case "M":
		fmt.Println(F.mem.data[4096:])
		return nil
	}
	wp = F.findWord(s)
	if wp != 0 {
		wp += 2
		inst := F.mem.Fetch(wp)
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

func (F *First) Run(input io.Reader) error {
	var err error
	F.in = input
	for err == nil {
		var x, y int
		inst := F.mem.Fetch(F.pc)
		F.pc++
		switch inst {
		case -1: // internal builtin "pushint"
			F.stack.Push(F.mem.Fetch(F.pc))
			F.pc++
		case 0: // builtin "halt"
			fmt.Println(F.pc - 1, "halt")
			F.pc--
			return nil
		case 1: // builtin "define", ":"
			err = F.define()
		case 2: // builtin "immediate"
			F.mem.Store(F.mem.Fetch(dep) - 1, -3)
		case 3: // builtin "_read"
			err = F._read()
		case 4: // builtin "fetch", "@"
			x, err = F.stack.Pop()
			y = F.mem.Fetch(x)
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
			k := bufio.NewReader(F.in)
			r, _, err = k.ReadRune()
			F.stack.Push(int(r))
		case 13: // builtin "_pick"
			x, err = F.stack.Pop()
			y = F.stack.data[x]
			F.stack.Push(y)
		default:
			err = F.rpush(F.pc)
			F.pc = inst
		}
	}
	if err == io.EOF {
		return nil
	}
	return err
}

func main() {
	first, err := NewFirst()
	if err != nil {
		fmt.Println(err)
	}
	err = first.Run(strings.NewReader(third))
	if err != nil {
		fmt.Println(err)
		return
	}
	if err = first.Run(bufio.NewReader(os.Stdin)); err != nil {
		fmt.Println(err)
	}
//	fmt.Println("Words:", first.strings)
//	fmt.Println("Stack:", first.stack.data)
//	fmt.Println(first.mem.data[4096:])
}
