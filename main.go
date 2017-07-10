package main

import (
	"flag"
	"fmt"
	"sort"
	"strings"

	"github.com/domino14/macondo/anagrammer"
	"github.com/domino14/macondo/lexicon"
)

var dawgPath = flag.String("dawgpath", "", "path for dawgs")
var defaultDist lexicon.LetterDistribution

type Stem struct {
	StemCombinations        uint64
	ModifiedStemProbability float64
	UsableTiles             uint8
	MMPR                    float64
	Alphagram               string
}

func (s *Stem) Printable(order int) string {
	return fmt.Sprintf("%v\t%v\t%.3f\t%v\t%.3f (combos %v)", order, s.Alphagram,
		s.ModifiedStemProbability, s.UsableTiles, s.MMPR, s.StemCombinations)
}

var stemMap map[string]map[rune]bool

func main() {

	flag.Parse()
	anagrammer.LoadDawgs(*dawgPath)
	defaultDist = lexicon.EnglishLetterDistribution()
	stemMap = make(map[string]map[rune]bool)

	sevens := anagrammer.Anagram("???????", anagrammer.Dawgs["America"],
		anagrammer.ModeExact)

	for _, seven := range sevens {
		stemDecompose(seven)
	}
	stems := processStems()
	for idx, stem := range stems[:10000] {
		fmt.Println(stem.Printable(idx + 1))
	}
}

func modifiedS(dist lexicon.LetterDistribution) lexicon.LetterDistribution {
	// Takes the distribution and modifies it to add 6 instead of 4 Ss
	// This is to keep in line with Mike Baron's calculation method for MMPR.
	dist.Distribution['S'] = 6
	return dist
}

// Decompose word into all substrings of word length less than 1, and add
// letters to stem map.
func stemDecompose(word string) {
	// w := lexicon.Word{Word: word, Dist: defaultDist}
	// alphagram := w.MakeAlphagram()

	for idx, char := range word {
		stem := word[:idx] + word[idx+1:]
		stemAlpha := lexicon.Word{Word: stem, Dist: defaultDist}.MakeAlphagram()
		_, ok := stemMap[stemAlpha]
		if !ok {
			stemMap[stemAlpha] = make(map[rune]bool)
		}
		stemMap[stemAlpha][char] = true
	}
}

func calcUsableTiles(stemAlpha string, tileMap map[rune]bool) uint8 {
	n := uint8(0)
	for rn := range tileMap {
		n += defaultDist.Distribution[rn] - uint8(strings.Count(stemAlpha, string(rn)))
	}
	if n > 0 {
		n += 2 // The blanks
	}
	return n
}

// Take stemMap and create a sorted slice of stems.
func processStems() []Stem {
	lexInfo := lexicon.LexiconInfo{
		LetterDistribution: modifiedS(lexicon.EnglishLetterDistribution()),
	}
	lexInfo.Initialize()

	tisane := lexInfo.Combinations("AEINST", false)
	stems := []Stem{}
	for stemAlpha, tiles := range stemMap {
		s := Stem{}

		s.Alphagram = stemAlpha
		s.UsableTiles = calcUsableTiles(stemAlpha, tiles)
		s.StemCombinations = lexInfo.Combinations(stemAlpha, false)
		s.ModifiedStemProbability = 1.5 * (float64(s.StemCombinations) / float64(tisane))
		s.MMPR = s.ModifiedStemProbability * float64(s.UsableTiles)
		stems = append(stems, s)
	}
	sort.Sort(ByMMPR(stems))
	return stems
}

type ByMMPR []Stem

func (a ByMMPR) Len() int      { return len(a) }
func (a ByMMPR) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByMMPR) Less(i, j int) bool {
	if a[i].MMPR == a[j].MMPR {
		return a[i].Alphagram < a[j].Alphagram
	}
	return a[i].MMPR > a[j].MMPR
}
