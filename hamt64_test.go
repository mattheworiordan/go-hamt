package hamt_test

import (
	"bytes"
	"log"
	"testing"
	"time"

	"github.com/lleo/go-hamt/hamt64"
)

func TestBuild64(t *testing.T) {
	var name = "TestBuild64"
	if Functional {
		name += ":functional:" + hamt64.TableOptionName[TableOption]
	} else {
		name += ":transient:" + hamt64.TableOptionName[TableOption]
	}

	var h = hamt64.New(Functional, TableOption)

	for _, sv := range BVS[:30] {
		var bs = sv.Bsl
		var v = sv.Val

		var inserted bool
		h, inserted = h.Put(bs, v)
		if !inserted {
			log.Printf("%s: failed to insert s=%q, v=%d", name, string(bs), v)
			t.Fatalf("%s: failed to insert s=%q, v=%d", name, string(bs), v)
		}

		//log.Print(h.LongString(""))
	}
}

func TestHamt64Put(t *testing.T) {
	var name = "TestHamt64Put"
	if Functional {
		name += ":functional:" + hamt64.TableOptionName[TableOption]
	} else {
		name += ":transient:" + hamt64.TableOptionName[TableOption]
	}

	StartTime[name] = time.Now()
	Hamt64 = hamt64.New(Functional, TableOption)
	for _, sv := range BVS {
		var bs = sv.Bsl
		var v = sv.Val

		var inserted bool
		Hamt64, inserted = Hamt64.Put(bs, v)
		if !inserted {
			log.Printf("%s: failed to Hamt64.Put(%q, %v)", name, string(bs), v)
			t.Fatalf("%s: failed to Hamt64.Put(%q, %v)", name, string(bs), v)
		}

		var val, found = Hamt64.Get(bs)
		if !found {
			log.Printf("%s: failed to Hamt64.Get(%q)", name, string(bs))
			//log.Print(Hamt64.LongString(""))
			t.Fatalf("%s: failed to Hamt64.Get(%q)", name, string(bs))
		}
		if val != v {
			log.Printf("%s: returned val,%d != expected v,%d for s=%q", name, val, v, string(bs))
			t.Fatalf("%s: returned val,%d != expected v,%d for s=%q", name, val, v, string(bs))
		}
	}
	RunTime[name] = time.Since(StartTime[name])

	StartTime["Hamt64.Stats()"] = time.Now()
	var stats = Hamt64.Stats()
	RunTime["Hamt64.Stats()"] = time.Since(StartTime["Hamt64.Stats()"])
	log.Printf("%s: stats=%+v;\n", name, stats)
}

func TestHamt64IterFunc(t *testing.T) {
	var name = "TestHamt64IterFunc"
	if Functional {
		name += ":functional:" + hamt64.TableOptionName[TableOption]
	} else {
		name += ":transient:" + hamt64.TableOptionName[TableOption]
	}

	if Hamt64 == nil {
		var err error
		Hamt64, err = buildHamt64(name, BVS, Functional, TableOption)
		if err != nil {
			log.Printf("%s: failed buildHamt64(%q, BVS#%d, %t, %s) => %s", name,
				name, len(BVS), Functional,
				hamt64.TableOptionName[TableOption], err)
			t.Fatalf("%s: failed buildHamt64(%q, BVS#%d, %t, %s) => %s", name,
				name, len(BVS), Functional,
				hamt64.TableOptionName[TableOption], err)
		}

		StartTime["Hamt64.Stats()"] = time.Now()
		var stats = Hamt64.Stats()
		RunTime["Hamt64.Stats()"] = time.Since(StartTime["Hamt64.Stats()"])
		log.Printf("%s: stats=%+v;\n", name, stats)
	}

	StartTime[name] = time.Now()

	var i int
	var next = Hamt64.Iter()
	for kv, ok := next(); ok; kv, ok = next() {
		var val, ok = Hamt64.Get(kv.Key)
		if !ok {
			t.Fatalf("failed to lookup %s in Hamt64", kv.Key)
		}

		if val != kv.Val {
			t.Fatalf("val,%v != kv.Val,%v\n", val, kv.Val)
		}

		i++
	}

	if len(BVS) != i {
		t.Fatalf("Expected len(BVS),%d go i,%d; Hamt64.Nentries()=%d;",
			len(BVS), i, Hamt64.Nentries())
	}

	RunTime[name] = time.Since(StartTime[name])
}

