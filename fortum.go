package main

import (
	"fmt"
	"os"
)

type Word struct {
	name string
	builtin func()
	code []int
}

var stack = make([]int, 0)
var mem = make([]int, 4096)
var dict = make([]Word, 0)

const rsp = 0

func builtin(dict []Word, name string, f func()) []Word {
	word := Word{name, f, nil}
	dict = append(dict, word)
	return dict
}

func TOS(stack []int) int {
	return stack[len(stack) - 1]
}

func pop(stack *[]int) int {
	x := TOS(*stack)
	s := *stack
	*stack = s[:len(s) - 1]
	return x
}

func noop() {
}

func minus() {
	stack[len(stack) - 2] -= TOS(stack)
	pop(&stack)
}

func multiply() {
	stack[len(stack) - 2] *= TOS(stack)
	pop(&stack)
}

func divide() {
	stack[len(stack) - 2] /= TOS(stack)
	pop(&stack)
}

func lessthan0() {
	x := TOS(stack)
	if (x < 0) {
		stack[len(stack) - 1] = 1
	} else {	
		stack[len(stack) - 1] = 0
	}
}

func echo() {
	x := pop(&stack)
	fmt.Printf("%c", x)
}

func key() {
	mem[rsp]++
	mem[mem[rsp]] = os.Stdin.ReadRune()
}

func execword() {
	word := dict[mem[mem[rsp]]]
	if (word.builtin != nil) {
		word.builtin()
	} else {
		for w := range word.code {
			mem[rsp]++
			mem[mem[rsp]] = w
			execword()
		}
	}
	mem[rsp]--
}

func store() {
	addr := TOS(stack)
	pop(&stack)
	if len(mem) < addr {
		bump := make([]int, addr - len(mem) + 1)
		mem = append(mem, bump...)
	}
	mem[addr] = TOS(stack)
}

func fetch() {
	addr := TOS(stack)
	if addr > len(mem) {
		stack[len(stack)-1] = 0
	} else {
		stack[len(stack)-1] = mem[addr]
	}
}

func main() {
	dict = builtin(dict, "noop", noop)
	dict = builtin(dict, "execword", execword)
	dict = builtin(dict, "-", minus)
	dict = builtin(dict, "*", multiply)
	dict = builtin(dict, "/", divide)
	dict = builtin(dict, "echo", echo)
	dict = builtin(dict, "key", key)
	dict = builtin(dict, "@", fetch)
	dict = builtin(dict, "!", store)

	stack = append(stack, 10, 79, 76, 76, 69, 72)
	mem[rsp] = 1
	execword()
	echo()
	echo()
	echo()
	echo()
	echo()
	echo()
	fmt.Println(stack)
}
