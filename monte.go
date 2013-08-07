// Usage:
// echo -e "hi\nbye" | (go build monte.go && ./monte --flagname=12) > ./test.txt
package main

/*
#cgo CFLAGS: -Wall -O3 -msse2 -DHAVE_SSE2 -DDSFMT_MEXP=19937
#include <malloc.h>
#include <stdio.h>
#include <errno.h>
#include <stdlib.h>
#include "dSFMT-src-2.2.1/dSFMT.c"
*/
import "C"

import (
  "flag"
  "fmt"
  "os"
  "bufio"
  "unsafe"
)

func Random() int {
    return int(C.rand())
}

type RandomGenerator struct {
  dsfmt *C.dsfmt_t
}

func main() {
  var dsfmt C.dsfmt_t
  fmt.Printf("%s\n", dsfmt)
  C.dsfmt_init_gen_rand(&dsfmt, 1234);
  size := int(unsafe.Sizeof(C.double(12)))
  fmt.Printf("size: %d\n", size)
  simulations := int(500)
  randoms := C.memalign(16, C.size_t(size * simulations))
  defer C.free(randoms)
  r := (*C.double)(randoms)
  C.dsfmt_fill_array_close_open(&dsfmt, r, C.int(simulations));
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


