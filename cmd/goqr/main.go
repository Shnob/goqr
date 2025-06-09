package main

import (
	"fmt"
	"image/png"
	"os"

	"github.com/Shnob/goqr/pkg/qr"
)

func main() {
	qr, err := qr.NewQr(2)
	if err != nil {
		fmt.Printf("Error creating QR: %s\n", err.Error())
		return
	}

	img := qr.GenerateDebugImage()

	file, err := os.Create("qr.png")
	if err != nil {
		fmt.Printf("Error creating file: %s\n", err.Error())
		return
	}

	png.Encode(file, img)
}
