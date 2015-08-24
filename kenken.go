package main

import "strings"
import "strconv"
import "fmt"

type region struct {
	origin    [2]int
	locations [][2]int
	knowledge [][]int
	op        byte
	value     int
}

type kenken struct {
	size             int
	regions          []region
	regionByLocation [][]*region
	knowledge        [][][]int
}

func parse(s string) kenken {
	var hconnections [][]bool
	var vconnections [][]bool
	var k kenken

	lines := strings.Split(s, "\n")
	lines = lines[2 : len(lines)-2]

	// Initialize size
	k.size = len(lines)/2 + 1

	// Initialize and compute connections
	hconnections = make([][]bool, k.size)
	vconnections = make([][]bool, k.size)
	for column := range hconnections {
		hconnections[column] = make([]bool, k.size)
		vconnections[column] = make([]bool, k.size)
	}
	k.knowledge = make([][][]int, k.size)
	for column := range k.knowledge {
		k.knowledge[column] = make([][]int, k.size)
		for row := range k.knowledge[column] {
			k.knowledge[column][row] = make([]int, k.size)
		}
	}

	k.regionByLocation = make([][]*region, k.size)
	for column := range k.regionByLocation {
		k.regionByLocation[column] = make([]*region, k.size)
	}

	// Initialize and compute regions.
	for linenum, line := range lines {
		row := linenum / 2
		if linenum%2 == 0 {
			opStrings := strings.Split(strings.Replace(line, ".", "|", -1), "|")[1:]
			for column, opString := range opStrings {
				s := strings.Replace(opString, " ", "", -1)
				if s != "" {
					var r region
					r.op = s[len(s)-1]
					r.value, _ = strconv.Atoi(s[:len(s)-1])
					r.origin = [2]int{column, row}
					k.regions = append(k.regions, r)
				}
			}
		}

		column := 0
		for _, char := range line[1 : len(line)-1] {
			if (linenum % 2) == 1 {
				if char == '+' {
					column += 1
				}
				if char == ' ' {
					vconnections[column][row] = true
				}
			} else {
				if char == '.' || char == '|' {
					column += 1
				}
				if char == '.' {
					hconnections[column-1][row] = true
				}
			}
		}
	}

	for i, r := range k.regions {
		k.regionByLocation[r.origin[0]][r.origin[1]] = &k.regions[i]
	}

	var done = false
	for done == false {
		done = true
		for column := range k.knowledge {
			for row := range k.knowledge {
				if vconnections[column][row] {
					if k.regionByLocation[column][row] == nil {
						k.regionByLocation[column][row] = k.regionByLocation[column][row+1]
					}
					if k.regionByLocation[column][row+1] == nil {
						k.regionByLocation[column][row+1] = k.regionByLocation[column][row]
					}
				}
				if hconnections[column][row] {
					if k.regionByLocation[column][row] == nil {
						k.regionByLocation[column][row] = k.regionByLocation[column+1][row]
					}
					if k.regionByLocation[column+1][row] == nil {
						k.regionByLocation[column+1][row] = k.regionByLocation[column][row]
					}
				}
				if k.regionByLocation[column][row] == nil {
					done = false
				}
			}
		}
	}

	for column := range k.knowledge {
		for row := range k.knowledge {
			k.regionByLocation[column][row].locations = append(k.regionByLocation[column][row].locations, [2]int{column, row})
		}
	}

	for i := range k.regions {
		r := &(k.regions[i])
		r.knowledge = make([][]int, len(r.locations))
		for j := range r.knowledge {
			r.knowledge[j] = make([]int, k.size)
		}
	}

	return k
}

func checkRegion(knowledge [][]int, op byte, target int) bool {
	var ret bool

	if len(knowledge) == 0 {
		if op == '+' {
			return target == 0
		}
		return target == 1
	}

	for i := range knowledge[0] {
		number := i + 1
		good := false

		if knowledge[0][i] & 2 == 2 {
			continue
		}

		if op == '+' {
			good = checkRegion(knowledge[1:], op, target - number) || good
		}

		if op == '*' {
			if (target % number) == 0 {
				good = checkRegion(knowledge[1:], op, target / number) || good
			}
		}

		if op == '-' {
			if len(knowledge) == 2 {
				good = checkRegion(knowledge[1:], op, number + target)
				good = checkRegion(knowledge[1:], op, number - target) || good
			} else {
				good = number == target
			}
		}

		if op == '/' {
			if len(knowledge) == 2 {
				good = checkRegion(knowledge[1:], op, number * target) || good
				if (number % target) == 0 {
					good = checkRegion(knowledge[1:], op, number / target) || good
				}
			} else {
				good = number == target
			}
		}

		if good {
			knowledge[0][i] |= 4
			ret = true
		}
	}

	return ret
}

