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
	"net/http"
	"regexp"
	"strconv"
)

// A Route handles the business logic of a Halfshell request. It contains a
// Processor and a Source. When a request is serviced, the appropriate route
// is chosen after which the image is retrieved from the source and
// processed by the processor.
type Route struct {
	Name           string
	Pattern        *regexp.Regexp
	ImagePathIndex int
	Processor      ImageProcessor
	Source         ImageSource
	Statter        Statter
}

// NewRouteWithConfig returns a pointer to a new Route instance created using
// the provided configuration settings.
func NewRouteWithConfig(config *RouteConfig, statterConfig *StatterConfig) *Route {
	return &Route{
		Name:           config.Name,
		Pattern:        config.Pattern,
		ImagePathIndex: config.ImagePathIndex,
		Processor:      NewImageProcessorWithConfig(config.ProcessorConfig),
		Source:         NewImageSourceWithConfig(config.SourceConfig),
		Statter:        NewStatterWithConfig(config, statterConfig),
	}
}

// ShouldHandleRequest accepts an HTTP request and returns a bool indicating
// whether the route should handle the request.
func (p *Route) ShouldHandleRequest(r *http.Request) bool {
	return p.Pattern.MatchString(r.URL.Path)
}

// SourceAndProcessorOptionsForRequest parses the source and processor options
// from the request.
func (p *Route) SourceAndProcessorOptionsForRequest(r *http.Request) (
	*ImageSourceOptions, *ImageProcessorOptions) {

	matches := p.Pattern.FindAllStringSubmatch(r.URL.Path, -1)[0]
	path := matches[p.ImagePathIndex]

	width, _ := strconv.ParseUint(r.FormValue("w"), 10, 32)
	height, _ := strconv.ParseUint(r.FormValue("h"), 10, 32)
	blurRadius, _ := strconv.ParseFloat(r.FormValue("blur"), 64)
	focalpoint := r.FormValue("focalpoint")

	scaleModeName := r.FormValue("scale_mode")
	scaleMode, _ := ScaleModes[scaleModeName]

	return &ImageSourceOptions{Path: path}, &ImageProcessorOptions{
		Dimensions: ImageDimensions{uint(width), uint(height)},
		BlurRadius: blurRadius,
		ScaleMode:  uint(scaleMode),
		Focalpoint: NewFocalpointFromString(focalpoint),
	}
}
