// License: MIT

/*
 * Convert blastn alignments to GFFv3 file
 */

package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

const (
	__reference   = "Chromosome"
	__method      = "denovo"
	__feature     = "gene"
	__description = "de novo predicted transcript"
	__version     = "0.1-bugfix"
)

func convert(strand, alnfile string) {
	//Parse a BLASTN alignment file and convert it to a GFFv3 file
	/*
	 * TODO: Need to include an additional stage where we
	 * sort the gff by coordinates. Currently it's unsorted.
	 */

	if strand != "watson" {
		if strand != "crick" {
			println("Strand should be either watson or crick")
			os.Exit(0)
		}
	}

	fh, err := os.Open(alnfile)
	if err != nil {
		fmt.Printf("Could not open file: %s\n", alnfile)
		os.Exit(0)
	}
	defer fh.Close()
	line := bufio.NewScanner(fh)
	id_count, b5_count, b3_coord, b5_coord := 0, 0, 0, 0
	var name string
	var split_line []string
	for line.Scan() {

		split_line = strings.Split(line.Text(), " ")

		if strings.Contains(line.Text(), "Query= ") {
			fmt.Printf("%s\t%s\t%s\t", __reference, __method, __feature)
			id_count += 1
			name = strings.Trim(split_line[1], ">")
		}

		split_line = strings.Split(line.Text(), " ")

		// Get coordinates
		if strings.Contains(line.Text(), "Sbjct") {
			if b5_count == 0 {
				b5_coord, _ = strconv.Atoi(split_line[2])
			}
			b5_count += 1//this updates everytime
			// this updates everytime, we just need the last
			// value
			b3_coord, _ = strconv.Atoi(split_line[6])
		}

		if strings.Contains(line.Text(), "Gapped") {
			// Print the strand coord information, phasing, etc
			if strand == "watson" {
				fmt.Printf("%d\t%d\t.\t+\t.\t", b5_coord, b3_coord)
			}
			if strand == "crick" {
				fmt.Printf("%d\t%d\t.\t-\t.\t", b3_coord, b5_coord)
			}

			b3_coord = 0
			b5_coord = 0
			fmt.Printf("ID \"%s\"; gene_id \"%s\"; Type: \"%s\"\n",
				name, name, __description)
		}

		if strings.Contains(line.Text(), "Effective") {
			id_count = 0
			b5_count = 0
		}
	}
}

func scanfile(file string, queue chan string) {
	// scanfile allows reading two files concurrently
	fh, err := os.Open(file)
	if err != nil {
		fmt.Printf("Could not open file: %s\n", file)
		os.Exit(0)
	}
	defer fh.Close()
	line := bufio.NewScanner(fh)
	for line.Scan() {
		queue <- line.Text()
	}
	close(queue)
}

func merge(fwdfile, revfile string) {
	// Merge two GFF files
	fmt.Printf("##gff-version 3\n")
	fmt.Printf("#!build blast2gff/denovo\n")

	queue := make(chan string)
	go scanfile(revfile, queue)

	fh, err := os.Open(fwdfile)
	if err != nil {
		fmt.Printf("Could not open file: %s", fwdfile)
		os.Exit(0)
	}
	defer fh.Close()

	line := bufio.NewScanner(fh)
	var split_line []string

	for line.Scan() {
		fwd_text := line.Text()
		rev_text := <-queue
		split_line = strings.Split(fwd_text, "\t")

		if split_line[3] == "0" {
			split_line = strings.Split(rev_text, "\t")
			if split_line[3] == "0" {
				continue
			}
			fmt.Println(rev_text)
		} else {
			fmt.Println(fwd_text)
		}
	}
}

func main() {

	if len(os.Args) == 1 {
		println("No commands given")
		os.Exit(0)
	}

	if os.Args[1] == "version" {
		fmt.Printf("blast2gff, built on: %s\n", __version)
		os.Exit(0)
	}

	if len(os.Args) != 4 {
		println("Invalid arguments")
		os.Exit(0)
	}

	switch os.Args[1] {

	case "convert":
		convert(os.Args[2], os.Args[3])

	case "merge":
		merge(os.Args[2], os.Args[3])

	default:
		println("Unrecognized command")

	}
}
