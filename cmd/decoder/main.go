package main

import (
	"flag"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	core "github.com/genkinomori/jpeg-censor/core"
)

var (
	flagInput   = flag.String("input", "", "Input file. If this argument is left empty, all images in current directory will be proceeded, with mask image and output image named automatically with _r. ")
	flagOutput  = flag.String("output", "", "Mask image. If this argument is left empty, the file with the same file name except a _r suffix would be used as mask image. ")
	flagQuality = flag.Int("quality", 90, "Output JPEG quality")
)

type Task struct {
	Input  string
	Output string
}

func addFilenamePrefix(filename, prefix, forceExt string) string {
	dir, filename := filepath.Split(filename)
	ext := filepath.Ext(filename)
	base := strings.TrimSuffix(filename, ext)
	if forceExt != "" {
		ext = forceExt
	}
	return dir + prefix + base + ext
}

func reportError(err error) {
	fmt.Println("Something happened. ")
	fmt.Println("Please contact the author for more information. ")
	fmt.Println("")
	panic(err)
}

func reportUserError(text string) {
	fmt.Println(text)
	os.Exit(1)
}

func readImage(filename string) image.Image {
	ext := strings.ToLower(filepath.Ext(filename))
	if _, err := os.Stat(filename); err != nil {
		if os.IsNotExist(err) {
			reportUserError(fmt.Sprintf("File `%s` not exist. ", filename))
		} else {
			reportError(err)
		}
	}
	f, err := os.Open(filename)
	if err != nil {
		reportError(err)
	}
	defer f.Close()
	if ext == ".jpg" || ext == ".jpeg" {
		ret, err := jpeg.Decode(f)
		if err != nil {
			reportError(err)
		}
		return ret
	} else if ext == ".png" {
		ret, err := png.Decode(f)
		if err != nil {
			reportError(err)
		}
		return ret
	}
	reportUserError(fmt.Sprintf("Unknown format: %s", ext))
	return nil
}

func main() {
	flag.Parse()

	quality := *flagQuality
	if quality < 0 {
		reportUserError(fmt.Sprintf("Invalid quality: %d", quality))
	}

	tasks := make([]*Task, 0)
	if *flagInput == "" {
		files, err := os.ReadDir(".")
		if err != nil {
			reportError(err)
		}
		for _, file := range files {
			filename := file.Name()
			ext := strings.ToLower(filepath.Ext(filename))
			if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
				continue
			}
			base := strings.TrimSuffix(filename, ext)
			if strings.HasSuffix(base, "_r") {
				continue
			}
			tasks = append(tasks, &Task{
				Input: filename,
			})
		}
	} else {
		tasks = append(tasks, &Task{
			Input:  *flagInput,
			Output: *flagOutput,
		})
	}

	if len(tasks) == 0 {
		reportUserError("No input file. ")
	}

	for _, task := range tasks {
		if task.Output == "" {
			task.Output = addFilenamePrefix(task.Input, "restored_", ".jpg")
		}

		im := readImage(task.Input)
		w, err := os.Create(task.Output)
		if err != nil {
			reportError(err)
		}
		fmt.Printf("%s -> %s\n", task.Input, task.Output)

		cr, err := core.UnlayoutCensoredImage(im)
		if err != nil {
			reportError(err)
		}
		u := core.Uncensor(cr)
		err = jpeg.Encode(w, u, &jpeg.Options{
			Quality: quality,
		})
		if err != nil {
			reportError(err)
		}
	}
}
