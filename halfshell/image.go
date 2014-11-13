// Copyright (c) 2014 Oyster
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package halfshell

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/rafikk/imagick/imagick"
)

var EmptyImageDimensions = ImageDimensions{}
var EmptyResizeDimensions = ResizeDimensions{}

type Image struct {
	Wand      *imagick.MagickWand
	Signature string
	destroyed bool
}

func NewImageFromBuffer(buffer io.Reader) (image *Image, err error) {
	bytes, err := ioutil.ReadAll(buffer)
	if err != nil {
		return nil, err
	}

	image = &Image{Wand: imagick.NewMagickWand()}
	err = image.Wand.ReadImageBlob(bytes)
	if err != nil {
		return nil, err
	}

	return image, nil
}

func NewImageFromFile(file *os.File) (image *Image, err error) {
	image, err = NewImageFromBuffer(file)
	return image, err
}

func (i *Image) GetMIMEType() string {
	return fmt.Sprintf("image/%s", strings.ToLower(i.Wand.GetImageFormat()))
}

func (i *Image) GetBytes() (bytes []byte, size int) {
	bytes = i.Wand.GetImageBlob()
	size = len(bytes)
	return bytes, size
}

func (i *Image) GetWidth() uint {
	return i.Wand.GetImageWidth()
}

func (i *Image) GetHeight() uint {
	return i.Wand.GetImageHeight()
}

func (i *Image) GetDimensions() ImageDimensions {
	return ImageDimensions{i.GetWidth(), i.GetHeight()}
}

func (i *Image) GetSignature() string {
	return i.Wand.GetImageSignature()
}

func (i *Image) Destroy() {
	if !i.destroyed {
		i.Wand.Destroy()
		i.destroyed = true
	}
}

type ImageDimensions struct {
	Width  uint
	Height uint
}

func (d ImageDimensions) AspectRatio() float64 {
	return float64(d.Width) / float64(d.Height)
}

func (d ImageDimensions) String() string {
	return fmt.Sprintf("%dx%d", d.Width, d.Height)
}

type ResizeDimensions struct {
	Scale ImageDimensions
	Crop  ImageDimensions
}

// Focalpoint is a float pair representing the location of the image subject.
// (0.5, 0.5) is the middle. (1, 1) is the bottom right. (0, 0) is the top left.
type Focalpoint struct {
	X float64
	Y float64
}

// NewFocalpointFromString splits the given string into a Focalpoint struct. The
// string format should be: "X,Y". For example: "0.1,0.1".
func NewFocalpointFromString(s string) (fp Focalpoint) {
	pair := strings.Split(s, ",")
	if len(pair) != 2 {
		return Focalpoint{0.5, 0.5}
	}

	x, _ := strconv.ParseFloat(pair[0], 64)
	y, _ := strconv.ParseFloat(pair[1], 64)
	return Focalpoint{x, y}
}