func TestHamt64IterChan(t *testing.T) {
	var name = "TestHamt64IterChan"
	if Functional {
		name += ":functional:" + hamt64.TableOptionName[TableOption]
	} else {
		name += ":transient:" + hamt64.TableOptionName[TableOption]
	}

	if Hamt64 == nil {
		var err error
		Hamt64, err = buildHamt64(name, BVS, Functional, TableOption)
		if err != nil {
			log.Printf("%s: failed buildHamt64(%q, BVS#%d, %t, %s) => %s", name,
				name, len(BVS), Functional,
				hamt64.TableOptionName[TableOption], err)
			t.Fatalf("%s: failed buildHamt64(%q, BVS#%d, %t, %s) => %s", name,
				name, len(BVS), Functional,
				hamt64.TableOptionName[TableOption], err)
		}

		StartTime["Hamt64.Stats()"] = time.Now()
		var stats = Hamt64.Stats()
		RunTime["Hamt64.Stats()"] = time.Since(StartTime["Hamt64.Stats()"])
		log.Printf("%s: stats=%+v;\n", name, stats)
	}

	StartTime[name] = time.Now()

	var i int
	for kv := range Hamt64.IterChan(0) {
		var val, ok = Hamt64.Get(kv.Key)
		if !ok {
			t.Fatalf("failed to lookup %s in Hamt64", kv.Key)
		}

		if val != kv.Val {
			t.Fatalf("val,%v != kv.Val,%v\n", val, kv.Val)
		}

		i++
	}

	if len(BVS) != i {
		t.Fatalf("Expected len(BVS),%d go i,%d", len(BVS), i)
	}

	RunTime[name] = time.Since(StartTime[name])
}

func TestHamt64IterChanCancel(t *testing.T) {
	var name = "TestHamt64IterChanCancel"
	if Functional {
		name += ":functional:" + hamt64.TableOptionName[TableOption]
	} else {
		name += ":transient:" + hamt64.TableOptionName[TableOption]
	}

	if Hamt64 == nil {
		var err error
		Hamt64, err = buildHamt64(name, BVS, Functional, TableOption)
		if err != nil {
			log.Printf("%s: failed buildHamt64(%q, BVS#%d, %t, %s) => %s", name,
				name, len(BVS), Functional,
				hamt64.TableOptionName[TableOption], err)
			t.Fatalf("%s: failed buildHamt64(%q, BVS#%d, %t, %s) => %s", name,
				name, len(BVS), Functional,
				hamt64.TableOptionName[TableOption], err)
		}

		StartTime["Hamt64.Stats()"] = time.Now()
		var stats = Hamt64.Stats()
		RunTime["Hamt64.Stats()"] = time.Since(StartTime["Hamt64.Stats()"])
		log.Printf("%s: stats=%+v;\n", name, stats)
	}

	StartTime[name] = time.Now()

	var i int
	var stopKey = BVS[0].Bsl // "aaa" but key from iter are random
	var iterChan, iterChanCancel = Hamt64.IterChanWithCancel(0)
	for kv := range iterChan {
		var val, ok = Hamt64.Get(kv.Key)
		if !ok {
			t.Fatalf("failed to lookup %s in Hamt64", kv.Key)
		}

		if val != kv.Val {
			t.Fatalf("val,%v != kv.Val,%v\n", val, kv.Val)
		}

		i++

		if bytes.Equal(kv.Key, stopKey) {
			iterChanCancel()
			break
		}
	}

	log.Printf("%s: stopped after %d iterations", name, i)

	RunTime[name] = time.Since(StartTime[name])
}

