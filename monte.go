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
  "io"
  //"bufio"
  "encoding/csv"
  "unsafe"
  "strconv"
  //"strings"
)

func Random() int {
    return int(C.rand())
}

type RandomGenerator struct {
  dsfmt *C.dsfmt_t
}

type WeightSet []float64

func (ws *WeightSet) String() string {
    return fmt.Sprintf("%d", *ws)
}

func (ws *WeightSet) Set(value string) error {
    tmp, err := strconv.ParseFloat(value, 64)
    if err != nil {
        *ws = append(*ws, -1)
    } else {
        *ws = append(*ws, tmp)
    }
    return nil
}

var weights WeightSet

func main() {
  simulations := flag.Int("simulations", 10000, "Number of simulations to run.")
  flag.Var(&weights, "weights", "How we should weight each group")
  flag.Parse()

  reader := csv.NewReader(os.Stdin)
  out := csv.NewWriter(os.Stdout)

  var dsfmt C.dsfmt_t
  C.dsfmt_init_gen_rand(&dsfmt, 1234);
  size := int(unsafe.Sizeof(C.double(12)))
  //http://stackoverflow.com/questions/6942837/how-to-call-this-c-function-from-go-language-with-cgo-tool/6944001#6944001
  randoms := C.memalign(16, C.size_t(size * *simulations))
  defer C.free(randoms)
  r := (*C.double)(randoms)
  var current_sim float64
  count := 0
  // This should be fleshed out a bit with: http://crypto.stanford.edu/~blynn/c2go/ch02.html
  for {
    arr, err := reader.Read()
    if err == io.EOF {
      break
    }
    y0, err := strconv.ParseFloat(arr[1], 64)
    y1, err := strconv.ParseFloat(arr[2], 64)
    y2, err := strconv.ParseFloat(arr[3], 64)
    fmt.Printf("y0 %s\n", y0)
    fmt.Printf("y1 %s\n", y1)
    fmt.Printf("y2 %s\n", y2)


    C.dsfmt_fill_array_close_open(&dsfmt, r, C.int(*simulations));

    for j := 0; j < *simulations; j++ {
      ptr := unsafe.Pointer( uintptr(randoms) + uintptr(size * j) )
      current_sim = *(*float64)(ptr)
      // here we can look up which group this should go to based on weights
      // Get the SimulationSummary for the group, and add y0, y1, y2
      count++
    }
  }
    // TODO: out.Write(fmt.Sprintf("%d, %s", *simulations, line))
    out.Flush()
  fmt.Printf("%s\n", current_sim)
  fmt.Printf("count %d\n", count)
}


