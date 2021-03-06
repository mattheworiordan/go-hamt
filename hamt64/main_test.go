package hamt64_test

import (
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"testing"
	"time"
	"unsafe"

	"github.com/lleo/go-hamt/hamt64"
	"github.com/lleo/hamt/hamt64/castable"
	"github.com/lleo/stringutil"
	"github.com/pkg/errors"
)

type StrVal struct {
	Str string
	Val interface{}
}

var Mega = 1024 * 1024
var InitHamtNumKvsForPut = 1 * Mega
var TwoMega = 2 * Mega
var numKvs = InitHamtNumKvsForPut + TwoMega // 3 * Mega
var SVS []StrVal
var KVS64 []hamt64.KeyVal

var Functional bool
var TableOption int

var Hamt64 hamt64.Hamt

var Inc = stringutil.Lower.Inc

var StartTime = make(map[string]time.Time)
var RunTime = make(map[string]time.Duration)

func TestMain(m *testing.M) {
	var fixedonly, sparseonly, hybrid, all bool
	flag.BoolVar(&fixedonly, "F", false,
		"Use fixed tables only and exclude S and H Options.")
	flag.BoolVar(&sparseonly, "S", false,
		"Use sparse tables only and exclude F and H Options.")
	flag.BoolVar(&hybrid, "H", false,
		"Use sparse tables initially and exclude F and S Options.")
	flag.BoolVar(&all, "A", false,
		"Run all Tests w/ Options set to FixedTables, SparseTables, and HybridTables")

	var functional, transient, both bool
	flag.BoolVar(&functional, "f", false,
		"Run Tests against HamtFunctional struct; excludes transient option")
	flag.BoolVar(&transient, "t", false,
		"Run Tests against HamtFunctional struct; excludes functional option")
	flag.BoolVar(&both, "b", false,
		"Run Tests against both transient and functional Hamt types.")

	var appendLog bool
	flag.BoolVar(&appendLog, "a", false,
		"Append to log file rather than re-creating it.")

	var logFn string
	flag.StringVar(&logFn, "l", "test.log",
		"set the log file name.")

	flag.Parse()

	// If all flag set, ignore fixedonly, sparseonly, and hybrid.
	if !all {

		// only one flag may be set between fixedonly, sparseonly, and hybrid
		if (fixedonly && (sparseonly || hybrid)) ||
			(sparseonly && (fixedonly || hybrid)) ||
			(hybrid && (sparseonly || fixedonly)) {
			flag.PrintDefaults()
			os.Exit(1)
		}
	}

	// If no flags given, run all tests.
	if !(all || fixedonly || sparseonly || hybrid) {
		all = true
	}

	if !both {
		if functional && transient {
			flag.PrintDefaults()
			os.Exit(1)
		}
	}

	if !(both || functional || transient) {
		both = true
	}

	log.SetFlags(log.Lshortfile)

	var logFile *os.File
	var err error
	if appendLog {
		logFile, err = os.OpenFile(logFn, os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Fatal(errors.Wrapf(err,
				"failed to os.OpenFile(%q, os.O_CREATE|os.O_APPEND, 0666)",
				logFn))
		}
	} else {
		logFile, err = os.Create(logFn)
		if err != nil {
			log.Fatal(errors.Wrapf(err, "failed to os.Create(%q)", logFn))
		}
	}
	defer logFile.Close()

	log.SetOutput(logFile)

	log.Println("TestMain: and so it begins...")

	SVS = buildStrVals("TestMain", numKvs)
	KVS64 = svs2kvs64("TestMain", SVS, castable.CastStringKey)

	log.Printf("TestMain: NumIndexBits=%d\n", hamt64.NumIndexBits)
	fmt.Printf("TestMain: NumIndexBits=%d\n", hamt64.NumIndexBits)
	log.Printf("TestMain: IndexLimit=%d\n", hamt64.IndexLimit)
	//fmt.Printf("TestMain: IndexLimit=%d\n", hamt64.IndexLimit)
	log.Printf("TestMain: DepthLimit=%d\n", hamt64.DepthLimit)
	//fmt.Printf("TestMain: DepthLimit=%d\n", hamt64.DepthLimit)

	log.Printf("TestMain: SizeofHamtTransient=%d\n",
		unsafe.Sizeof(hamt64.HamtTransient{}))
	//fmt.Printf("TestMain: SizeofHamtTransient=%d\n",
	//	unsafe.Sizeof(hamt64.HamtTransient{}))
	log.Printf("TestMain: SizeofHamtFunctional=%d\n",
		unsafe.Sizeof(hamt64.HamtFunctional{}))
	//fmt.Printf("TestMain: SizeofHamtFunctional=%d\n",
	//	unsafe.Sizeof(hamt64.HamtFunctional{}))
	log.Printf("TestMain: SizeofHamtBase=%d\n", hamt64.SizeofHamtBase)
	//fmt.Printf("TestMain: SizeofHamtBase=%d\n", hamt64.SizeofHamtBase)
	log.Printf("TestMain: SizeofFixedTable=%d\n", hamt64.SizeofFixedTable)
	//fmt.Printf("TestMain: SizeofFixedTable=%d\n", hamt64.SizeofFixedTable)
	log.Printf("TestMain: SizeofSparseTable=%d\n", hamt64.SizeofSparseTable)
	//fmt.Printf("TestMain: SizeofSparseTable=%d\n", hamt64.SizeofSparseTable)

	// // This is an attempt to make the first benchmarks faster. My theory is
	// // that we needed to build up the heap. This worked a little bit, I don't
	// // know if it is really worth it or should I do more.
	// StartTime["fat throw away"] = time.Now()
	// foo, _ := buildHamt64("foo", KVS, true, hamt64.FixedTables)
	// _, found := foo.Get([]byte("aaa"))
	// if !found {
	// 	panic("foo failed to find \"aaa\"")
	// }
	// RunTime["fat throw away"] = time.Since(StartTime["fat throw away"])

	// execute
	var xit int
	if all {
		if both {
			Functional = false
			log.Printf("TestMain: Functional=%t;\n", Functional)
			fmt.Printf("TestMain: Functional=%t;\n", Functional)

			xit = executeAll(m)
			if xit != 0 {
				log.Printf("\n", RunTimes())
				os.Exit(xit)
			}

			Hamt64 = nil

			Functional = true
			log.Printf("TestMain: Functional=%t;\n", Functional)
			fmt.Printf("TestMain: Functional=%t;\n", Functional)

			xit = executeAll(m)
		} else if functional {
			Functional = true
			log.Printf("TestMain: Functional=%t;\n", Functional)
			fmt.Printf("TestMain: Functional=%t;\n", Functional)

			xit = executeAll(m)
		} else if transient {
			Functional = false
			log.Printf("TestMain: Functional=%t;\n", Functional)
			fmt.Printf("TestMain: Functional=%t;\n", Functional)

			xit = executeAll(m)
		}
	} else {
		if hybrid {
			TableOption = hamt64.HybridTables
		} else if fixedonly {
			TableOption = hamt64.FixedTables
		} else /* if sparseonly */ {
			TableOption = hamt64.SparseTables
		}

		if both {
			Functional = false

			log.Printf("TestMain: Functional=%t;\n", Functional)
			fmt.Printf("TestMain: Functional=%t;\n", Functional)
			log.Printf("TestMain: TableOption=%s;\n",
				hamt64.TableOptionName[TableOption])
			fmt.Printf("TestMain: TableOption=%s;\n",
				hamt64.TableOptionName[TableOption])

			xit = m.Run()
			if xit != 0 {
				log.Printf("\n", RunTimes())
				os.Exit(xit)
			}

			Hamt64 = nil
			Functional = true

			log.Printf("TestMain: Functional=%t;\n", Functional)
			fmt.Printf("TestMain: Functional=%t;\n", Functional)
			log.Printf("TestMain: TableOption=%s;\n",
				hamt64.TableOptionName[TableOption])
			fmt.Printf("TestMain: TableOption=%s;\n",
				hamt64.TableOptionName[TableOption])

			xit = m.Run()
		} else {
			if functional {
				Functional = true
			} else /* if transient */ {
				Functional = false
			}

			log.Printf("TestMain: Functional=%t;\n", Functional)
			fmt.Printf("TestMain: Functional=%t;\n", Functional)
			log.Printf("TestMain: TableOption=%s;\n",
				hamt64.TableOptionName[TableOption])
			fmt.Printf("TestMain: TableOption=%s;\n",
				hamt64.TableOptionName[TableOption])
			xit = m.Run()
		}
	}

	log.Println("\n", RunTimes())
	os.Exit(xit)
}

