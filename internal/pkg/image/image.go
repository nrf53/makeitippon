package image

import (
	"bufio"
	"image"
	"image/draw"
	"image/png"
	"io/ioutil"
	"log"
	"math"
	"os"
	"unicode/utf8"

	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

const (
	colorCode = "FFE100"
	baseimg   = "./templates/images/makeitippon.png"
	fontfile  = "./templates/fonts/GenEiGothicP-H-KL.otf"
	outimg    = "./templates/images/outimage.png"
)

func Text2img(twettText string) {
	// Open base image
	file, err := os.Open(baseimg)
	if err != nil {
		log.Println("[ERROR] Failed to open base image")
		log.Fatal(err)
	}

	defer file.Close()

	// Read Font File
	ftBin, err := ioutil.ReadFile(fontfile)
	if err != nil {
		log.Println("[ERROR] Failed to Read Font File")
		log.Fatal(err)
	}

	// Parse Font File to usable type
	ft, err := opentype.Parse(ftBin)
	if err != nil {
		log.Println("[ERROR] Failed to Parse Font File to usable type")
		log.Fatal(err)
	}

	// Decode base image
	img, err := png.Decode(file)
	if err != nil {
		log.Println("[ERROR] Failed to Decode base image")
		log.Fatal(err)
	}

	// Prepare image to be drawn
	dst := image.NewRGBA(img.Bounds())
	//fmt.Println(dst.Rect)
	draw.Draw(dst, dst.Bounds(), img, image.ZP, draw.Src)

	// Text
	//text := "うわああ"

	// Set Font option parms
	opt := opentype.FaceOptions{
		Size:    66,
		DPI:     72,
		Hinting: font.HintingNone,
	}

	// New font
	face, err := opentype.NewFace(ft, &opt)
	if err != nil {
		log.Println("[ERROR] Failed to New face")
		log.Fatal(err)
	}

	// Define position to draw
	//	x, y := 0, 0

	//	dot := fixed.Point26_6{X: fixed.Int26_6(x * 128), Y: fixed.Int26_6(y * 72)}
	//	dot := fixed.Point26_6{X: fixed.Int26_6(x * 64), Y: fixed.Int26_6(y * 26)}

	// Define how to draw
	d := &font.Drawer{
		Dst:  dst,
		Src:  image.Black,
		Face: face,
		Dot:  fixed.Point26_6{},
	}
	y := 20 + int(math.Ceil(opt.Size*opt.DPI/72))
	dy := int(math.Ceil(opt.Size * 1 * opt.DPI / 72))
	d.Dot = fixed.Point26_6{
		X: (fixed.I(1280) - d.MeasureString(twettText)) / 2,
		Y: fixed.I(y),
	}
	strlen := float64(d.MeasureString(twettText).Floor())
	imgw := float64(fixed.I(1280).Ceil())
	lines := math.Ceil(strlen / imgw)
	splitlen := (utf8.RuneCountInString(twettText) / int(lines))

	// for debug
	/*
		fmt.Println(d.MeasureString(twettText).Floor())
		fmt.Println(lines)
		fmt.Println(utf8.RuneCountInString(twettText))
		fmt.Println(splitlen)
	*/

	// tweet text to []string
	runes := []rune(twettText)
	var strary []string
	for i := 0; i < len(runes); i += splitlen {
		if i+splitlen < len(runes) {
			strary = append(strary, string(runes[i:(i+splitlen)]))
		} else {
			strary = append(strary, string(runes[i:]))
		}
	}

	//	fmt.Println(strary)

	for _, s := range strary {
		d.Dot = fixed.P(10, y)
		d.DrawString(s)
		y += dy
	}
	//	y += dy

	// Output and Finalize
	newFile, err := os.Create(outimg)
	if err != nil {
		log.Println("[ERROR] Failed to Output")
		log.Fatal(err)
	}

	defer newFile.Close()

	b := bufio.NewWriter(newFile)
	if err := png.Encode(b, dst); err != nil {
		log.Println("[ERROR] Failed to encode")
		log.Fatal(err)
	}
	err = b.Flush()
	if err != nil {
		log.Println("[ERROR] Failed to flush")
		log.Println(err)
		os.Exit(1)
	}

}
