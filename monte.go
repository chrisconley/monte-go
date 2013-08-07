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
  simulations := *flag.Int("simulations", 10000, "Number of simulations to run.")
  iterations := *flag.Int("iterations", 10000, "Number of iterations to run - temporary") // this doesn't seem to be working - always uses default
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
    out.WriteString(fmt.Sprintf("%d, %s", simulations, line))
    out.Flush()

    //fmt.Printf("%s\n", line)
  }

  var dsfmt C.dsfmt_t
  fmt.Printf("%s\n", dsfmt)
  C.dsfmt_init_gen_rand(&dsfmt, 1234);
  size := int(unsafe.Sizeof(C.double(12)))
  fmt.Printf("size: %d\n", size)
  //http://stackoverflow.com/questions/6942837/how-to-call-this-c-function-from-go-language-with-cgo-tool/6944001#6944001
  randoms := C.memalign(16, C.size_t(size * simulations))
  defer C.free(randoms)
  r := (*C.double)(randoms)
  var current_sim float64
  count := 0
  for i := 0; i < iterations; i++ {
    C.dsfmt_fill_array_close_open(&dsfmt, r, C.int(simulations));

    for j := 0; j < simulations; j++ {
      ptr := unsafe.Pointer( uintptr(randoms) + uintptr(size * j) )
      current_sim = *(*float64)(ptr)
      count++
    }
  }
  fmt.Printf("%s\n", current_sim)
  fmt.Printf("count %d\n", count)
}


