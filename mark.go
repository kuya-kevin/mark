// Copyright 2011 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Generating random text: a Markov chain algorithm

Based on the program presented in the "Design and Implementation" chapter
of The Practice of Programming (Kernighan and Pike, Addison-Wesley 1999).
See also Computer Recreations, Scientific American 260, 122 - 125 (1989).

A Markov chain algorithm generates text by creating a statistical model of
potential textual suffixes for a given prefix. Consider this text:

	I am not a number! I am a free man!

Our Markov chain algorithm would arrange this text into this set of prefixes
and suffixes, or "chain": (This table assumes a prefix length of two words.)

	Prefix       Suffix

	"" ""        I
	"" I         am
	I am         a
	I am         not
	a free       man!
	am a         free
	am not       a
	a number!    I
	number! I    am
	not a        number!

To generate text using this table we select an initial prefix ("I am", for
example), choose one of the suffixes associated with that prefix at random
with probability determined by the input statistics ("a"),
and then create a new prefix by removing the first word from the prefix
and appending the suffix (making the new prefix is "am a"). Repeat this process
until we can't find any suffixes for the current prefix or we exceed the word
limit. (The word limit is necessary as the chain table may contain cycles.)

Our version of this program reads text from standard input, parsing it into a
Markov chain, and writes generated text to standard output.
The prefix and output lengths can be specified using the -prefix and -words
flags on the command-line.
*/
package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

// Prefix is a Markov chain prefix of one or more words.
type Prefix []string

// String returns the Prefix as a string (for use as a map key).
func (p Prefix) String() string {
	return strings.Join(p, " ")
}

// Shift removes the first word from the Prefix and appends the given word.
func (p Prefix) Shift(word string) {
	copy(p, p[1:])
	p[len(p)-1] = word
}

// Chain contains a map ("chain") of prefixes to a list of suffixes.
// A prefix is a string of prefixLen words joined with spaces.
// A suffix is a single word. A prefix can have multiple suffixes.
type Chain struct {
	chain         map[string][]string       // used in READ, holds a prefix and suffix
	prefixLen     int                       // order of the markov chain
	text          [][]string                // used as a helper to make modelfile formatted in lexicographical order
	pairmap       map[string]map[string]int // used in GENERATE, takes in text from modelfile and generates output from this data structure
	prefixStorage []string                  // used as a helper to check if the current prefix has already been used
}

// NewChain returns a new Chain with prefixes of prefixLen words.
func NewChain(prefixLen int) *Chain {
	return &Chain{make(map[string][]string), prefixLen, make([][]string, 0), make(map[string]map[string]int), make([]string, 0)}
}

// helper function: checks if a string is in a list of strings
func Find(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}

// Build reads text from the provided Reader and
// parses it into prefixes and suffixes that are stored in Chain.
func (c *Chain) Build(r io.Reader) {
	br := bufio.NewReader(r)
	p := make(Prefix, c.prefixLen) //make a list of strings w/ length prefixLen
	currtext := make([]string, 0)
	for {
		var s string
		if _, err := fmt.Fscan(br, &s); err != nil {
			break
		}

		currtext = append(currtext, s)

		key := p.String() //joins elements of p together, separates with " "

		if c.pairmap[key] == nil { // if there is not already an instance of c.[key], make one
			c.pairmap[key] = make(map[string]int)
		}
		c.pairmap[key][s] += 1

		if !Find(c.chain[key], s) { //if the current string is not already in the set of choices, add it
			c.chain[key] = append(c.chain[key], s)
		}

		//fmt.Println(c.chain[key])
		p.Shift(s)
	}
	c.text = append(c.text, currtext) //add the current block of text into the list of texts
	//fmt.Println(c.text) //debug

}

// WriteModel writes a model frequency file with correct formatting to outFile, specified by user
func (c *Chain) WriteModel(outFile string) { //use chain variables.
	f, err := os.Create(outFile)
	if err != nil {
		panic("problem creating output file")
	}
	defer f.Close()

	fmt.Fprintln(f, c.prefixLen) //add order number to first line of outputFile

	for j := range c.text {
		p := make(Prefix, c.prefixLen)
		for i := 0; i < len(c.text[j]); i++ {
			choices := c.chain[p.String()] //returns map of suffixes + freqs
			if len(choices) == 0 {         // if no more suggested words, break
				break
			}

			next := c.text[j][i] //next word is chosen, used when shifting the prefix over a word

			line := make([]string, 0) //holds the whole row of input: prefix + suffixes + freq

			if !Find(c.prefixStorage, p.String()) { // if  current prefix has not already been used
				c.prefixStorage = append(c.prefixStorage, p.String())
				// format current prefix
				for _, word := range p {
					if word == "" { //format empty spaces
						line = append(line, strconv.Quote(word))
					} else {
						line = append(line, word)
					}
				} //end prefix for

				//format chosen suffix
				for _, word := range choices {
					if word == "" {
						line = append(line, strconv.Quote(word))
						//do I need to add freq of "" ?
					} else {
						line = append(line, word)
						line = append(line, strconv.Itoa(c.pairmap[p.String()][word]))
					}
				}

				line_string := strings.Join(line, " ") //creates a string from list
				fmt.Fprintln(f, line_string)           //prints to outputfile
				p.Shift(next)
			} else {
				p.Shift(next)
			} //end if

		} //end for

	}
	//fmt.Println("text length is ", c.text)
	//prefixStorage := make([]string, 0)

}

