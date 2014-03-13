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
	"io/ioutil"
	"net/http"
)

// Image contains a byte array of the image data and its MIME type.
// TODO: See if we can use the std library's Image type without incurring
// the hit of extra copying.
type Image struct {
	Bytes    []byte
	MimeType string
}

// Returns a pointer to a new Image created from an HTTP response object.
func NewImageFromHTTPResponse(httpResponse *http.Response) (*Image, error) {
	imageBytes, err := ioutil.ReadAll(httpResponse.Body)
	defer httpResponse.Body.Close()
	if err != nil {
		return nil, err
	}

	return &Image{
		Bytes:    imageBytes,
		MimeType: httpResponse.Header.Get("Content-Type"),
	}, nil
}

// Width and height of an image.
type ImageDimensions struct {
	Width  uint64
	Height uint64
}

// Returns the image dimension's aspect ratio.
func (d ImageDimensions) AspectRatio() float64 {
	return float64(d.Width) / float64(d.Height)
}

func (d ImageDimensions) String() string {
	return fmt.Sprintf("%dx%d", d.Width, d.Height)
}
