package main

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	"image/png"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

// Can be used to split a image sliced by SpriteDecomposer
type Sprite struct {
	Image            string    `xml:"image,attr"`
	TransparentColor string    `xml:"transparentColor,attr"`
	Animation        Animation `xml:"animation"`
}

func loopOverBounds(r image.Rectangle, operator func(x, y int)) {
	for x := r.Min.X; x < r.Max.X; x++ {
		for y := r.Min.Y; y < r.Max.Y; y++ {
			operator(x, y)
		}
	}
}

func strToColor(str string) color.Color {
	ret := color.RGBA{R: 0, G: 0, B: 0, A: 0}
	fmt.Sscanf(str, "%x", &ret.R)
	fmt.Sscanf(str, "%x", &ret.G)
	fmt.Sscanf(str, "%x", &ret.B)
	return ret
}

func (s *Sprite) Split(img image.Image, outputDir string, prefix string) error {
	return s.Animation.SplitTo(img, outputDir, prefix)
}

type Animation struct {
	Cuts []Cut `xml:"cut"`
}

func (a *Animation) SplitTo(img image.Image, out string, prefix string) error {
	parts := len(a.Cuts)

	padCount := len(fmt.Sprintf("%d", parts))
	nameFmt := fmt.Sprintf("-%%0%dd.png", padCount)

	var err error
	for i, _ := range a.Cuts {
		err = a.Cuts[i].SplitTo(img, filepath.Join(out, prefix+fmt.Sprintf(nameFmt, i)))
		if err != nil {
			return err
		}
	}
	return nil
}

type Cut struct {
	W   int `xml:"w,attr"`
	H   int `xml:"h,attr"`
	X   int `xml:"x,attr"`
	Y   int `xml:"y,attr"`
	Row int `xml:"row,attr"`
	Col int `xml:"col,attr"`
}

type subImager interface {
	SubImage(r image.Rectangle) image.Image
}

func (c *Cut) SplitTo(img image.Image, outfile string) error {
	if img, ok := img.(subImager); ok {
		sub := img.SubImage(c.AsRect())
		file, err := os.Create(outfile)
		if err != nil {
			return err
		}
		defer file.Close()
		return png.Encode(file, sub)
	} else {
		return fmt.Errorf("Cannot take a SubImage")
	}
}

func (c *Cut) AsRect() image.Rectangle {
	return image.Rectangle{
		Min: image.Point{X: c.X, Y: c.Y},
		Max: image.Point{X: c.X + c.W, Y: c.Y + c.H},
	}
}

func splitIt(xmlFile string) {
	buf, err := ioutil.ReadFile(xmlFile)
	if err != nil {
		log.Printf("Error reading file. %v", err)
		return
	}
	sprite := &Sprite{}
	err = xml.Unmarshal(buf, sprite)
	log.Printf("Sprite file: %v", sprite)
	if err != nil {
		log.Printf("Unmarshal error: %v", err)
		return
	}

	dir, _ := filepath.Split(xmlFile)

	srcImage := filepath.Join(dir, sprite.Image)

	buf, err = ioutil.ReadFile(srcImage)

	if err != nil {
		log.Printf("Error reading image file. Cause: %v", err)
		return
	}

	fullImage, _, err := image.Decode(bytes.NewBuffer(buf))
	if err != nil {
		log.Printf("Error decoding image. Cause: %v", err)
		return
	}
	err = sprite.Split(fullImage, dir, "part")
	if err != nil {
		log.Printf("Error spliting the image. Cause: %v", err)
	}
}