func countKnowledge(k kenken) (count int) {
	for column := range k.knowledge {
		for row := range k.knowledge[column] {
			for value := range k.knowledge[column][row] {
				if k.knowledge[column][row][value] > 0 {
					count += 1
				}
			}
		}
	}
	return
}

func countOnes(in int) int {
	i := uint(in)
	i = (i & 0x5555) + ((i >> 1) & 0x5555)
	i = (i & 0x3333) + ((i >> 2) & 0x3333)
	i = (i & 0x0f0f) + ((i >> 4) & 0x0f0f)
	i = (i & 0x00ff) + ((i >> 8) & 0x00ff)
	return int(i)
}

func checkTwos(k kenken) {
	possibles := make ([][]int, k.size)
	for column := range possibles {
		possibles[column] = make([]int, k.size)
		for row := range k.knowledge {
			for value := range k.knowledge {
				if k.knowledge[column][row][value] == 0 {
					possibles[column][row] |= 1 << uint(value)
				}
				if k.knowledge[column][row][value] == 1 {
					possibles[column][row] = 0
					break
				}
			}
		}
	}

	// columns
	for column := range k.knowledge {
		for row0 := range k.knowledge {
			for row1 := range k.knowledge {
				if row1 <= row0 {
					continue
				}
				if possibles[column][row0] == possibles[column][row1] {
					if countOnes(possibles[column][row0]) == 2 {
						for row := range k.knowledge {
							if row == row0 {
								continue
							}
							if row == row1 {
								continue
							}
							for value := range k.knowledge {
								if k.knowledge[column][row0][value] == 0 {
									k.knowledge[column][row][value] = 2
								}
							}
						}
					}
				}
			}
		}
	}

	// rows
	for row := range k.knowledge {
		for column0 := range k.knowledge {
			for column1 := range k.knowledge {
				if column1 <= column0 {
					continue
				}
				if possibles[column0][row] == possibles[column1][row] {
					if countOnes(possibles[column0][row]) == 2 {
						for column := range k.knowledge {
							if column == column0 {
								continue
							}
							if column == column1 {
								continue
							}
							for value := range k.knowledge {
								if k.knowledge[column0][row][value] == 0 {
									k.knowledge[column][row][value] = 2
								}
							}
						}
					}
				}
			}
		}
	}
}

func propagate(k kenken) {
	// Locations
	for column := range k.knowledge {
		for row := range k.knowledge {
			sum := 0
			for value := range k.knowledge {
				sum += k.knowledge[column][row][value]
			}
			for value := range k.knowledge {
				exclusiveSum := sum - k.knowledge[column][row][value]
				if exclusiveSum == 2 * (k.size - 1) {
					k.knowledge[column][row][value] = 1
				}
				if exclusiveSum & 1 == 1 {
					k.knowledge[column][row][value] = 2
				}
			}
		}
	}

	// Columns
	for row := range k.knowledge {
		for value := range k.knowledge {
			sum := 0
			for column := range k.knowledge {
				sum += k.knowledge[column][row][value]
			}
			for column := range k.knowledge {
				exclusiveSum := sum - k.knowledge[column][row][value]
				if exclusiveSum == 2 * (k.size - 1) {
					k.knowledge[column][row][value] = 1
				}
				if exclusiveSum & 1 == 1 {
					k.knowledge[column][row][value] = 2
				}
			}
		}
	}

	// Rows
	for column := range k.knowledge {
		for value := range k.knowledge {
			sum := 0
			for row := range k.knowledge {
				sum += k.knowledge[column][row][value]
			}
			for row := range k.knowledge {
				exclusiveSum := sum - k.knowledge[column][row][value]
				if exclusiveSum == 2 * (k.size - 1) {
					k.knowledge[column][row][value] = 1
				}
				if exclusiveSum & 1 == 1 {
					k.knowledge[column][row][value] = 2
				}
			}
		}
	}
}

