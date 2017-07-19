package main

import (
	"flag"
	"fmt"
	"sort"
	"strings"

	"github.com/domino14/macondo/anagrammer"
	"github.com/domino14/macondo/lexicon"
)

var (
	// Flagged values.
	dawgPath   = flag.String("dawgpath", "", "path for dawgs")
	searchType = flag.String("type", "type1",
		"the type of search - type1, type2, type3, stems")
	// Type 1 is bingos in the top 100 stems by MMPR. Unfortunately this
	// changed since OWL2..
	// Type 2 is the letters in U DATE NO GIRLS
	// with the following repetition frequencies
	// A(3) D(1) E(4) G(1) I(3) L(1) N(2) O(2) R(2) S(2) T(2) U(1)
	// Type 3 is 7s with >= probability than HUNTERS and 8s >= NOTIFIED
	// (why NOTIFIED? ask M.B...)
	// Stems generates a list of stems of length `searchLength - 1`, sorted
	// by MMPR.

	topNStems    = flag.Int("topn", 100, "top N stems of this length")
	searchLength = flag.Int("length", 7, "length of word")
	stemLexicon  = flag.String("lexicon", "America", "name of lexicon")
)

var defaultDist lexicon.LetterDistribution

type Stem struct {
	StemCombinations        uint64
	ModifiedStemProbability float64
	UsableTiles             uint8
	MMPR                    float64
	Alphagram               string
}

func (s *Stem) Printable(order int) string {
	return fmt.Sprintf("%v\t%v\t%.4f\t%v\t%.4f\t%v", order, s.Alphagram,
		s.ModifiedStemProbability, s.UsableTiles, s.MMPR, s.StemCombinations)
}

func main() {

	flag.Parse()
	if *searchLength != 7 && *searchLength != 8 {
		fmt.Println("Error - stem length must be 7 or 8")
		return
	}
	anagrammer.LoadDawgs(*dawgPath)
	defaultDist = lexicon.EnglishLetterDistribution()

	wordPrinter := func(words []string) {
		for _, word := range words {
			fmt.Println(word)
		}
		fmt.Println(len(words))
	}

	if *searchType == "type1" {
		words, _ := calculateTypeIs(*searchLength)
		for _, word := range words {
			fmt.Println(word)
		}
		fmt.Println(len(words))
	} else if *searchType == "type2" {
		words := calculateTypeIIs(*searchLength)
		wordPrinter(words)

	} else if *searchType == "type3" {
		words := calculateTypeIIIs(*searchLength)
		wordPrinter(words)

	} else if *searchType == "stems" {
		stems := calculateStems(*searchLength, *topNStems)
		fmt.Println("#\talpha\tMSP\tUT\tMMPR\tstemcombos")
		for idx, stem := range stems {
			fmt.Println(stem.Printable(idx + 1))
		}
	}
}

// calculate Type I bingos and put them in a map.
// Type I 7s are made from the top n 6 letter stems
// Type I 8s are also made from the same stems + 2 letters.
func calculateTypeIs(length int) ([]string, map[string]bool) {
	stems := calculateStems(7, 100)
	wordMap := map[string]bool{}
	blanks := "?"
	if length == 8 {
		blanks = "??"
	}
	for _, stem := range stems {
		theseWords := anagrammer.Anagram(stem.Alphagram+blanks,
			anagrammer.Dawgs[*stemLexicon], anagrammer.ModeExact)
		for _, word := range theseWords {
			wordMap[word] = true
		}
	}
	words := []string{}
	for word := range wordMap {
		words = append(words, word)
	}
	sort.Strings(words)
	return words, wordMap
}

