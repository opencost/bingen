package stringutil_test

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"
	"time"
	"unsafe"

	"github.com/opencost/bingen/pkg/util/stringutil"
)

var alpha = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

type bankTest struct {
	Bank     func(string) string
	BankFunc func(string, func() string) string
	Clear    func()
}

var (
	standardBankTest = bankTest{
		Bank:     stringutil.Bank,
		BankFunc: stringutil.BankFunc,
		Clear:    stringutil.ClearBank,
	}
)

func RandSeqWith(r *rand.Rand, n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = alpha[r.Intn(len(alpha))] // #nosec No need for a cryptographic strength random here
	}
	return string(b)
}

func generateBenchData(totalStrings, totalUnique int) [][]byte {
	randStrings := make([]string, 0, totalStrings)
	r := rand.New(rand.NewSource(27644437))

	// create totalUnique unique strings
	for range totalUnique {
		randStrings = append(
			randStrings,
			fmt.Sprintf("%s/%s/%s", RandSeqWith(r, 10), RandSeqWith(r, 10), RandSeqWith(r, 10)),
		)
	}

	// set the seed such that the resulting "remainder" strings are deterministic for each bench
	r = rand.New(rand.NewSource(1523942))

	// append a random selection from 0-totalUnique to the list.
	for range totalStrings - totalUnique {
		randStrings = append(randStrings, strings.Clone(randStrings[r.Intn(totalUnique)]))
	}

	// shuffle the list of strings
	r.Shuffle(totalStrings, func(i, j int) { randStrings[i], randStrings[j] = randStrings[j], randStrings[i] })

	stringBytes := make([][]byte, 0, totalStrings)
	for _, str := range randStrings {
		stringBytes = append(stringBytes, []byte(str))
	}
	return stringBytes
}

func benchmarkStringBank(b *testing.B, bt bankTest, totalStrings, totalUnique int, useBankFunc bool) {
	b.StopTimer()
	randStrings := generateBenchData(totalStrings, totalUnique)

	b.StartTimer()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for bb := 0; bb < totalStrings; bb++ {
			bytes := randStrings[bb]
			str := unsafe.String(unsafe.SliceData(bytes), len(bytes))

			if useBankFunc {
				bt.BankFunc(str, func() string {
					return string(bytes)
				})
			} else {
				bt.Bank(str)
			}
		}
		b.StopTimer()
		bt.Clear()
		b.StartTimer()
	}
}

func BenchmarkStringBank90PercentDuplicate(b *testing.B) {
	prevBank := stringutil.GetStringBank()
	defer func() {
		stringutil.UpdateStringBank(prevBank)
	}()

	sb := stringutil.NewStringBank()
	stringutil.UpdateStringBank(sb)

	benchmarkStringBank(b, standardBankTest, 1_000_000, 100_000, false)
}

func BenchmarkStringBank75PercentDuplicate(b *testing.B) {
	prevBank := stringutil.GetStringBank()
	defer func() {
		stringutil.UpdateStringBank(prevBank)
	}()

	sb := stringutil.NewStringBank()
	stringutil.UpdateStringBank(sb)

	benchmarkStringBank(b, standardBankTest, 1_000_000, 250_000, false)
}

func BenchmarkStringBank50PercentDuplicate(b *testing.B) {
	prevBank := stringutil.GetStringBank()
	defer func() {
		stringutil.UpdateStringBank(prevBank)
	}()

	sb := stringutil.NewStringBank()
	stringutil.UpdateStringBank(sb)

	benchmarkStringBank(b, standardBankTest, 1_000_000, 500_000, false)
}

func BenchmarkStringBank25PercentDuplicate(b *testing.B) {
	prevBank := stringutil.GetStringBank()
	defer func() {
		stringutil.UpdateStringBank(prevBank)
	}()

	sb := stringutil.NewStringBank()
	stringutil.UpdateStringBank(sb)

	benchmarkStringBank(b, standardBankTest, 1_000_000, 750_000, false)
}

func BenchmarkStringBankNoDuplicate(b *testing.B) {
	prevBank := stringutil.GetStringBank()
	defer func() {
		stringutil.UpdateStringBank(prevBank)
	}()

	sb := stringutil.NewStringBank()
	stringutil.UpdateStringBank(sb)

	benchmarkStringBank(b, standardBankTest, 1_000_000, 1_000_000, false)
}

func BenchmarkStringBankFunc90PercentDuplicate(b *testing.B) {
	prevBank := stringutil.GetStringBank()
	defer func() {
		stringutil.UpdateStringBank(prevBank)
	}()

	sb := stringutil.NewStringBank()
	stringutil.UpdateStringBank(sb)

	benchmarkStringBank(b, standardBankTest, 1_000_000, 100_000, true)
}

func BenchmarkStringBankFunc75PercentDuplicate(b *testing.B) {
	prevBank := stringutil.GetStringBank()
	defer func() {
		stringutil.UpdateStringBank(prevBank)
	}()

	sb := stringutil.NewStringBank()
	stringutil.UpdateStringBank(sb)

	benchmarkStringBank(b, standardBankTest, 1_000_000, 250_000, true)
}

func BenchmarkStringBankFunc50PercentDuplicate(b *testing.B) {
	prevBank := stringutil.GetStringBank()
	defer func() {
		stringutil.UpdateStringBank(prevBank)
	}()

	sb := stringutil.NewStringBank()
	stringutil.UpdateStringBank(sb)

	benchmarkStringBank(b, standardBankTest, 1_000_000, 100_000, true)
}

