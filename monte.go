// Usage:
// go build monte.go
// time echo -e "a,1,1,1\nb,2,10,100" | (./monte --simulations=10000 --weights=1 --weights=2)
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

type Summary struct {
  y0 float64
  y1 float64
  y2 float64
}

type SimulationSummaries [][]*Summary

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

func getAssignment (weightDistribution []float64 sim float64) int {
  assignment := 0
  for assignment < len(weightDistribution) {
    if sim < weightDistribution[assignment] {
      break
    }
    assignment++
  }
  return assignment
}

func calculateWeightDistribution (weights []float64) []float64 {
  totalWeight := 0.0
  for _, weight := range weights {
    totalWeight += weight
  }

  weightDistribution := []float64{}
  runningWeight := 0.0
  for _, weight := range weights {
    runningWeight += weight
    normalizedWeight := runningWeight / totalWeight
    weightDistribution = append(weightDistribution, normalizedWeight)
  }

  fmt.Printf("weightDistribution %s\n", weightDistribution)
  return weightDistribution
}

func parseCsvRecord(csvRecord []string) (float64, float64, float64, error) {
    y0, err := strconv.ParseFloat(csvRecord[1], 64)
    y1, err := strconv.ParseFloat(csvRecord[2], 64)
    y2, err := strconv.ParseFloat(csvRecord[3], 64)
    return y0, y1, y2, err
}

// Initialize our SimulationSummaries slice
// There's gotta be a better way to do this
func initSimulationSummaries(numSimulations int, numGroups int) SimulationSummaries {
  var simulations SimulationSummaries
  simulations = make(SimulationSummaries, numSimulations)
  for s := 0; s < numSimulations; s++ {
    simulations[s] = make([]*Summary, numGroups)
    for g := 0; g < numGroups; g++ {
      simulations[s][g] = &Summary {}
    }
  }

  return simulations
}

func prepSimulationSummaries(simulations SimulationSummaries, numSimulations int, numGroups int) [][]string {
  flattenedSummaries := [][]string{}
  for s := 0; s < numSimulations; s++ {
    for g := 0; g < numGroups; g++ {
      summary := simulations[s][g]
      csvRecord := []string{strconv.Itoa(s), strconv.Itoa(g), strconv.FormatFloat(summary.y0, 'f', -1, 64), strconv.FormatFloat(summary.y1, 'f', -1, 64), strconv.FormatFloat(summary.y2, 'f', -1, 64)}
      flattenedSummaries = append(flattenedSummaries, csvRecord)
    }
  }
  return flattenedSummaries
}

func main() {
  simulations := flag.Int("simulations", 10000, "Number of simulations to run.")
  flag.Var(&weights, "weights", "How we should weight each group")
  flag.Parse()

  // This should be fleshed out a bit with: http://crypto.stanford.edu/~blynn/c2go/ch02.html
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

  results := initSimulationSummaries(*simulations, len(weights))

  weightDistribution := calculateWeightDistribution(weights)

  for { // every line in the csv reader
    csvRecord, err := reader.Read()
    if err == io.EOF {
      break
    }

    y0, y1, y2, err := parseCsvRecord(csvRecord)
    if err != nil {
      fmt.Printf("%v\n", err)
      break
    }

    C.dsfmt_fill_array_close_open(&dsfmt, r, C.int(*simulations));

    for j := 0; j < *simulations; j++ {
      ptr := unsafe.Pointer( uintptr(randoms) + uintptr(size * j) )
      current_sim = *(*float64)(ptr)

      assignment := getAssignment(weightDistribution, current_sim)

      results[j][assignment].y0 += y0
      results[j][assignment].y1 += y1
      results[j][assignment].y2 += y2
    }
  }
  // TODO: out.Write(fmt.Sprintf("%d, %s", *simulations, line))
  flattenedSummaries := prepSimulationSummaries(results, *simulations, len(weights))
  out.WriteAll(flattenedSummaries)
}


