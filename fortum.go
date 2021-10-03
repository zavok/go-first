package main

import (
	"fmt"
	"bufio"
	"os"
	"io"
	"strconv"
	"strings"
)

const rsp int = 0
const dep int = 1

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
	first.mem.Store(rsp, 4)
	first.in = strings.NewReader(
		"halt : immediate")
	for i := 0; i < 3; i++ {
		first.define()
		first.mem.Store(first.mem.Fetch(dep) - 1, -3)
		first.Compile(i, 4149)
	}
	first.in = strings.NewReader(
		"_read @ ! - * / <0 exit echo key _pick _loop")
	for i:=3 ; i < 14; i++ {
		first.define()
		first.Compile(i, 4149)
	}

	// 4149 should be a pointer to exit as it is defined here

	first.define();
	first.mem.Store(first.mem.Fetch(dep) -1, -3)
	first.Compile(3, first.lwp + 3)
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
	F.mem.Store(F.mem.Fetch(rsp), val)
	F.mem.data[rsp]++	
	return nil
}

func (F *First) rpop() (int, error) {
	if F.mem.Fetch(rsp) <= 4 {
		return 0, fmt.Errorf("rstack is empty")
	}
	F.mem.data[rsp]--
	x := F.mem.Fetch(F.mem.Fetch(rsp))
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
	F.strings = append(F.strings, s)
	id := len(F.strings) - 1
	F.Compile(F.lwp, id, -2)
	F.lwp = nwp
	return nil
}

func (F *First) _read() error {
	var s string
	_, err := fmt.Fscan(F.in, &s)
	if err != nil {
		return err
	}
	wp := F.findWord(s)
	if wp != 0 {
		err = F.rpush(F.pc)
		F.pc = wp + 2
		return err
	}
	num, err := strconv.Atoi(s)
	if err != nil {
		return err
	}
	F.Compile(-1, num)
	return nil
}

func (F *First) Run(input io.Reader) error {
	var err error
	F.in = input
	F.pc = F.findWord("_loop") + 3 
	for err == nil {
		var x, y int
		inst := F.mem.Fetch(F.pc)
		F.pc++
		switch inst {
		case -1: // internal builtin "pushint"
			F.stack.Push(F.mem.Fetch(F.pc))
			F.pc++
		case -2: // internal builtin "compile me"
			/* "a pointer to the word's data field is
			 * appended to the dictionary" ????
			 */
			F.Compile(F.pc)
			F.pc, err = F.rpop();
		case -3: // internal builtin "run me"
			/* "the word's data field is taken to be
			 * a stream of pointers to words, and is
			 * executed" ????
			 */
		case 0: // builtin "halt"
			fmt.Println(F.pc - 1, "halt")
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
			/* We rpop twice because we need
			 * to exit "exit" as well
			 */
			_, err = F.rpop()
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
	var code string = ": L immediate exit L"
	err = first.Run(strings.NewReader(code))
	if err != nil {
		fmt.Println(err)
	} else {
		err = first.Run(os.Stdin)
		if err != nil {
			fmt.Println(err)
		}
	}
//	fmt.Println("Words:", first.strings)
//	fmt.Println("Stack:", first.stack.data)
//	fmt.Println(first.mem.data[4096:])
}