func BenchmarkStringBankFunc25PercentDuplicate(b *testing.B) {
	prevBank := stringutil.GetStringBank()
	defer func() {
		stringutil.UpdateStringBank(prevBank)
	}()

	sb := stringutil.NewStringBank()
	stringutil.UpdateStringBank(sb)

	benchmarkStringBank(b, standardBankTest, 1_000_000, 750_000, true)
}

func BenchmarkStringBankFuncNoDuplicate(b *testing.B) {
	prevBank := stringutil.GetStringBank()
	defer func() {
		stringutil.UpdateStringBank(prevBank)
	}()

	sb := stringutil.NewStringBank()
	stringutil.UpdateStringBank(sb)

	benchmarkStringBank(b, standardBankTest, 1_000_000, 1_000_000, true)
}

func BenchmarkNoOpStringBankFunc90PercentDuplicate(b *testing.B) {
	prevBank := stringutil.GetStringBank()
	defer func() {
		stringutil.UpdateStringBank(prevBank)
	}()

	sb := stringutil.NewNoOpStringBank()
	stringutil.UpdateStringBank(sb)

	benchmarkStringBank(b, standardBankTest, 1_000_000, 100_000, true)
}

func BenchmarkNoOpStringBankFunc75PercentDuplicate(b *testing.B) {
	prevBank := stringutil.GetStringBank()
	defer func() {
		stringutil.UpdateStringBank(prevBank)
	}()

	sb := stringutil.NewNoOpStringBank()
	stringutil.UpdateStringBank(sb)

	benchmarkStringBank(b, standardBankTest, 1_000_000, 250_000, true)
}

func BenchmarkNoOpStringBankFunc50PercentDuplicate(b *testing.B) {
	prevBank := stringutil.GetStringBank()
	defer func() {
		stringutil.UpdateStringBank(prevBank)
	}()

	sb := stringutil.NewNoOpStringBank()
	stringutil.UpdateStringBank(sb)

	benchmarkStringBank(b, standardBankTest, 1_000_000, 100_000, true)
}

func BenchmarkNoOpStringBankFunc25PercentDuplicate(b *testing.B) {
	prevBank := stringutil.GetStringBank()
	defer func() {
		stringutil.UpdateStringBank(prevBank)
	}()

	sb := stringutil.NewNoOpStringBank()
	stringutil.UpdateStringBank(sb)

	benchmarkStringBank(b, standardBankTest, 1_000_000, 750_000, true)
}

func BenchmarkNoOpStringBankFuncNoDuplicate(b *testing.B) {
	prevBank := stringutil.GetStringBank()
	defer func() {
		stringutil.UpdateStringBank(prevBank)
	}()

	sb := stringutil.NewNoOpStringBank()
	stringutil.UpdateStringBank(sb)

	benchmarkStringBank(b, standardBankTest, 1_000_000, 1_000_000, true)
}

const LruCapacity = 500_000
const LruEvictInterval = 5 * time.Second

func BenchmarkLruStringBankFunc90PercentDuplicate(b *testing.B) {
	prevBank := stringutil.GetStringBank()
	defer func() {
		stringutil.UpdateStringBank(prevBank)
	}()

	sb := stringutil.NewLruStringBank(LruCapacity, LruEvictInterval)
	defer func() {
		if lruBank, ok := sb.(interface{ Stop() }); ok {
			lruBank.Stop()
		}

	}()

	stringutil.UpdateStringBank(sb)
	benchmarkStringBank(b, standardBankTest, 1_000_000, 100_000, true)
}

func BenchmarkLruStringBankFunc75PercentDuplicate(b *testing.B) {
	prevBank := stringutil.GetStringBank()
	defer func() {
		stringutil.UpdateStringBank(prevBank)
	}()

	sb := stringutil.NewLruStringBank(LruCapacity, LruEvictInterval)
	defer func() {
		if lruBank, ok := sb.(interface{ Stop() }); ok {
			lruBank.Stop()
		}
	}()

	stringutil.UpdateStringBank(sb)
	benchmarkStringBank(b, standardBankTest, 1_000_000, 250_000, true)
}

func BenchmarkLruStringBankFunc50PercentDuplicate(b *testing.B) {
	prevBank := stringutil.GetStringBank()
	defer func() {
		stringutil.UpdateStringBank(prevBank)
	}()

	sb := stringutil.NewLruStringBank(LruCapacity, LruEvictInterval)
	defer func() {
		if lruBank, ok := sb.(interface{ Stop() }); ok {
			lruBank.Stop()
		}
	}()

	stringutil.UpdateStringBank(sb)
	benchmarkStringBank(b, standardBankTest, 1_000_000, 100_000, true)
}

func BenchmarkLruStringBankFunc25PercentDuplicate(b *testing.B) {
	prevBank := stringutil.GetStringBank()
	defer func() {
		stringutil.UpdateStringBank(prevBank)
	}()

	sb := stringutil.NewLruStringBank(LruCapacity, LruEvictInterval)
	defer func() {
		if lruBank, ok := sb.(interface{ Stop() }); ok {
			lruBank.Stop()
		}
	}()

	stringutil.UpdateStringBank(sb)
	benchmarkStringBank(b, standardBankTest, 1_000_000, 750_000, true)
}

func BenchmarkLruStringBankFuncNoDuplicate(b *testing.B) {
	prevBank := stringutil.GetStringBank()
	defer func() {
		stringutil.UpdateStringBank(prevBank)
	}()

	sb := stringutil.NewLruStringBank(LruCapacity, LruEvictInterval)
	defer func() {
		if lruBank, ok := sb.(interface{ Stop() }); ok {
			lruBank.Stop()
		}
	}()

	stringutil.UpdateStringBank(sb)
	benchmarkStringBank(b, standardBankTest, 1_000_000, 1_000_000, true)
}
