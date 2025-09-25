package main

import (
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// TODO: Add -o flag for output filename
// TODO: Add -k flag to keep compressed file, otherwise delete by default
// TODO: buffer copy to output file to prevent excess writing
// TODO: Add -l option to allow supply of bytes to modify buffer size
// TODO: Add functionality to convert "G|g, M|m, K|k" to gigs, megs, kilos respectively

func main() {

	if len(os.Args) < 2 {
		log.Fatalln("Must provide 'compress' or 'decompress' subcommand along with filename")
	}

	if len(os.Args) < 3 {
		log.Fatalln("Expected a filename in addition to subcommand")
	}

	switch os.Args[1] {
	case "compress":
		compress(os.Args[2])
	case "decompress":
		decompress(os.Args[2])
	default:
		log.Fatalln("Unknown subroutine:", os.Args[1])
	}

}

func compress(filename string) {

	inFile, err := os.Open(filename)

	if err != nil {
		log.Fatalln("read infile:", filename)
	}
	defer inFile.Close()

	zipFilename := fmt.Sprintf("%s.gz", filename)
	outFile, err := os.Create(zipFilename)

	if err != nil {
		log.Fatalln("create outfile:", err)
	}

	defer outFile.Close()

	wr := gzip.NewWriter(outFile)
	defer wr.Close()

	_, err = io.Copy(wr, inFile)

	if err != nil {
		log.Fatalln("write to compessed file:", err)
	}

}

func decompress(filename string) {

	ext := filepath.Ext(filename)
	if ext != ".gz" {
		log.Fatalln("file format is not .gz")
	}

	// should be .gz after this check, but nothing else is supported
	// so might as well be explicit about what to remove
	outFileName := strings.TrimSuffix(filename, ".gz")

	inFile, err := os.Open(filename)
	if err != nil {
		log.Fatalln("read compressed file:", err)
	}

	defer inFile.Close()

	// create reader before output file in case reader fails
	// we don't then need to worry about cleaning up output file
	r, err := gzip.NewReader(inFile)

	if err != nil {
		log.Fatalln("create gzip reader:", err)
	}

	defer r.Close()

	outFile, err := os.OpenFile(
		outFileName,
		os.O_WRONLY|os.O_CREATE|os.O_EXCL,
		0644,
	)
	defer outFile.Close()

	io.CopyN(outFile, r)

}
