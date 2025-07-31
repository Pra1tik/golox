# Lox Interpreter in Go

This is a **tree-walking interpreter** for the [Lox programming language](https://craftinginterpreters.com/), written in Go.  

---

## Features

- Complete implementation of Lox:
  - Variables, scopes
  - First-class functions & closures
  - Classes and single inheritance
  - Method binding and `this`
  - Static resolution of variables
- Error reporting with line numbers
- REPL and script execution
- Written idiomatically in Go

---

## Setup

### Installation

Clone the repository:

```bash
git clone https://github.com/Pra1tik/golox.git
cd golox
```

Build the binary:

```bash
go build -o golox
```

This will produce an executable called `golox`.

---

## Usage

### Run a Lox program

```bash
./golox examples/hello.lox
```

### Start an interactive REPL

```bash
./golox
```

Type Lox expressions and statements directly:

```golox
> var name = "Lox";
> print "Hello, " + name + "!";
Hello, Lox!
```

---

## 📄 Sample Programs

### Hello World
```golox
print "Hello, world!";
```

### Variables and Expressions
```golox
var a = 10;
var b = 20;
print a + b; // 30
```

### Functions and Recursion
```golox
fun fib(n) {
  if (n <= 1) return n;
  return fib(n - 1) + fib(n - 2);
}

print fib(10); // 55
```

### Closures
```golox
fun makeCounter() {
  var count = 0;
  fun counter() {
    count = count + 1;
    print count;
  }
  return counter;
}

var counter = makeCounter();
counter(); // 1
counter(); // 2
```

### Classes and Inheritance
```golox
class Animal {
  speak() {
    print "Some sound";
  }
}

class Dog < Animal {
  speak() {
    print "Woof!";
  }
}

var d = Dog();
d.speak(); // Woof!
```

---

## 🗂 Project Structure

```
.
├── main.go              # Program entry point
├── lexer/               # Tokenizer
├── parser/              # AST generator
├── ast/                 # Expression and statement definitions
├── interpret/           # Tree-walking evaluator
├── resolve/             # Variable resolution and scope checker
├── examples/            # Sample Lox programs

```