func executeAll(m *testing.M) int {
	TableOption = hamt64.SparseTables

	log.Printf("TestMain: TableOption=%s;\n",
		hamt64.TableOptionName[TableOption])
	fmt.Printf("TestMain: TableOption=%s;\n",
		hamt64.TableOptionName[TableOption])

	var xit = m.Run()
	if xit != 0 {
		log.Println("\n", RunTimes())
		os.Exit(1)
	}

	Hamt64 = nil
	TableOption = hamt64.FixedTables

	log.Printf("TestMain: TableOption=%s;\n",
		hamt64.TableOptionName[TableOption])
	fmt.Printf("TestMain: TableOption=%s;\n",
		hamt64.TableOptionName[TableOption])

	xit = m.Run()
	if xit != 0 {
		log.Println("\n", RunTimes())
		os.Exit(1)
	}

	Hamt64 = nil
	TableOption = hamt64.HybridTables

	log.Printf("TestMain: TableOption=%s;\n",
		hamt64.TableOptionName[TableOption])
	fmt.Printf("TestMain: TableOption=%s;\n",
		hamt64.TableOptionName[TableOption])

	xit = m.Run()

	return xit
}

func buildStrVals(prefix string, num int) []StrVal {
	var name = fmt.Sprintf("%s-buildStrVals-%d", prefix, num)
	StartTime[name] = time.Now()

	var kvs = make([]StrVal, num)
	var s = "aaa"

	for i := 0; i < num; i++ {
		kvs[i] = StrVal{s, i}
		s = Inc(s)
	}

	RunTime[name] = time.Since(StartTime[name])
	return kvs
}

