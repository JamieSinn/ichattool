package main

import (
	"../.."
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var (
	targetFile = flag.String("f", "", "Single extraction: Target plist/ichat file to decode.")
	targetDir  = flag.String("d", "", "Bulk extraction: Target directory to extract plist/ichat files from.")
	outputDir  = flag.String("o", "", "Output directory to put extracted images in.")
	imagesOnly = flag.Bool("i", false, "Enables image only extraction")
	//verbose = flag.Bool("v", false, "Enable verbose logging output.")
)

func main() {
	flag.Parse()

	if *targetFile != "" && *targetDir != "" {
		// error - cannot specify both
		fmt.Println("Target file and target Dir are specified")
		return
	}
	if *targetFile != "" {
		if *imagesOnly {
			images := iChatTool.ExtractImages(*targetFile)
			writeImages(images)
		}
	} else if *targetDir != "" {
		files := getFiles(*targetDir)
		for _, file := range files {
			if *imagesOnly {
				images := iChatTool.ExtractImages(file)
				writeImages(images)
			} else {
				iChatTool.ExtractData(file)
			}
		}
	}

}

func getFiles(path string) (files []string) {
	fmt.Println(path)
	err := filepath.Walk(path,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if strings.HasSuffix(info.Name(), ".ichat") || strings.HasSuffix(info.Name(), ".plist") {
				files = append(files, path)
			}
			return nil
		})
	if err != nil {

	}
	return
}

func writeImages(images []iChatTool.AttachedImage) {
	for _, img := range images {
		size := len(img.ImageBytes)

		f, err := os.Create(filepath.Join(*outputDir, *targetFile+"_"+strconv.Itoa(size)+"."+img.ImageType))
		if err != nil {
			break
		}
		w := bufio.NewWriter(f)
		n, err := w.Write(img.ImageBytes)
		fmt.Printf("\twrote %d bytes\n", n)
		if err != nil {

			fmt.Println("\tInvalid image. Byte data:")
			fmt.Println(img.ImageBytes[:10])
			fmt.Println(img.ImageBytes[size-10:])
			fmt.Println(err)

			break
		}
	}
}
