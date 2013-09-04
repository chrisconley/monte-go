// Usage:

// From local:
// rsync -avt --delete --exclude=".git" . analytics:/tmp/monte

// On dev box
// go build simulate.go
// time head -n 100000 test/samples.csv | (./simulate --simulations=10000 --weights 5 --weights 5) > /mnt/tmp/montego-results.csv

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
  "log"
  "sync"
)

type RandomGenerator struct {
  dsfmt *C.dsfmt_t
}

type Summary struct {
  y0 float64
  y1 float64
  y2 float64
  mu sync.Mutex
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

func runCsvRecordSimulations(reader *csv.Reader, simulationSummaries SimulationSummaries, numSimulations int, dsfmt *C.dsfmt_t, weightDistribution []float64) error {
    // Read from csv until we hit the end of the file
    csvRecord, err := reader.Read()
    if err == io.EOF {
      return err
    }

    // If we encounter an error attempting to grab y0, y1, y2 from this csv record, exit.
    y0, y1, y2, err := parseCsvRecord(csvRecord)
    if err != nil {
      log.Fatalf("%v\n", err)
    }

    // Allocate aligned memory to hold the random numbers we'll generate for the number of simulations we want to run.
    size := int(unsafe.Sizeof(C.double(12)))
    randoms := C.memalign(16, C.size_t(size * numSimulations))
    //defer C.free(randoms) // TODO: If we make a function for this stuff in the loop, we can use defer

    // Generates double precision floating point
    // pseudorandom numbers which distribute in the range [0, 1) to the
    // array held at the `randoms` pointer.
    C.dsfmt_fill_array_close_open(dsfmt, (*C.double)(randoms), C.int(numSimulations));

    for i := 0; i < numSimulations; i++ {
      // Here we grab a pointer to the next random number and grab the value
      // at that location as a float.
      ptr := unsafe.Pointer( uintptr(randoms) + uintptr(size * i) )
      currentRandom := *(*float64)(ptr)

      // Get the group assignment and update the appropriate simulationSummary
      assignment := getAssignment(weightDistribution, currentRandom)
      summary := simulationSummaries[i][assignment]
      summary.y0 += y0
      summary.y1 += y1
      summary.y2 += y2
    }

    C.free(randoms)
    return nil
}

// TODO: Profiling: http://blog.golang.org/profiling-go-programs
func main() {
  var numSimulations int
  var weights WeightSet
  flag.IntVar(&numSimulations, "simulations", 10000, "Number of simulations to run.")
  flag.Var(&weights, "weights", "How we should weight each group")
  flag.Parse()

  numGroups := len(weights)
  simulationSummaries := initSimulationSummaries(numSimulations, numGroups)
  weightDistribution := calculateWeightDistribution(weights)

  reader := csv.NewReader(os.Stdin)
  out := csv.NewWriter(os.Stdout)

  // Initialize and seed the Double SIMD Fast Mersenne Twister.
  var dsfmt C.dsfmt_t
  C.dsfmt_init_gen_rand(&dsfmt, 1234);

  for {
    err := runCsvRecordSimulations(reader, simulationSummaries, numSimulations, &dsfmt, weightDistribution)
    if err == io.EOF {
      break
    }
  }

  // Write out the simulationSummaries to the csv writer.
  flattenedSummaries := prepSimulationSummaries(simulationSummaries, numSimulations, numGroups)
  out.WriteAll(flattenedSummaries)
}