func TestHamt64Get(t *testing.T) {
	var name = "TestHamt64Get"
	if Functional {
		name += ":functional:" + hamt64.TableOptionName[TableOption]
	} else {
		name += ":transient:" + hamt64.TableOptionName[TableOption]
	}

	if Hamt64 == nil {
		var err error
		Hamt64, err = buildHamt64(name, BVS, Functional, TableOption)
		if err != nil {
			log.Printf("%s: failed buildHamt64(%q, BVS#%d, %t, %s) => %s", name,
				name, len(BVS), Functional,
				hamt64.TableOptionName[TableOption], err)
			t.Fatalf("%s: failed buildHamt64(%q, BVS#%d, %t, %s) => %s", name,
				name, len(BVS), Functional,
				hamt64.TableOptionName[TableOption], err)
		}

		StartTime["Hamt64.Stats()"] = time.Now()
		var stats = Hamt64.Stats()
		RunTime["Hamt64.Stats()"] = time.Since(StartTime["Hamt64.Stats()"])
		log.Printf("%s: stats=%+v;\n", name, stats)
	}

	StartTime[name] = time.Now()
	for _, sv := range BVS {
		var bs = sv.Bsl
		var v = sv.Val

		var val, found = Hamt64.Get(bs)
		if !found {
			log.Printf("%s: Failed to Hamt64.Get(%q)", name, string(bs))
			log.Print(Hamt64.LongString(""))
			t.Fatalf("%s: Failed to Hamt64.Get(%q)", name, string(bs))
		}
		if val != v {
			log.Printf("%s: retrieved val,%d != expected v,%d for s=%q", name, val, v, string(bs))
			t.Fatalf("%s: retrieved val,%d != expected v,%d for s=%q", name, val, v, string(bs))
		}
	}
	RunTime[name] = time.Since(StartTime[name])
}

func TestHamt64Del(t *testing.T) {
	var name = "TestHamt64Del"
	if Functional {
		name += ":functional:" + hamt64.TableOptionName[TableOption]
	} else {
		name += ":transient:" + hamt64.TableOptionName[TableOption]
	}

	if Hamt64 == nil {
		var err error
		Hamt64, err = buildHamt64(name, BVS, Functional, TableOption)
		if err != nil {
			log.Printf("%s: failed buildHamt64(%q, BVS#%d, %t, %s) => %s", name,
				name, len(BVS), Functional,
				hamt64.TableOptionName[TableOption], err)
			t.Fatalf("%s: failed buildHamt64(%q, BVS#%d, %t, %s) => %s", name,
				name, len(BVS), Functional,
				hamt64.TableOptionName[TableOption], err)
		}

		StartTime["Hamt64.Stats()"] = time.Now()
		var stats = Hamt64.Stats()
		RunTime["Hamt64.Stats()"] = time.Since(StartTime["Hamt64.Stats()"])
		log.Printf("%s: stats=%+v;\n", name, stats)
	}

	StartTime[name] = time.Now()
	for _, sv := range BVS {
		var bs = sv.Bsl
		var v = sv.Val

		var val interface{}
		var deleted bool
		Hamt64, val, deleted = Hamt64.Del(bs)
		if !deleted {
			log.Printf("%s: Failed to Hamt64.Del(%q)", name, string(bs))
			log.Print(Hamt64.LongString(""))
			t.Fatalf("%s: Failed to Hamt64.Del(%q)", name, string(bs))
		}
		if val != v {
			log.Printf("%s: retrieved val,%d != expected v,%d for s=%q", name, val, v, string(bs))
			t.Fatalf("%s: retrieved val,%d != expected v,%d for s=%q", name, val, v, string(bs))
		}
	}
	RunTime[name] = time.Since(StartTime[name])
}

var BenchHamt64Get hamt64.Hamt
var BenchHamt64Get_Functional bool