func svs2kvs64(
	prefix string,
	svs []StrVal,
	fn func(string) hamt64.KeyI,
) []hamt64.KeyVal {
	var name = fmt.Sprintf("%s-svs2kvs64-%d", prefix, len(svs))
	StartTime[name] = time.Now()

	var kvs = make([]hamt64.KeyVal, len(svs))

	for i, sv := range svs {
		kvs[i] = hamt64.KeyVal{fn(sv.Str), sv.Val}
	}

	RunTime[name] = time.Since(StartTime[name])
	return kvs
}

func buildHamt64(
	prefix string,
	kvs []hamt64.KeyVal,
	functional bool,
	opt int,
) (hamt64.Hamt, error) {
	var name = fmt.Sprintf("%s-buildHamt64-%d", prefix, len(kvs))

	StartTime[name] = time.Now()
	var h = hamt64.New(functional, opt)
	for _, kv := range kvs {
		var k = kv.Key
		var v = kv.Val

		var inserted bool
		h, inserted = h.Put(k, v)
		if !inserted {
			return nil, fmt.Errorf("failed to Put(%q, %d)", k, v)
		}
	}
	RunTime[name] = time.Since(StartTime[name])

	return h, nil
}

func RunTimes() string {
	// Grab list of keys from RunTime map; MAJOR un-feature of Go!
	var ks = make([]string, len(RunTime))
	var i int = 0
	for k := range RunTime {
		ks[i] = k
		i++
	}
	sort.Strings(ks)

	var s = ""

	s += "Key                                                Val\n"
	s += "==================================================+==========\n"

	var tot time.Duration
	for _, k := range ks {
		v := RunTime[k]
		s += fmt.Sprintf("%-50s %s\n", k, v)
		tot += v
	}
	s += fmt.Sprintf("%50s %s\n", "TOTAL", tot)

	return s
}
