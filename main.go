package main

import (
	"compress/gzip"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var modCharMap = map[string]int64{
	"G": 1e9,
	"M": 1e6,
	"K": 1e3,
}

func main() {

	log.SetFlags(0)

	if len(os.Args) < 2 {
		log.Fatalln("Must provide 'compress' or 'decompress' subcommand along with filename")
	}

	prog := os.Args[1]

	fs := flag.NewFlagSet(prog, flag.ContinueOnError)

	var outputFilename string
	fs.StringVar(&outputFilename, "o", "", "Name of output file. Default value is to copy the input filename, value is ignored if multiple files are passed")

	var copySizeStr string
	fs.StringVar(&copySizeStr, "l", "4G", "Max number of bytes to copy. Defaults to 4G (maximum number of bytes stored in gzip size struct). Accepts in number (i.e. 4096) or with extension (4G, 4M, 4K). If no extension is provided, assumes that the provided number is in bytes.")

	var keepFile bool
	fs.BoolVar(&keepFile, "k", false, "If -k is provided, the input file will not be deleted if operation is completed successfully")

	if err := fs.Parse(os.Args[2:]); err != nil {
		log.Fatalln("unable to parse flags:", err)
	}

	filenames := fs.Args()

	if len(filenames) < 1 {
		log.Fatalln("Expected a filename in addition to subcommand and flags")
	}

	if len(filenames) > 1 && outputFilename != "" {
		log.Print("Warning, output filename will be ignored for multiple input filename options.")
		outputFilename = ""
	}

	copySize := covertStringSizeToBytes(copySizeStr)

	switch prog {
	case "compress":
		compress(outputFilename, copySize, filenames...)
	case "decompress":
		decompress(outputFilename, copySize, filenames...)
	default:
		log.Fatalln("Unknown subroutine:", prog)
	}

	// now we can remove, as if the subroutine exited successfully,
	// we can assume the appropraite files are closed

	if !keepFile {
		for _, file := range filenames {
			err := os.Remove(file)

			// no reason to error terminate if one file
			// can't be removed, just let user know
			if err != nil {
				log.Println("unable to remove:", file, "\n", err, "\n")
			}

		}
	}

}

func compress(outputFilename string, copySize int64, filenames ...string) {

	for _, filename := range filenames {
		inFile, err := os.Open(filename)

		if err != nil {
			log.Fatalln("read infile:", filename)
		}
		defer inFile.Close()

		if outputFilename == "" {
			outputFilename = fmt.Sprintf("%s.gz", filename)
		}

		if filepath.Ext(outputFilename) != ".gz" {
			log.Fatalln("invalid file extension:", filepath.Base(outputFilename))
		}

		outFile, err := os.Create(outputFilename)

		if err != nil {
			log.Fatalln("create outfile:", err)
		}

		defer outFile.Close()

		wr := gzip.NewWriter(outFile)
		defer wr.Close()

		err = compressCopierN(wr, inFile, copySize)

		if err != nil {
			wr.Close()
			os.Remove(outputFilename)
		}

	}

}

func decompress(outputFilename string, copySize int64, filenames ...string) {

	for _, filename := range filenames {
		ext := filepath.Ext(filename)
		if ext != ".gz" {
			log.Fatalln("file format is not .gz:", filepath.Base(filename))
		}

		// should be .gz after this check, but nothing else is supported
		// so might as well be explicit about what to remove

		if outputFilename == "" {
			outputFilename = strings.TrimSuffix(filename, ".gz")
		}

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
			outputFilename,
			os.O_WRONLY|os.O_CREATE|os.O_EXCL,
			0644,
		)

		if err != nil {
			log.Fatalln("unable to open output file:", err)
		}

		defer outFile.Close()

		err = compressCopierN(outFile, r, copySize)

		if err != nil {
			outFile.Close()
			os.Remove(outputFilename)
		}

	}

}

func compressCopierN(w io.Writer, r io.Reader, size int64) error {

	// this feels like a real hack, but it's all I can come up with now
	// in theory, there exists a file somewhere that will be the exact size
	// provided and in that case the total size will just need to be bumped up
	// to allow more larger files, this won't really change that much, but it
	// will prevent a "zip bomb" in the event that one is encountered.

	written, err := io.CopyN(w, r, size)

	if err != io.EOF && err != nil {
		log.Fatalln("Unable to write unzipped file", err)
	} else if written == size {

		log.Println("Needed filesize is larger than the maximum provided filesize")
		return errors.New("filesize too large")

	}

	return nil

}

func covertStringSizeToBytes(sizeStr string) int64 {
	sizeInt, err := strconv.ParseInt(sizeStr, 10, 64)

	// non-conventional, but the meat of the function is only going to
	// happen if there is a suffix for G/M/K to indicate higher order
	// bytes

	if err == nil {
		return sizeInt
	}

	modChar := sizeStr[len(sizeStr)-1:]

	modChar = strings.ToUpper(modChar)

	mod, exists := modCharMap[modChar]

	if !exists {
		log.Fatalln("Unable to use", modChar, "as modifier for bytes")
	}

	sizeInt, err = strconv.ParseInt(sizeStr[:len(sizeStr)-1], 10, 64)

	if err != nil {
		log.Fatalln("Unable to convert", sizeStr, "to int")
	}

	return sizeInt * mod

}