func BenchmarkHamt64Get(b *testing.B) {
	var name = "BenchmarkHamt64Get"
	if Functional {
		name += ":functional:" + hamt64.TableOptionName[TableOption]
	} else {
		name += ":transient:" + hamt64.TableOptionName[TableOption]
	}

	if BenchHamt64Get == nil || BenchHamt64Get_Functional != Functional {
		BenchHamt64Get_Functional = Functional

		var err error
		BenchHamt64Get, err = buildHamt64(name, BVS, Functional, TableOption)
		if err != nil {
			log.Printf("%s: failed buildHamt64(%q, BVS#%d, %t, %s) => %s", name,
				name, len(BVS), false, hamt64.TableOptionName[TableOption], err)
			b.Fatalf("%s: failed buildHamt64(%q, BVS#%d, %t, %s) => %s", name,
				name, len(BVS), false, hamt64.TableOptionName[TableOption], err)
		}
	}

	log.Printf("%s: b.N=%d", name, b.N)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var j = i % len(BVS)
		var bs = BVS[j].Bsl
		var v = BVS[j].Val

		var val, found = BenchHamt64Get.Get(bs)
		if !found {
			log.Printf("%s: Failed to h.Get(%q)", name, string(bs))
			//log.Print(h.LongString(""))
			b.Fatalf("%s: Failed to h.Get(%q)", name, string(bs))
		}
		if val != v {
			log.Printf("%s: retrieved val,%d != expected v,%d for s=%q", name, val, v, string(bs))
			b.Fatalf("%s: retrieved val,%d != expected v,%d for s=%q", name, val, v, string(bs))
		}
	}
}

var BenchHamt64_T2F hamt64.Hamt

func BenchmarkHamt64_T2F_Get(b *testing.B) {
	var name = "BenchmarkHamt64_T2F_Get"
	name += ":functional:" + hamt64.TableOptionName[TableOption]

	if BenchHamt64_T2F == nil {
		var err error
		BenchHamt64_T2F, err = buildHamt64(name, BVS, false, TableOption)
		if err != nil {
			log.Printf("%s: failed buildHamt64(%q, BVS#%d, %t, %s) => %s", name,
				name, len(BVS), false, hamt64.TableOptionName[TableOption], err)
			b.Fatalf("%s: failed buildHamt64(%q, BVS#%d, %t, %s) => %s", name,
				name, len(BVS), false, hamt64.TableOptionName[TableOption], err)
		}
		BenchHamt64_T2F = BenchHamt64_T2F.ToFunctional()
	}

	log.Printf("%s: Transient-to-Functional; b.N=%d;", name, b.N)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var j = i % len(BVS)
		var bs = BVS[j].Bsl
		var v = BVS[j].Val

		var val, found = BenchHamt64_T2F.Get(bs)
		if !found {
			log.Printf("%s: Failed to BenchHamt64_T2F.Get(%q)", name, string(bs))
			//log.Print(h.LongString(""))
			b.Fatalf("%s: Failed to BenchHamt64_T2F.Get(%q)", name, string(bs))
		}
		if val != v {
			log.Printf("%s: retrieved val,%d != expected v,%d for s=%q", name, val, v, string(bs))
			b.Fatalf("%s: retrieved val,%d != expected v,%d for s=%q", name, val, v, string(bs))
		}
	}
}

var BenchHamt64_F2T hamt64.Hamt

func BenchmarkHamt64_F2T_Get(b *testing.B) {
	var name = "BenchmarkHamt64_F2T_Get"
	name += ":transient:" + hamt64.TableOptionName[TableOption]

	if BenchHamt64_F2T == nil {
		var err error
		BenchHamt64_F2T, err = buildHamt64(name, BVS, true, TableOption)
		if err != nil {
			log.Printf("%s: failed buildHamt64(%q, BVS#%d, %t, %s) => %s", name,
				name, len(BVS), false, hamt64.TableOptionName[TableOption], err)
			b.Fatalf("%s: failed buildHamt64(%q, BVS#%d, %t, %s) => %s", name,
				name, len(BVS), false, hamt64.TableOptionName[TableOption], err)
		}
		BenchHamt64_F2T = BenchHamt64_F2T.ToTransient()
	}

	log.Printf("%s: Functional-to-Transient; b.N=%d;", name, b.N)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var j = i % len(BVS)
		var bs = BVS[j].Bsl
		var v = BVS[j].Val

		var val, found = BenchHamt64_F2T.Get(bs)
		if !found {
			log.Printf("%s: Failed to BenchHamt64_F2T.Get(%q)", name, string(bs))
			//log.Print(h.LongString(""))
			b.Fatalf("%s: Failed to BenchHamt64_F2T.Get(%q)", name, string(bs))
		}
		if val != v {
			log.Printf("%s: retrieved val,%d != expected v,%d for s=%q", name, val, v, string(bs))
			b.Fatalf("%s: retrieved val,%d != expected v,%d for s=%q", name, val, v, string(bs))
		}
	}
}

