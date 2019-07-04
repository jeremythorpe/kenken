
package main

import "fmt"
import "io/ioutil"
import "math"
import "math/rand"
import "net/http"
import "os"
import "sort"
import "strings"

type scrabble struct {
	counts map[byte]int
	values map[byte]int
	dict map[string]bool
	lagrange map[byte]float64
}

func (scr *scrabble) get_words() {
	words_file := "/tmp/words"
	if _, err := os.Stat(words_file); os.IsNotExist(err) {
		url := "https://norvig.com/ngrams/enable1.txt"
		resp, err := http.Get(url)
		if err != nil { panic(err) }
		html, err := ioutil.ReadAll(resp.Body)
		if err != nil { panic(err) }
		ioutil.WriteFile(words_file, html, 0644)
	}
	content, err := ioutil.ReadFile(words_file)
	if err != nil { panic(err) }
	words := strings.Split(string(content), "\n")
	scr.dict = map[string]bool{}
	for _, word := range words {
		scr.dict[word] = true
	}
}

func (scr *scrabble) subwords(word string) []string {
	var ret []string
	for i := 0; i < len(word); i++ {
		for j := 1 + i; j <= len(word); j++ {
			if scr.dict[strings.ToLower(word[i:j])] {
				ret = append(ret, word[i:j])
			}
		}
	}
	return ret
}

func (scr *scrabble) set_lagrange() {
	dict_counts := map[byte]int{}
	for word := range scr.dict {
		for _, letter := range word {
			dict_counts[byte(letter)] += 1
		}
	}
	scr.lagrange = map[byte]float64{}
	for letter := range scr.counts {
		scr.lagrange[letter] = 10.0
		scr.lagrange[letter] -= 0.5 * math.Log(float64(dict_counts[letter]))
		scr.lagrange[letter] += 0.2 * math.Pow(float64(scr.counts[letter]), 0.2) * float64(scr.values[letter])
	}
	scr.lagrange['v'] += -4
	scr.lagrange['c'] += -1
	scr.lagrange['f'] += -1
	scr.lagrange['g'] += -1
	scr.lagrange['a'] += -4
	scr.lagrange['e'] += -4
	scr.lagrange['i'] += -4
	scr.lagrange['o'] += -4
	scr.lagrange['u'] += -4
	scr.lagrange['*'] = 6.0
}

func (scr *scrabble) randomize_lagrange() {
	for letter := range scr.lagrange {
		scr.lagrange[letter] += 6.0 * rand.NormFloat64()
	}
}

func (scr *scrabble) get_value(word string) int {
	total := 0
	for _, subword := range scr.subwords(word) {
		for _, letter := range subword {
			total += scr.values[byte(letter)]
		}
	}
	return total
}

func (scr *scrabble) limit_to_set(best map[string]float64) {
	for word := range best {
		counts := map[byte]int{}
		for _, letter := range word {
			counts[byte(letter)] += 1
		}
		for letter, count := range counts {
			if count > scr.counts[letter] {
				delete(best, word)
				break
			}
		}
	}
}

func (scr *scrabble) entropy(word string) float64 {
	counts := map[byte]int{}
	for _, letter := range word {
		counts[byte(letter)] += 1
	}

	ways := 1.
	for letter := byte('a'); letter <= 'z'; letter ++ {
		for i := counts[letter]; i < scr.counts[letter]; i++ {
			ways *= float64(1 + i)
			ways /= float64(1 + i - counts[letter])
		}
	}
	return math.Log(ways)
}

func (scr *scrabble) get_lagrange_value(word string) float64 {
	value := float64(scr.get_value(word))
	for _, letter := range word {
		value -= scr.lagrange[byte(letter)]
	}
	value += scr.lagrange['*'] * scr.entropy(word)
	return value
}

func (scr *scrabble) compliment(word string) string {
	comp := ""
	counts := map[byte]int{}
	for _, letter := range word {
		counts[byte(letter)] += 1
	}
	for letter := byte('a'); letter <= 'z'; letter ++ {
		for i := counts[letter]; i < scr.counts[letter]; i++ {
			comp += string(letter)
		}
	}
	return comp
}

func limit_to(best map[string]float64, n int) {
	if n > len(best) {
		return
	}
	var values []float64
	for _, value := range best {
		values = append(values, value)
	}
	sort.Float64s(values)
	thresh := values[len(values) - n]
	for word, value := range best {
		if value < thresh {
			delete(best, word)
		}
	}
}

