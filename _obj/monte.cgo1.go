// Created by cgo - DO NOT EDIT

//line monte.go:3
package main
//line monte.go:14

//line monte.go:13
import (
	"flag"
	"fmt"
	"os"
	"bufio"
)
//line monte.go:21

//line monte.go:20
func Random() int {
	return int(_Cfunc_rand())
}
//line monte.go:29

//line monte.go:28
func main() {
//line monte.go:31

//line monte.go:30
	var ip = flag.Int("flagname", 1234, "help message for flagname")
			flag.Parse()
//line monte.go:34

//line monte.go:33
	reader := bufio.NewReader(os.Stdin)
			out := bufio.NewWriter(os.Stdout)
//line monte.go:37

//line monte.go:36
	for {
				line, err := reader.ReadString('\n')
//line monte.go:40

//line monte.go:39
		if err != nil {
//line monte.go:42

//line monte.go:41
			break
		}
		out.WriteString(fmt.Sprintf("%d, %s", *ip, line))
		out.Flush()
//line monte.go:48

//line monte.go:47
	}
}