func BenchmarkHamt64Put(b *testing.B) {
	var name = "BenchmarkHamt64Put"
	if Functional {
		name += ":functional:" + hamt64.TableOptionName[TableOption]
	} else {
		name += ":transient:" + hamt64.TableOptionName[TableOption]
	}

	if b.N+InitHamtNumBvsForPut > len(BVS) {
		log.Printf("%s: Can't run: b.N+num > len(BVS)", name)
		b.Fatalf("%s: Can't run: b.N+num > len(BVS)", name)
	}

	var bvs = BVS[:InitHamtNumBvsForPut]

	var h, err = buildHamt64(name, bvs, Functional, TableOption)
	if err != nil {
		log.Printf("%s: failed buildHamt64(%q, BVS[:%d], %t, %s) => %s", name,
			name, InitHamtNumBvsForPut, Functional,
			hamt64.TableOptionName[TableOption], err)
		b.Fatalf("%s: failed buildHamt64(%q, BVS[:%d], %t, %s) => %s", name,
			name, InitHamtNumBvsForPut, Functional,
			hamt64.TableOptionName[TableOption], err)
	}

	log.Printf("%s: b.N=%d;", name, b.N)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var bs = BVS[InitHamtNumBvsForPut+i].Bsl
		var v = BVS[InitHamtNumBvsForPut+i].Val

		var added bool
		h, added = h.Put(bs, v)
		if !added {
			log.Printf("%s: failed to h.Put(%q, %d)", name, string(bs), v)
			//log.Print(h.LongString(""))
			b.Fatalf("%s: failed to h.Put(%q, %d)", name, string(bs), v)
		}
	}
}

func BenchmarkHamt64_T2F_Put(b *testing.B) {
	var name = "BenchmarkHamt64Put_T2F"
	name += ":functional:" + hamt64.TableOptionName[TableOption]

	var InitHamtNumBvsForPut int //= 1000000 // 1 million; allows b.N=3,000,000
	if b.N+InitHamtNumBvsForPut > len(BVS) {
		log.Printf("%s: Can't run: b.N+num > len(BVS)", name)
		b.Fatalf("%s: Can't run: b.N+num > len(BVS)", name)
	}

	var bvs = BVS[:InitHamtNumBvsForPut]

	var h, err = buildHamt64(name, bvs, false, TableOption)
	if err != nil {
		log.Printf("%s: failed buildHamt64(%q, BVS[:%d], %t, %s) => %s", name,
			name, InitHamtNumBvsForPut, Functional,
			hamt64.TableOptionName[TableOption], err)
		b.Fatalf("%s: failed buildHamt64(%q, BVS[:%d], %t, %s) => %s", name,
			name, InitHamtNumBvsForPut, Functional,
			hamt64.TableOptionName[TableOption], err)
	}
	h = h.ToFunctional()

	log.Printf("%s: Transient-to-Functional; b.N=%d;", name, b.N)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var bs = BVS[InitHamtNumBvsForPut+i].Bsl
		var v = BVS[InitHamtNumBvsForPut+i].Val

		var added bool
		h, added = h.Put(bs, v)
		if !added {
			log.Printf("%s: failed to h.Put(%q, %d)", name, string(bs), v)
			//log.Print(h.LongString(""))
			b.Fatalf("%s: failed to h.Put(%q, %d)", name, string(bs), v)
		}
	}
}