// Generate returns a string of at most n words generated from Chain.
// modify: read from the modelfreq table you generate in Build()
// modify: output to use the frequencies you've stored.
func (c *Chain) Generate(n int) string {

	randIndex := rand.Intn(len(c.prefixStorage)) // choose random index to start from

	p := make(Prefix, c.prefixLen)
	p = strings.Split(c.prefixStorage[randIndex], " ") // converts the string key stored in c.pairmap into a prefix
	var words []string                                 // holds output

	words = append(words, p.String())
	for i := 0; i < n-c.prefixLen; i++ {
		choices := c.pairmap[p.String()] //returns a map
		if len(choices) == 0 {           //at end of chain
			break
		}

		keys := make([]string, 0, len(choices)) // this list will hold the suffix strings w/ their relative frequencies
		for key := range choices {              // for each suffix string
			for j := 0; j < choices[key]; j++ { //for each time we saw the suffix
				// append the key once for each time it was recorded in the freqmap
				// for example: suffix["how":4, "can": 2] gets recorded in keys[] as -> [how, how, how, how, can, can]
				keys = append(keys, key)

			}
		}
		// choose a random suffix within keys[]
		var suffix string
		suffix = keys[rand.Intn(len(keys))]

		words = append(words, suffix)
		p.Shift(suffix)

	}
	return strings.Join(words, " ")
}

// ReadChainFromFile takes in a filename from user and generates a frequency table, stored in c.pairmap[]
func ReadChainFromFile(modelfile string) *Chain {

	model, err := os.Open(modelfile)
	if err != nil {
		panic("error opening file")
	}
	defer model.Close()

	scanner := bufio.NewScanner(model)
	scanner.Scan()

	order, err := strconv.Atoi(scanner.Text()) // scan first line for markov order number
	if err != nil {
		panic("error parsing integer value from line 1")
	}

	c := NewChain(order)

	for scanner.Scan() { // scan one line at a time
		fields := strings.Fields(scanner.Text())

		// fix if quotes in line
		for index, str := range fields {
			if str == strconv.Quote("") {
				fields[index] = str[1 : len(str)-1]
			}
		}

		// join first two elements as a prefix, everything after are suffixes
		pre := strings.Join(fields[0:order], " ")
		c.prefixStorage = append(c.prefixStorage, pre)
		// if we have not seen a previous instance of prefix in the pairmap, make one
		if c.pairmap[pre] == nil {
			c.pairmap[pre] = make(map[string]int)
		}

		// format suffixes, add prefixes and suffixes to pairmap w frequency
		numSuffix := len(fields[order:]) / 2
		sIndex := order
		for i := 0; i < numSuffix; i++ { // iterate thru the suffix strings
			freq, err := strconv.Atoi(fields[sIndex+i+1]) // look at the frequencies, not the suffix string
			if err != nil {
				fmt.Println("error here: current field is: ", fields)
			}
			c.pairmap[pre][fields[sIndex+i]] += freq
			sIndex += 1
		}
	}
	return c
}

// run: ./mark read N outfilename infile1 infile2... (N: order of the chain, int. any number of input files)
// read each inputfile, make one collective freq table, sort the freq table, save freq table to outile
// run: ./mark generate modelfile N (modelfile: name of saved file from READ, n: words to output, int.)
// read freq table in modelfile, use it to generate n words of output
func main() {
	// Register command-line flags.
	rand.Seed(time.Now().UnixNano()) // Seed the random number generator.
	//numWords := flag.Int("words", 100, "maximum number of words to print")
	//prefixLen := flag.Int("prefix", 2, "prefix length in words")

	flag.Parse() // Parse command-line flags.

	if os.Args[1] == "read" || os.Args[1] == "READ" { //READ start
		order, err := strconv.Atoi(os.Args[2]) //

		outputFile := os.Args[3]
		if err != nil {
			panic("trouble parsing int")
		}
		c := NewChain(order) // Initialize a new Chain, markov length order
		files := os.Args[4:] //filenames start at 4
		for index := range files {
			var collection io.Reader                  //initialize io.Reader
			fmt.Println("reading file", files[index]) //debug
			f, err := os.Open(files[index])
			if err != nil {
				panic("trouble opening file")
			}
			defer f.Close()
			collection = f
			c.Build(collection) // new READ: Build chains/freqmap from file

		}

		// if gotten here: should have built the chain + freqmap in chain struct. Now write to outputfile
		c.WriteModel(outputFile)

	} //end READ

	if os.Args[1] == "generate" || os.Args[1] == "GENERATE" {
		modelfile := os.Args[2]
		numWords, err := strconv.Atoi(os.Args[3])
		if err != nil {
			panic("trouble parsing integer numWords from command line")
		}

		c := ReadChainFromFile(modelfile)
		text := c.Generate(numWords)
		fmt.Println(text)

	}

} // end main