func calculateTypeIIs(length int) []string {
	_, type1s := calculateTypeIs(length)
	ret := []string{}
	//A(3) D(1) E(4) G(1) I(3) L(1) N(2) O(2) R(2) S(2) T(2) U(1)
	words := anagrammer.Anagram("AAADEEEEGIIILNNOORRSSTTU",
		anagrammer.Dawgs[*stemLexicon], anagrammer.ModeBuild)
	for _, word := range words {
		if len(word) == length {
			_, ok := type1s[word]
			if !ok {
				ret = append(ret, word)
			}
		}
	}
	sort.Strings(ret)
	return ret
}

func calculateTypeIIIs(length int) []string {
	var refCombos uint64
	lexInfo := lexicon.LexiconInfo{
		LetterDistribution: lexicon.EnglishLetterDistribution(),
	}
	lexInfo.Initialize()
	if length == 7 {
		refCombos = lexInfo.Combinations("EHNRSTU", false)
	} else if length == 8 {
		refCombos = lexInfo.Combinations("DEFIINOT", false)
	}
	_, wordMap := calculateTypeIs(length)
	typeIIs := calculateTypeIIs(length)
	// insert typeIIs to word map
	for _, word := range typeIIs {
		wordMap[word] = true
	}
	typeIIIs := []string{}

	for _, word := range allWords(length) {
		// Check if it's not in map.
		_, ok := wordMap[word]
		if !ok {
			if lexInfo.Combinations(word, false) >= refCombos {
				typeIIIs = append(typeIIIs, word)
			}
		}
	}
	sort.Strings(typeIIIs)
	return typeIIIs
}

func allWords(length int) []string {
	return anagrammer.Anagram(strings.Repeat("?", length),
		anagrammer.Dawgs[*stemLexicon], anagrammer.ModeExact)
}

// Calculate the top `stemCutoff` stems of length `length - 1`
func calculateStems(length int, stemCutoff int) []Stem {
	words := allWords(length)
	stemMap := map[string]map[rune]bool{}
	for _, word := range words {
		stemDecompose(word, stemMap)
	}
	stems := processStems(stemMap)
	if stemCutoff > len(stems)-1 {
		stemCutoff = len(stems) - 1
	}
	return stems[:stemCutoff]
}

func modifiedS(dist lexicon.LetterDistribution) lexicon.LetterDistribution {
	// Takes the distribution and modifies it to add 6 instead of 4 Ss
	// This is to keep in line with Mike Baron's calculation method for MMPR.
	dist.Distribution['S'] = 6
	return dist
}

// Decompose word into all substrings of word length less than 1, and add
// letters to stem map.
func stemDecompose(word string, stemMap map[string]map[rune]bool) {
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

// Calculates the usable tiles for a stem alphagram. Assumes tileMap
// has at least one letter in it.
func calcUsableTiles(stemAlpha string, tileMap map[rune]bool) uint8 {
	n := int8(0)
	if len(tileMap) == 0 {
		panic("stemAlpha invalid: " + stemAlpha)
	}
	for rn := range tileMap {
		n += int8(defaultDist.Distribution[rn]) - int8(strings.Count(stemAlpha, string(rn)))
	}
	n += 2 // The blanks

	return uint8(n)
}

// Take stemMap and create a sorted slice of stems.
func processStems(stemMap map[string]map[rune]bool) []Stem {
	lexInfo := lexicon.LexiconInfo{
		LetterDistribution: modifiedS(lexicon.EnglishLetterDistribution()),
	}
	lexInfo.Initialize()

	var baseStemCombos uint64
	if *searchLength == 7 {
		baseStemCombos = lexInfo.Combinations("AEINST", false)
	} else if *searchLength == 8 {
		baseStemCombos = lexInfo.Combinations("AEINRST", false)
	}
	stems := []Stem{}
	for stemAlpha, tiles := range stemMap {
		s := Stem{}

		s.Alphagram = stemAlpha
		s.UsableTiles = calcUsableTiles(stemAlpha, tiles)
		s.StemCombinations = lexInfo.Combinations(stemAlpha, false)
		s.ModifiedStemProbability = 1.5 * (float64(s.StemCombinations) / float64(baseStemCombos))
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