func force_q(best map[string]float64, n int) {
	for word := range best {
		if word[0] != 'q' {
			delete(best, word)
		}
	}
}

func add_suffixes(word string, suffix_by_prefix [101]map[string][]string) {
	for i := 1; i < len(word); i++ {
		if i == len(suffix_by_prefix) {
			break
		}
		prefix := word[:i]
		suffix := word[i:]
		suffix_by_prefix[len(suffix)][prefix] = append(suffix_by_prefix[len(suffix)][prefix], suffix)
	}
}

func (scr *scrabble) print_stats(words map[string]float64) {
	stats := map[byte]int{}
	for word := range words {
		for _, letter := range word {
			stats[byte(letter)] += 1
		}
	}
	for letter := byte('a'); letter <= 'z'; letter ++ {
		used := float64(stats[letter]) / (float64(scr.counts[letter] * len(words)))
		fmt.Println("%s: %d", string(letter), used)
	}
}

func (scr *scrabble) print_best_score(words map[string]float64) {
	high_score := 0
	for word := range words {
		score := scr.get_value(word)
		if score > high_score {
			high_score = score
		}
	}
	for word := range words {
		score := scr.get_value(word)
		if score == high_score {
			fmt.Printf("%s %s: %d (%d)\n", word, scr.compliment(word), scr.get_value(word), len(word))
		}
	}
}

func (scr *scrabble) get_best() {
	scr.set_lagrange()
	scr.randomize_lagrange()

	var prefix_by_length [101]map[string]float64
	var suffix_by_prefix [101]map[string][]string
	for i := range(prefix_by_length) {
		prefix_by_length[i] = map[string]float64{}
	}
	for i := range(suffix_by_prefix) {
		suffix_by_prefix[i] = map[string][]string{}
	}
	for word := range(scr.dict) {
		prefix_by_length[len(word)][word] = scr.get_lagrange_value(word)
		add_suffixes(word, suffix_by_prefix)
	}
	for i := range(prefix_by_length) {
		for extension := 1; extension <= 15; extension ++ {
			for overlap := 1; overlap <= 6; overlap ++ {
				if extension + overlap > i {
					continue
				}
				for prefix := range prefix_by_length[i - extension] {
					poverlap := prefix[i - overlap - extension:]
					for _, suffix := range suffix_by_prefix[extension][poverlap] {
						word := prefix + suffix
						prefix_by_length[i][word] = scr.get_lagrange_value(word)
					}
				}
			}
		}
		scr.limit_to_set(prefix_by_length[i])
		limit_to(prefix_by_length[i], 1)
		//scr.print_best_score(prefix_by_length[i])
		//if i == 50 {
		//	scr.print_stats(prefix_by_length[i])
		//}
	}
	fmt.Println(scr.lagrange['*'])
	for i := 95; i <= 100; i++ {
		scr.print_best_score(prefix_by_length[i])
	}
}

func main() {
	var scr scrabble

	scr.counts = map[byte]int {'a': 9, 'b': 2, 'c': 2, 'd': 4, 'e': 12, 'f':2, 'g': 3, 'h': 2, 'i': 9, 'j': 1, 'k': 1, 'l': 4, 'm': 2, 'n': 6, 'o': 8, 'p': 2, 'q': 1, 'r': 6, 's': 4, 't': 6, 'u': 4, 'v': 2, 'w': 2, 'x': 1, 'y': 2, 'z': 1, }
	scr.values = map[byte]int {'a': 1, 'b': 3, 'c': 3, 'd': 2, 'e': 1, 'f':4, 'g': 2, 'h': 4, 'i': 1, 'j': 8, 'k': 5, 'l': 1, 'm': 3, 'n': 1, 'o': 1, 'p': 3, 'q': 10, 'r': 1, 's': 1, 't': 1, 'u': 1, 'v': 4, 'w': 4, 'x': 8, 'y': 4, 'z': 10, }
	scr.get_words()

	scr.counts['s'] += 2

	for word := range scr.dict {
		fmt.Println(word)
		break
	}

	for i := 0; i < 1000; i++ {
		scr.get_best()
	}
	//s := "codeveloperswithereaboutsizedamaskingShallowedjinniforbyeSquiredeveloperSurtaxingoticyanoftagamainut"
	//s := "emblazonerSwitheredevelopersuperoxidescantingabyeSquiredamaskingshallowedjinnifactoryoutavofagoutiti"
	//fmt.Println(scr.get_value(s))
	//fmt.Println(scr.get_value(strings.ToLower(s)))
}