func printsolved(k kenken) {
	fmt.Println()
	for row := range k.knowledge {
		for column := range k.knowledge[row] {
			for value := range k.knowledge[row][column] {
				if k.knowledge[column][row][value] == 1 {
					fmt.Printf("%d ", value + 1)
				}
			}
		}
		fmt.Println()
	}
}

func printk(k kenken) {
	return
	for row := range k.knowledge {
		for column := range k.knowledge[row] {
			for value := range k.knowledge[row][column] {
				fmt.Printf("%d", k.knowledge[column][row][value])
			}
			fmt.Print(" ")
		}
		fmt.Println()
	}
	fmt.Println(countKnowledge(k))
	//fmt.Println(errors(k))
}

/*
func errors(k kenken) int {

	solution := [][]int{
	[]int{2, 1, 5, 8, 6, 3, 9, 7, 4},
	[]int{6, 4, 3, 5, 2, 9, 7, 1, 8},
	[]int{9, 8, 2, 6, 5, 7, 3, 4, 1},
	[]int{5, 6, 9, 4, 8, 2, 1, 3, 7},
	[]int{8, 5, 7, 3, 9, 1, 4, 2, 6},
	[]int{4, 3, 6, 1, 7, 8, 5, 9, 2},
	[]int{7, 2, 4, 9, 1, 5, 8, 6, 3},
	[]int{1, 7, 8, 2, 3, 4, 6, 5, 9},
	[]int{3, 9, 1, 7, 4, 6, 2, 8, 5}}

	errors := 0
	for column := range k.knowledge {
		for row := range k.knowledge {
			for value := range k.knowledge {
				if k.knowledge[column][row][value] == 2 {
					if solution[row][column] == value + 1{
						fmt.Println(column, row, value, 2)
						errors++
					}
				}
				if k.knowledge[column][row][value] == 1 {
					if solution[row][column] != value + 1{
						errors++
						fmt.Println(column, row, value, 1)
					}
				}
			}
		}
	}
	return errors
}
*/

func solve(k kenken) {

	knowledgeCount := 0
	for {
		for _, r := range k.regions {
			for i := range r.knowledge {
				location := r.locations[i]
				for value := range k.knowledge {
					r.knowledge[i][value] = k.knowledge[location[0]][location[1]][value]
				}
			}

			checkRegion(r.knowledge, r.op, r.value)

			for i := range r.knowledge {
				location := r.locations[i]
				for value := range k.knowledge {
					if r.knowledge[i][value] & 4 == 0 {
						k.knowledge[location[0]][location[1]][value] |= 2
					}
				}
			}
		}

		printk(k)
		propagate(k)

		printk(k)
		checkTwos(k)
		printk(k)

		newKnowledgeCount := countKnowledge(k)
		if knowledgeCount == newKnowledgeCount {
			break
		}
		knowledgeCount = newKnowledgeCount
	}
}

