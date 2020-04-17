// License: MIT

/*
 * Convert strand-specific blastn alignment files to GFFv3 annotation format.
 *
 * Currently intended for use with prokaryotic genomes.
 *
 * TODO: 
 *  1) Sort GFF by coordinates.
 *  2) Handle eukaryotes too.
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
	__method      = "de novo"
	__feature     = "gene"
	__description = "de novo predicted transcript"
	__version     = "0.1"
)

/* Print help text and exit */
func help() {
	fmt.Printf("\nUsage:\n\n")
	fmt.Printf("    Forward strand alignment to GFF:\n")
	fmt.Printf("	~$ blast2gff convert watson alignment_foward_strand.aln > watson.gff\n\n")
	fmt.Printf("    Reverse strand alignment to GFF:\n")
	fmt.Printf("	~$ blast2gff convert crick alignment_reverse_strand.aln > crick.gff\n\n")
	fmt.Printf("    Merge:\n")
	fmt.Printf("	~$ blast2gff merge watson.gff crick.gff > transcript_annotation.gff\n\n")
}

/*
 * Convert BLASTN alignment output to GFFv3
 *
 * TODO: Need to sort the gff by coordinates. Currently unsorted.
 */
func convert(strand, alnfile string) {

	if strand != "watson" {
		if strand != "crick" {
			help()
			os.Exit(0)
		}
	}

	fh, err := os.Open(alnfile)
	if err != nil {
		fmt.Printf("Could not open alignment file: %s\n", alnfile)
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
			b5_count += 1
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

/* Read files concurrently */
func scanfile(file string, queue chan string) {
	fh, err := os.Open(file)
	if err != nil {
		fmt.Printf("Could not open GFF file: %s\n", file)
		os.Exit(0)
	}
	defer fh.Close()
	line := bufio.NewScanner(fh)
	for line.Scan() {
		queue <- line.Text()
	}
	close(queue)
}

/* Merge two GFF files */
func merge(fwdfile, revfile string) {

	fmt.Printf("##gff-version 3\n")
	fmt.Printf("#!build blast2gff/de_novo\n")

	queue := make(chan string)
	go scanfile(revfile, queue)

	fh, err := os.Open(fwdfile)
	if err != nil {
		fmt.Printf("Could not open GFF file: %s", fwdfile)
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
		println("No commands given. Try 'blast2gff help'")
		os.Exit(0)
	}

	if os.Args[1] == "version" {
		fmt.Printf("blast2gff v%s\n", __version)
		os.Exit(0)
	}

	if os.Args[1] == "help" {
		fmt.Printf("blast2gff v%s\n", __version)
		help()
		os.Exit(0)
	}

	if len(os.Args) != 4 {
		println("Whoops! Try 'blast2gff help'")
		os.Exit(0)
	}

	switch os.Args[1] {

	case "convert":
		convert(os.Args[2], os.Args[3])

	case "merge":
		merge(os.Args[2], os.Args[3])

	default:
		println("Whoops! Try 'blast2gff help'")

	}
}
