package barcode

import (
	"fmt"

	"github.com/bieber/barcode"
	"gocv.io/x/gocv"
)

// RunScan scan and decode barcode from camera
func RunScan() {
	webcam, err := gocv.VideoCaptureDevice(0)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer webcam.Close()

	img := gocv.NewMat()
	defer img.Close()

	for {
		if ok := webcam.Read(&img); !ok {
			fmt.Printf("cannot read device %d\n", 0)
			return
		}
		if img.Empty() {
			continue
		}
		webcam.Read(&img)

		scanner := barcode.NewScanner().SetEnabledAll(true)

		imgObj, _ := img.ToImage()

		src := barcode.NewImage(imgObj)
		symbols, _ := scanner.ScanImage(src)

		for _, s := range symbols {
			data := s.Data
			fmt.Println(data)
		}
	}
}
