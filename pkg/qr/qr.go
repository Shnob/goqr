package qr

import (
	"fmt"
	"image"
	"image/color"
)

type Qr struct {
	QrType         QrType
	encodingRegion EncodingRegion
}

func NewQr(qrType QrType) (qr *Qr, err error) {
	if qrType < 1 || qrType > 44 {
		err = QrError{msg: "qrType must be between 1 and 44"}
		return
	}

	qr = &Qr{
		QrType:         qrType,
		encodingRegion: GenerateEncodingRegion(qrType),
	}

	return
}

func (q *Qr) GenerateBlankImage() (img *image.Gray) {
	wid := int(q.QrType.Width())
	img = image.NewGray(image.Rect(0, 0, wid, wid))

	// Make image white
	for x := range wid {
		for y := range wid {
			img.SetGray(x, y, color.Gray{255})
		}
	}

	// Add timing patterns
	timingOffset := 6
	if q.QrType.IsMicro() {
		timingOffset = 0
	}

	for i := range wid {
		if i%2 == 1 {
			continue
		}

		img.SetGray(i, timingOffset, color.Gray{0})
		img.SetGray(timingOffset, i, color.Gray{0})
	}

	// Add the top left finder pattern
	imageAddFinderPattern(img, 0, 0)

	// Optionally add the top right and bottom left finder patterns
	if !q.QrType.IsMicro() {
		imageAddFinderPattern(img, wid-7, 0)
		imageAddFinderPattern(img, 0, wid-7)
	}

	// Add alignment patterns
	alignmentPositions := alignmentPositions[q.QrType]

	for i, x := range alignmentPositions {
		for j, y := range alignmentPositions {
			if i == 0 && j == 0 ||
				i == 0 && j == len(alignmentPositions)-1 ||
				j == 0 && i == len(alignmentPositions)-1 {
				continue
			}

			imageAddAlignmentPattern(img, x, y)
		}
	}

	return
}

func (q *Qr) GenerateDebugImage() (img *image.Gray) {
	img = q.GenerateBlankImage()

	for i, block := range q.encodingRegion {
		// Color based on which block the module is in
		col := color.Gray{uint8(i)%8*16 + 64}

		for _, module := range block {
			img.Set(int(module.X), int(module.Y), col)
		}
	}

	return
}

func imageAddFinderPattern(img *image.Gray, left, top int) {
	for i := range 7 {
		for j := range 7 {
			if (i == 1 || i == 5) && (j != 0 && j != 6) ||
				(j == 1 || j == 5) && (i != 0 && i != 6) {
				continue
			}

			img.SetGray(i+left, j+top, color.Gray{0})
		}
	}
}

func imageAddAlignmentPattern(img *image.Gray, x, y int) {
	for i := range 5 {
		for j := range 5 {
			if (i == 1 || i == 3) && (j != 0 && j != 4) ||
				(j == 1 || j == 3) && (i != 0 && i != 4) {
				continue
			}

			img.SetGray(i+x-2, j+y-2, color.Gray{0})
		}
	}
}

// QrType 1-40 -> version 1-40
// QrType 41-44 -> micro 1-4
type QrType uint8

func (t QrType) Width() int {
	if t <= 40 {
		return 4*int(t) + 17
	} else {
		return 2*int(t-40) + 9
	}
}

func (t QrType) IsMicro() bool {
	return t > 40
}

type qrCoord struct {
	X uint8
	Y uint8
}

type EncodingRegion [][]qrCoord

// Generates the layout of blocks in the encoding region, given a QR type
func GenerateEncodingRegion(qrType QrType) EncodingRegion {
	er := make([][]qrCoord, 0)

	wid := uint8(qrType.Width())
	// The X coordinate of the vertical timing pattern
	timingX := 5
	if qrType > 40 {
		timingX = 0
	}

	isGoingUp := true
	currPos := qrCoord{wid - 1, wid - 1}

	nextPos := func() {
		if (currPos.X%2 == 0 && currPos.X > uint8(timingX)) ||
			(currPos.X%2 == 1 && currPos.X < uint8(timingX)) ||
			currPos.X == uint8(timingX) {
			currPos.X -= 1
		} else {
			currPos.X += 1
			if isGoingUp {
				currPos.Y -= 1
			} else {
				currPos.Y += 1
			}
		}

		if currPos.Y >= wid {
			currPos.X -= 2
			if isGoingUp {
				currPos.Y += 1
				isGoingUp = false
			} else {
				currPos.Y -= 1
				isGoingUp = true
			}
		}
	}

	currBlock := make([]qrCoord, 0, 8)

	addPointToBlock := func() {
		currBlock = append(currBlock, currPos)

		if len(currBlock) == 8 {
			// Push the block to the list of blocks
			er = append(er, currBlock)
			// Make a new block
			currBlock = make([]qrCoord, 0, 8)
		}
	}

	for currPos.X < wid && currPos.Y < wid {
		fmt.Printf("Point %d,%d\n", currPos.X, currPos.Y)
		addPointToBlock()
		nextPos()
	}

	return er
}

type QrError struct {
	msg string
}

func (e QrError) Error() string {
	return e.msg
}

var alignmentPositions = map[QrType][]int{
	41: {},
	42: {},
	43: {},
	44: {},
	1:  {},
	2:  {6, 18},
	3:  {6, 22},
	4:  {6, 26},
	5:  {6, 30},
	6:  {6, 34},
	7:  {6, 22, 38},
	8:  {6, 24, 42},
	9:  {6, 26, 46},
	10: {6, 28, 50},
	11: {6, 30, 54},
	12: {6, 32, 58},
	13: {6, 34, 62},
	14: {6, 26, 46, 66},
	15: {6, 26, 48, 70},
	16: {6, 26, 50, 74},
	17: {6, 30, 54, 78},
	18: {6, 30, 56, 82},
	19: {6, 30, 58, 86},
	20: {6, 34, 62, 90},
	21: {6, 28, 50, 72, 94},
	22: {6, 26, 50, 74, 98},
	23: {6, 30, 54, 78, 102},
	24: {6, 28, 54, 80, 106},
	25: {6, 32, 58, 84, 110},
	26: {6, 30, 58, 86, 114},
	27: {6, 34, 62, 90, 118},
	28: {6, 26, 50, 74, 98, 122},
	29: {6, 30, 54, 78, 102, 126},
	30: {6, 26, 52, 78, 104, 130},
	31: {6, 30, 56, 82, 108, 134},
	32: {6, 34, 60, 86, 112, 138},
	33: {6, 30, 58, 86, 114, 142},
	34: {6, 34, 62, 90, 118, 146},
	35: {6, 30, 54, 78, 102, 126, 150},
	36: {6, 24, 50, 76, 102, 128, 154},
	37: {6, 28, 54, 80, 106, 132, 158},
	38: {6, 32, 58, 84, 110, 136, 162},
	39: {6, 26, 54, 82, 110, 138, 166},
	40: {6, 30, 58, 86, 114, 142, 170},
}