func main() {

x := []string{`
+-----+-----+-----+-----+-----+-----+-----+-----+-----+
|15*  |2/   |2/   |25+  .     .     .     |320* .     |
+     +     +     +-----+-----+-----+-----+-----+     +
|     |     |     |11+  |5+   |1-   |3-   .     |     |
+     +-----+-----+     +     +     +-----+-----+     +
|     |45*  |42*  |     |     |     |2/   |14+  |     |
+-----+     +     +-----+-----+-----+     +     +-----+
|14*  |     |     |80*  .     |5+   |     |     .     |
+     +-----+-----+-----+     +     +-----+-----+-----+
|     |12+  .     .     |     |     |14*  |45*  |2/   |
+-----+-----+-----+-----+-----+-----+     +     +     +
|2/   |13+  |1-   |1-   |14*  .     |     |     |     |
+     +     +     +     +-----+-----+     +-----+-----+
|     |     |     |     |6-   .     |     |5-   .     |
+-----+-----+-----+-----+-----+-----+-----+-----+-----+
|11+  .     |3/   |28*  |4/   |11+  |3-   |1-   .     |
+-----+-----+     +     +     +     +     +-----+-----+
|5-   .     |     |     |     |     |     |1-   .     |
+-----+-----+-----+-----+-----+-----+-----+-----+-----+
`,`
+-----+-----+-----+-----+-----+-----+-----+-----+-----+
|2/   |1-   |3/   |56*  |4-   .     |11+  |2/   .     |
+     +     +     +     +-----+-----+     +-----+-----+
|     |     |     |     |2-   |72*  |     |17+  |3/   |
+-----+-----+-----+-----+     +     +-----+     +     +
|7-   .     |24*  .     |     |     |2/   |     |     |
+-----+-----+-----+-----+-----+-----+     +     +-----+
|14+  |12*  |11+  .     |9+   .     |     |     |54*  |
+     +     +-----+-----+-----+-----+-----+-----+     +
|     |     |90*  |11+  .     |3/   |320* .     |     |
+-----+-----+     +-----+-----+     +-----+     +-----+
|15+  |     .     |16*  .     |     |5-   |     |9+   |
+     +-----+-----+-----+     +-----+     +-----+     +
|     .     |7*   .     |     |3/   |     |45*  |     |
+-----+-----+-----+-----+-----+     +-----+     +-----+
|4-   |2-   .     |13+  .     |     |     .     .     |
+     +-----+-----+-----+-----+-----+-----+-----+-----+
|     |2/   .     |1-   .     |3-   .     |1-   .     |
+-----+-----+-----+-----+-----+-----+-----+-----+-----+
`,`
+-----+-----+-----+-----+-----+-----+-----+-----+-----+
|432* .     |11+  .     |15*  .     |7+   |9+   |2/   |
+     +-----+-----+-----+-----+-----+     +     +     +
|     |5-   .     |72*  |4-   |2/   |     |     |     |
+-----+-----+-----+     +     +     +-----+-----+-----+
|21*  .     |     .     |     |     |14+  .     |15+  |
+-----+-----+-----+-----+-----+-----+-----+-----+     +
|4/   .     |2-   .     |2-   |17+  |5-   |11+  |     |
+-----+-----+-----+-----+     +     +     +     +-----+
|36*  |6+   .     |3-   |     |     |     |     |9*   |
+     +-----+-----+     +-----+     +-----+-----+     +
|     |5-   |24*  |     |7-   |     |1-   .     |     |
+     +     +     +-----+     +-----+-----+-----+-----+
|     |     |     |8+   |     |9+   .     |42*  .     |
+-----+-----+-----+     +-----+-----+-----+-----+-----+
|2/   .     |54*  |     |1-   .     |8*   |15*  .     |
+-----+-----+     +-----+-----+-----+     +-----+-----+
|3-   .     |     |42*  .     |     .     |7+   .     |
+-----+-----+-----+-----+-----+-----+-----+-----+-----+
`,`
+-----+-----+-----+-----+-----+-----+-----+-----+-----+
|3/   |4/   |2-   |13+  |2/   .     |23+  .     |2/   |
+     +     +     +     +-----+-----+     +-----+     +
|     |     |     |     |7-   .     |     |4/   |     |
+-----+-----+-----+-----+-----+-----+-----+     +-----+
|7776*.     .     |2-   |80*  |21*  .     |     |6-   |
+-----+     +     +     +     +-----+-----+-----+     +
|3-   |     .     |     |     .     |6*   .     |     |
+     +-----+-----+-----+-----+-----+-----+     +-----+
|     |12+  .     |3/   .     |20+  .     |     |3/   |
+-----+-----+-----+-----+-----+     +-----+-----+     +
|3-   |5+   |5-   .     |     .     |4-   .     |     |
+     +     +-----+-----+-----+-----+-----+-----+-----+
|     |     |12+  |18*  |48*  |40*  .     |30*  |3/   |
+-----+-----+     +     +     +-----+-----+     +     +
|3/   |2-   |     |     |     .     |14+  |     |     |
+     +     +-----+-----+     +-----+     +-----+-----+
|     |     |7*   .     |     |     .     |13+  .     |
+-----+-----+-----+-----+-----+-----+-----+-----+-----+
`,`
+---+---+---+---+---+---+
|3/ .   |15*.   |24*|3- |
+---+---+---+---+   +   +
|6* .   |6+ .   |   |   |
+---+---+---+---+---+----
|6* |1- .   |2/ .   |6* |
+   +---+---+---+---+   +
|   |5+ .   |8+ .   |   |
+---+---+---+---+---+----
|4- .   |6+ |3/ |3/ |1- |
+---+---+   +   +   +   +
|1- .   |   |   |   |   |
+---+---+---+---+---+---+
`}

	for i := range x {
		k := parse(x[i])
		solve(k)
		fmt.Print(x[i])
		printsolved(k)
		printk(k)
	}
}