func BenchmarkHamt64Del(b *testing.B) {
	var name = "BenchmarkHamt64Del"
	if Functional {
		name += ":functional:" + hamt64.TableOptionName[TableOption]
	} else {
		name += ":transient:" + hamt64.TableOptionName[TableOption]
	}

	var h, err = buildHamt64(name, BVS[:TwoKK], Functional, TableOption)
	if err != nil {
		log.Printf("%s: failed buildHamt64(%q, BVS:%d, %t, %s) => %s", name,
			name, len(BVS), Functional,
			hamt64.TableOptionName[TableOption], err)
		b.Fatalf("%s: failed buildHamt64(%q, BVS:%d, %t, %s) => %s", name,
			name, len(BVS), Functional,
			hamt64.TableOptionName[TableOption], err)
	}

	log.Printf("%s: b.N=%d;", name, b.N)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var bs = BVS[i].Bsl
		var v = BVS[i].Val

		var deleted bool
		var val interface{}
		h, val, deleted = h.Del(bs)
		if !deleted {
			log.Printf("%s: failed to h.Del(%q)", name, string(bs))
			//log.Print(h.LongString(""))
			b.Fatalf("%s: failed to h.Del(%q)", name, string(bs))
		}
		if val != v {
			log.Printf("%s: failed val,%d != v,%d", name, val, v)
			b.Fatalf("%s: failed val,%d != v,%d", name, val, v)
		}
	}
}

func BenchmarkHamt64IterFunc(b *testing.B) {
	var name = "BenchmarkHamt64Iter"
	if Functional {
		name += ":functional:" + hamt64.TableOptionName[TableOption]
	} else {
		name += ":transient:" + hamt64.TableOptionName[TableOption]
	}

	// We will reuse the BenchHamtGet tree for iterator benchmarking.
	if BenchHamt64Get == nil || BenchHamt64Get_Functional != Functional {
		BenchHamt64Get_Functional = Functional

		var err error
		BenchHamt64Get, err = buildHamt64(name, BVS, Functional, TableOption)
		if err != nil {
			log.Printf("%s: failed buildHamt64(%q, BVS#%d, %t, %s) => %s", name,
				name, len(BVS), false, hamt64.TableOptionName[TableOption], err)
			b.Fatalf("%s: failed buildHamt64(%q, BVS#%d, %t, %s) => %s", name,
				name, len(BVS), false, hamt64.TableOptionName[TableOption], err)
		}
	}

	log.Printf("%s: b.N=%d", name, b.N)
	b.ResetTimer()

	var i int
	var next = BenchHamt64Get.Iter()
	for kv, ok := next(); ok; kv, ok = next() {
		if len(kv.Key) < 0 {
			b.Fatal("stupid test to touch the kv")
		}

		if i >= b.N {
			break
		}
		i++
	}

	if i != b.N {
		b.Fatalf("Failed to run b.N,%d iterations; only ran %d.", b.N, i)
	}
}

func BenchmarkHamt64IterChan(b *testing.B) {
	var name = "BenchmarkHamt64Iter"
	if Functional {
		name += ":functional:" + hamt64.TableOptionName[TableOption]
	} else {
		name += ":transient:" + hamt64.TableOptionName[TableOption]
	}

	// We will reuse the BenchHamtGet tree for iterator benchmarking.
	if BenchHamt64Get == nil || BenchHamt64Get_Functional != Functional {
		BenchHamt64Get_Functional = Functional

		var err error
		BenchHamt64Get, err = buildHamt64(name, BVS, Functional, TableOption)
		if err != nil {
			log.Printf("%s: failed buildHamt64(%q, BVS#%d, %t, %s) => %s", name,
				name, len(BVS), false, hamt64.TableOptionName[TableOption], err)
			b.Fatalf("%s: failed buildHamt64(%q, BVS#%d, %t, %s) => %s", name,
				name, len(BVS), false, hamt64.TableOptionName[TableOption], err)
		}
	}

	log.Printf("%s: b.N=%d", name, b.N)
	b.ResetTimer()

	var i int
	var iterChan, iterChanCancel = BenchHamt64Get.IterChanWithCancel(0)
	for kv := range iterChan {
		if len(kv.Key) < 0 {
			b.Fatal("stupid test to touch the kv")
		}

		if i >= b.N {
			iterChanCancel()
			break
		}
		i++
	}

	if i != b.N {
		b.Fatalf("Failed to run b.N,%d iterations; only ran %d.", b.N, i)
	}
}
