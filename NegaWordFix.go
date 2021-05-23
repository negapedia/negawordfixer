package main

import (
	"flag"
	"fmt"
	"github.com/marcochilese/Go-Trie"
	"github.com/marcochilese/negawordfixer/src/fsutils"
	"github.com/marcochilese/negawordfixer/src/processing"
	"io/ioutil"
	"os"
	"path"
	"time"
)

func buildTrieAndReplacementDict(pathToDict string) (*trie.Trie, *map[string]string) {
	mytrie := trie.BuildTrieFromDictionary(pathToDict)

	// in replacementDict are stored all the replacements in order to
	// speedup the replacement process when stored enough history
	replacementDict := make(map[string]string)
	return mytrie, &replacementDict
}

func getNewestFileInDir(dir string) string {
	if dir[len(dir)-1] != "/"[0] {
		dir += "/"
	}

	files, _ := ioutil.ReadDir(dir)
	var newestFile string
	var newestTime int64 = 0
	for _, f := range files {
		if f.Name() == ".DS_Store" {
			continue
		}
		fi, err := os.Stat(dir + f.Name())
		if err != nil {
			fmt.Println(err)
		}
		currTime := fi.ModTime().Unix()
		if currTime > newestTime {
			newestTime = currTime
			newestFile = f.Name()
		}
	}
	return path.Join(dir, newestFile)
}

func main() {
	tarPathPtr := flag.String("tar", "", "Path to negapedia-LANG.tar.gz")
	langPtr := flag.String("lang", "en", "Negapedia language")
	verbosePtr := flag.Bool("verbose", false, "Negapedia language")
	flag.Parse()

	*tarPathPtr = getNewestFileInDir(*tarPathPtr)
	pathToDict := path.Join("./dictionary_data/", *langPtr+".txt")

	logger := ioutil.Discard
	if *verbosePtr {
		logger = os.Stdout
	}

	fmt.Println("--- NegaWordsFixer ---")
	fmt.Println("Run with config:\n\tLang: " + *langPtr +
		"\n\tDict: " + pathToDict +
		"\n\tTar: " + *tarPathPtr + "\n\t")

	mytrie, replacementDict := buildTrieAndReplacementDict(pathToDict)

	fmt.Println("Extraction start")
	start := time.Now()
	tmpDir, err := fsutils.ExtractTarGz(*tarPathPtr)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Extraction done in ", time.Now().Sub(start))

	filesToProcess := fsutils.GetFilesList(tmpDir, false)

	fmt.Println("To process: ", len(filesToProcess))
	fmt.Println("Processing start")
	start = time.Now()
	for _, file := range filesToProcess {
		err := processing.ProcessPage(file, *mytrie, replacementDict, &logger)
		if err != nil {
			os.RemoveAll(tmpDir)
		}

	}
	fmt.Println("Processing done in ", time.Now().Sub(start))

	fmt.Fprintln(logger, "Compression start")
	start = time.Now()
	err = fsutils.CompressTarGz(tmpDir, path.Join(*tarPathPtr))
	if err != nil {
		fmt.Println(err)
	}
	os.RemoveAll(tmpDir)
	fmt.Println("Compression done in ", time.Now().Sub(start))
	fmt.Println("Done.")
}