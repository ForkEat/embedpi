package eink

import (
	"fmt"
	"image"
	"os"

	"github.com/fogleman/gg"
	"github.com/otaviokr/go-epaper-lib"
)

func NewEinkDevice() (*epaper.EPaper, error) {
	// Create new EPaperDisplay handler.
	epd, err := epaper.New(epaper.Model2in7bw)
	if err != nil {
		return nil, err
	}

	epd.Init()
	//epd.ClearScreen()
	return epd, err
}

func DisplayText(text string, epd *epaper.EPaper) {
	context := gg.NewContext(epaper.Model2in7bw.Width, epaper.Model2in7bw.Height)
	context.SetRGB(1, 1, 1)
	context.Clear()
	context.SetRGB(0, 0, 0)
	context.DrawStringWrapped(text, 0, 0, 0, 0, float64(epaper.Model2in7bw.Width), 1.4, gg.AlignLeft)

	printImage(epd, context.Image())
}

func printImage(epd *epaper.EPaper, img image.Image) {
	epd.AddLayer(img, 0, 0, false)
	epd.PrintDisplay()
}

func printImageRotated(epd *epaper.EPaper, imageFile string) {
	reader, err := os.Open(imageFile)
	if err != nil {
		fmt.Printf("ERROR while loading image: %+v\n", err)
	}
	defer reader.Close()

	m, _, err := image.Decode(reader)
	if err != nil {
		fmt.Printf("ERROR while decoding image: %+v\n", err)
	}

	r := epd.Rotate(m)

	epd.AddLayer(r, 0, 0, false)
	epd.PrintDisplay()
}

func printImageRotatedPosition(epd *epaper.EPaper, imageFile string, x, y int) {
	reader, err := os.Open(imageFile)
	if err != nil {
		fmt.Printf("ERROR while loading image: %+v\n", err)
	}
	defer reader.Close()

	m, _, err := image.Decode(reader)
	if err != nil {
		fmt.Printf("ERROR while decoding image: %+v\n", err)
	}

	r := epd.Rotate(m)

	epd.AddLayer(r, x, y, false)
	epd.PrintDisplay()
}
