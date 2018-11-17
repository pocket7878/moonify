package main // import "github.com/pocket7878/moonify"

import (
	"bufio"
	"fmt"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"math"
	"os"
	"strconv"
)

const (
	debug = true
)

func usage() {
	fmt.Fprintf(os.Stderr, "%s <image_file> <width> <height>", os.Args[0])
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func ceil(n, m int) int {
	res := n / m * m
	if n%m != 0 {
		res += m
	}
	return res
}

func main() {
	if len(os.Args) != 4 {
		usage()
		return
	}

	img, err := loadImage(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	w, err := strconv.Atoi(os.Args[2])
	if err != nil {
		log.Fatal(err)
	}
	h, err := strconv.Atoi(os.Args[3])
	if err != nil {
		log.Fatal(err)
	}

	b := img.Bounds()
	dx := ceil(b.Dx()/w, 4)
	fw := b.Dx() / w * w
	dy := ceil(b.Dy()/h, 4)
	fh := b.Dy() / h * h

	binImg := binaryImg(grayScaleImg(img))

	for y := b.Min.Y; y < fh; y += dy {
		for x := b.Min.X; x < fw; x += dx {
			fmt.Print(calcMoon(binImg, x, y, x+dx-1, y+dy-1))
		}
		fmt.Println()
	}
}

func allOf(ary []int, p func(int) bool) bool {
	for _, i := range ary {
		if !p(i) {
			return false
		}
	}
	return true
}

func allOne(ary []int) bool {
	return allOf(ary, func(i int) bool { return i == 1 })
}

func allZero(ary []int) bool {
	return allOf(ary, func(i int) bool { return i == 0 })
}

func calcMoon(img *image.Gray, x0, y0, x1, y1 int) string {
	dx := (x1 - x0 + 1) / 4
	if dx < 1 {
		dx = 1
	}

	result := make([]int, 0)
	for x := x0; x <= x1; x += dx {
		result = append(result, lightOrDark(img, x, y0, x+dx-1, y1))
	}

	if len(result) != 4 {
		fmt.Fprintln(os.Stderr, x0, " ~ ", x1, " dx:", dx)
		return string(rune(0x1F311))
	}

	if allZero(result) {
		return string(rune(0x1F311))
	}
	if allZero(result[0:2]) && result[3] == 1 {
		return string(rune(0x1F312))
	}
	if result[0] == 0 && result[1] == 0 && allOne(result[2:]) {
		return string(rune(0x1F313))
	}
	if result[0] == 0 && allOne(result[1:]) {
		return string(rune(0x1F314))
	}
	if allOne(result) {
		return string(rune(0x1F315))
	}
	if allOne(result[0:2]) && result[3] == 0 {
		return string(rune(0x1F316))
	}
	if result[0] == 1 && result[1] == 1 && allZero(result[2:]) {
		return string(rune(0x1F317))
	}
	if result[0] == 1 && allZero(result[1:]) {
		return string(rune(0x1F318))
	}
	if result[0] == 1 && result[1] == 0 && result[2] == 1 && result[3] == 0 {
		return string(rune(0x1F317))
	}
	if result[0] == 1 && result[1] == 0 && result[2] == 1 && result[3] == 1 {
		return string(rune(0x1F313))
	}
	if result[0] == 0 && result[1] == 1 && result[2] == 0 && result[3] == 1 {
		return string(rune(0x1F313))
	}
	if result[0] == 1 && result[1] == 1 && result[2] == 0 && result[3] == 1 {
		return string(rune(0x1F317))
	}

	//一旦新月にする
	return string(rune(0x1F311))
}

func lightOrDark(img *image.Gray, x0, y0, x1, y1 int) int {
	size := (x1 - x0 + 1) * (y1 - y0 + 1)
	lightSum := 0
	for x := x0; x <= x1; x++ {
		for y := y0; y <= y1; y++ {
			lightSum += pixelBinary(img, x, y)
		}
	}
	if lightSum > size/2 {
		return 1
	}
	return 0
}

func pixelBinary(img *image.Gray, x, y int) int {
	c := img.At(x, y)
	r, _, _, _ := c.RGBA()
	if r == 0 {
		return 0
	}
	return 1
}

func loadImage(filepath string) (image.Image, error) {
	f, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	in := bufio.NewReader(f)
	_, format, err := image.DecodeConfig(in)
	if err != nil {
		return nil, err
	}
	_, err = f.Seek(0, io.SeekStart)
	if err != nil {
		return nil, err
	}
	in.Reset(f)

	var img image.Image
	switch format {
	case "png":
		img, err = png.Decode(in)
	case "gif":
		img, err = gif.Decode(in)
	case "jpeg":
		img, err = jpeg.Decode(in)
	default:
		return nil, fmt.Errorf("Unsupported image format: %s", format)
	}
	if err != nil {
		return nil, err
	}

	return img, nil
}

//画像をグレースケール化
func grayScaleImg(img image.Image) *image.Gray {
	b := img.Bounds()
	grayImg := image.NewGray(b)
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			c := img.At(x, y)
			gray := color.GrayModel.Convert(c)
			grayImg.Set(x, y, gray)
		}
	}
	return grayImg
}

//画像を二値化
func binaryImg(img *image.Gray) *image.Gray {
	//画素値のヒストグラムを計算
	hist := make(map[uint8]int)
	b := img.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			c := img.At(x, y)
			col := c.(color.Gray)
			hist[col.Y]++
		}
	}

	//最大のクラス分離を計算する
	maxTh := 0
	maxS := -10.0
	for th := 0; th < 256; th++ {
		n1 := 0
		for i := 0; i < th; i++ {
			n1 += hist[uint8(i)]
		}
		n2 := 0
		for i := th; i < 256; i++ {
			n2 += hist[uint8(i)]
		}

		var mu1 int
		if n1 == 0 {
			mu1 = 0
		} else {
			for i := 0; i < th; i++ {
				mu1 += i * hist[uint8(i)] / n1
			}
		}

		var mu2 int
		if n2 == 0 {
			mu2 = 0
		} else {
			for i := th; i < 256; i++ {
				mu2 += i * hist[uint8(i)] / n2
			}
		}

		s := math.Pow(float64(n1*n2*(mu1-mu2)), 2.0)
		if s > maxS {
			maxTh = th
			maxS = s
		}
	}

	bImg := image.NewGray(b)
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			c := img.At(x, y)
			col := c.(color.Gray)
			if col.Y >= uint8(maxTh) {
				bImg.Set(x, y, color.RGBA{
					255, 255, 255, 255,
				})
			} else {
				bImg.Set(x, y, color.RGBA{
					0, 0, 0, 255,
				})
			}
		}
	}
	return bImg
}

func writeImage(img image.Image, filepath string) error {
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	png.Encode(out, img)
	return nil
}
