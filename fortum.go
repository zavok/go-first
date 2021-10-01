package main

import "fmt"

type Word struct {
	name string
	builtin func()
	code []int
}

var stack = make([]int, 0)
var rstack = make([]int, 0)
var mem = make([]int, 0)
var dict = make([]Word, 0)

func builtin(dict []Word, name string, f func()) []Word {
	word := Word{name, f, nil}
	dict = append(dict, word)
	return dict
}

func TOS(stack []int) int {
	return stack[len(stack) - 1]
}

func noop() {
}

func execword() {
	word := dict[TOS(rstack)]
	if (word.builtin != nil) {
		word.builtin()
	} else {
		for w := range word.code {
			rstack = append(rstack, w)
			execword()
		}
	}
	rstack = rstack[:len(rstack) - 1]
}

func store() {
	addr := TOS(stack)
	stack = stack[:len(stack)-1]
	if len(mem) < addr {
		bump := make([]int, addr - len(mem) + 1)
		mem = append(mem, bump...)
	}
	mem[addr] = TOS(stack)
}

func fetch() {
	addr := TOS(stack)
	if (addr > len(mem) {
		stack[len(stack)-1] = 0
	} else {
		stack[len(stack)-1] = mem[addr]
	}
}

func main() {
	dict = builtin(dict, "noop", noop)
	dict = builtin(dict, "execword", execword)
	stack = append(stack, 1, 2, 3, 4)
	rstack = append(rstack, 0)
	execword()
	store()
	fetch()
	fmt.Println(mem)
	fmt.Println(stack)
}
