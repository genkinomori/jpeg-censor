package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/genkinomori/jpeg-censor/core"
)

var (
	flagInput          = flag.String("input", "", "Input file. If this argument is left empty, all images in current directory will be proceeded, with mask image and output image named automatically with _m or _o. ")
	flagMask           = flag.String("mask", "", "Mask image. If this argument is left empty, the file with the same file name except a _m suffix would be used as mask image. ")
	flagOutput         = flag.String("output", "", "Mask image. If this argument is left empty, the file with the same file name except a _o suffix would be used as mask image. ")
	flagQuality        = flag.Int("quality", 90, "Output JPEG quality")
	flagMaskGridSize   = flag.Int("maskgrid", 1, "Size of maskgrid")
	flagColorDistThres = flag.Int("colordistthres", 50, "Color distance threshold")
)

type Task struct {
	Input  string
	Mask   string
	Output string
}

func addFilenameSuffix(filename, suffix, forceExt string) string {
	ext := filepath.Ext(filename)
	base := strings.TrimSuffix(filename, ext)
	if forceExt != "" {
		ext = forceExt
	}
	return base + suffix + ext
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

func colorDistSquared(a, b color.Color) int {
	ra, ga, ba, _ := a.RGBA()
	rb, gb, bb, _ := b.RGBA()
	dr, dg, db := (int(ra)-int(rb))>>8, (int(ga)-int(gb))>>8, (int(ba)-int(bb))>>8
	return dr*dr + dg*dg + db*db
}

func main() {
	flag.Parse()

	maskgrid := *flagMaskGridSize
	if maskgrid < 0 {
		reportUserError(fmt.Sprintf("Invalid maskgrid: %d", maskgrid))
	}
	colordistThreshold := *flagColorDistThres
	if colordistThreshold < 0 {
		reportUserError(fmt.Sprintf("Invalid colordistthres: %d", colordistThreshold))
	}
	cdsq := colordistThreshold * colordistThreshold
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
			if strings.HasSuffix(base, "_m") || strings.HasSuffix(base, "_o") {
				continue
			}
			tasks = append(tasks, &Task{
				Input: filename,
			})
		}
	} else {
		tasks = append(tasks, &Task{
			Input:  *flagInput,
			Mask:   *flagMask,
			Output: *flagOutput,
		})
	}

	if len(tasks) == 0 {
		reportUserError("No input file. ")
	}

	for _, task := range tasks {
		if task.Mask == "" {
			task.Mask = addFilenameSuffix(task.Input, "_m", "")
		}
		if task.Output == "" {
			task.Output = addFilenameSuffix(task.Input, "_o", ".jpg")
		}

		input := readImage(task.Input)
		mask := readImage(task.Mask)
		bounds := input.Bounds()
		if mask.Bounds() != bounds {
			reportUserError(fmt.Sprintf("Different image size for `%s` and `%s`. ", task.Input, task.Mask))
		}

		maskgrey := image.NewGray(bounds)
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				ci := input.At(x, y)
				cm := mask.At(x, y)
				if colorDistSquared(ci, cm) >= cdsq {
					maskgrey.SetGray(x, y, color.Gray{255})
				} else {
					maskgrey.SetGray(x, y, color.Gray{0})
				}
			}
		}

		w, err := os.Create(task.Output)
		if err != nil {
			reportError(err)
		}
		fmt.Printf("%s -> %s\n", task.Input, task.Output)

		cr := core.Censor(input, maskgrey, maskgrid)
		layouted := core.LayoutCensoredImage(cr)
		err = jpeg.Encode(w, layouted, &jpeg.Options{
			Quality: quality,
		})
		if err != nil {
			reportError(err)
		}
	}
}
