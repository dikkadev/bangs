package bangs

import (
	"fmt"
	"log/slog"
	"math"
	"os"
	"testing"
)

var sizes = []int{10, 20, 100, 500, 1e3, 1e4, 5e4, 1e5, 5e5, 1e6, 5e6, 1e7, 5e7, 1e8}

func generateRandomBangs(N int) BangList {
	bl := BangList{
		Entries: make(map[string]Entry, N),
		byBang:  make(map[string]Entry, N),
		len:     N,
	}

	alphabet := "abcdefghijklmnopqrstuvwxyz"
	totalGenerated := 0
	length := 1

	for totalGenerated < N {
		possibleCombos := int(math.Pow(26, float64(length)))
		totalGenerated += possibleCombos
		length++
	}

	length--

	generatedCount := 0
	for i := 0; i < length; i++ {
		for j := 0; j < 26; j++ {
			if generatedCount >= N {
				break
			}
			bang := string(alphabet[j])
			for k := 0; k < i; k++ {
				bang += string(alphabet[j])
			}
			bl.Entries[bang] = Entry{
				Bang: bang,
				URL:  QueryURL("https://www.google.com/search?q={}"),
			}
			bl.byBang[bang] = bl.Entries[bang]
			generatedCount++
		}
	}

	return bl

}

func BenchmarkPrepareInputPreComp(b *testing.B) {
	for _, size := range sizes {
		bl := generateRandomBangs(size)

		allBangs := make([]string, 0, len(bl.Entries))
		for _, bang := range bl.Entries {
			allBangs = append(allBangs, bang.Bang)
		}

		//all inputs
		inputs := make([]string, len(allBangs))
		for i, bang := range allBangs {
			inputs[i] = fmt.Sprintf("!%s some query text", bang)
		}

		b.ResetTimer()
		slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, nil)))

		b.Run(fmt.Sprintf("PrepareInput-%d", size), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				for _, input := range inputs {
					// _, _, err := bl.PrepareInputNaive(input)
					_, _, err := bl.PrepareInput(input)
					if err != nil {
						b.Errorf("PrepareInputPreComp failed: %v", err)
					}
				}
			}
		})
	}
}

func BenchmarkPrepareInputNaive(b *testing.B) {
	for _, size := range sizes {
		bl := generateRandomBangs(size)

		allBangs := make([]string, 0, len(bl.Entries))
		for _, bang := range bl.Entries {
			allBangs = append(allBangs, bang.Bang)
		}

		inputs := make([]string, len(allBangs))
		for i, bang := range allBangs {
			inputs[i] = fmt.Sprintf("!%s some query text", bang)
		}

		b.ResetTimer()
		slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, nil)))

		b.Run(fmt.Sprintf("PrepareInput-%d", size), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				for _, input := range inputs {
					_, _, err := bl.PrepareInputNaive(input)
					// _, _, err := bl.PrepareInput(input)
					if err != nil {
						b.Errorf("PrepareInputNaive failed: %v", err)
					}
				}
			}
		})
	}
}
