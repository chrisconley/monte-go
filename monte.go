// Usage:
// echo -e "hi\nbye" | (go build monte.go && ./monte --flagname=12) > ./test.txt
package main

import (
  "flag"
  "fmt"
  "os"
  "bufio"
)

func main() {
  var ip = flag.Int("flagname", 1234, "help message for flagname")
  flag.Parse()
  //fmt.Printf("%d\n", *ip)
  reader := bufio.NewReader(os.Stdin)
  out := bufio.NewWriter(os.Stdout)
  // This should be fleshed out a bit with: http://crypto.stanford.edu/~blynn/c2go/ch02.html
  for {
    line, err := reader.ReadString('\n')

    if err != nil {
      // You may check here if err == io.EOF
      break
    }
    out.WriteString(fmt.Sprintf("%d, %s", *ip, line))
    out.Flush()

    //fmt.Printf("%s\n", line)
  }
}


