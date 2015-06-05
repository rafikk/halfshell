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
	"net/url"
	"strings"
)

const (
	ImageSourceTypeHttp ImageSourceType = "http"
)

type HttpImageSource struct {
	Config *SourceConfig
	Logger *Logger
}

func NewHttpImageSourceWithConfig(config *SourceConfig) ImageSource {
	return &HttpImageSource{
		Config: config,
		Logger: NewLogger("source.http.%s", config.Name),
	}
}

func (s *HttpImageSource) GetImage(request *ImageSourceOptions) (*Image, error) {
	httpRequest := s.getHttpRequest(request)
	httpResponse, err := http.DefaultClient.Do(httpRequest)
	defer httpResponse.Body.Close()
	if err != nil {
		s.Logger.Warnf("Error downlading image: %v", err)
		return nil, err
	}
	if httpResponse.StatusCode != 200 {
		return nil, fmt.Errorf("Error downlading image (url=%v)", httpRequest.URL)
	}
	image, err := NewImageFromBuffer(httpResponse.Body)
	if err != nil {
		responseBody, _ := ioutil.ReadAll(httpResponse.Body)
		s.Logger.Warnf("Unable to create image from response body: %v (url=%v)", string(responseBody), httpRequest.URL)
		return nil, err
	}
	s.Logger.Infof("Successfully retrieved image from http: %v", httpRequest.URL)
	return image, nil
}

func (s *HttpImageSource) getHttpRequest(request *ImageSourceOptions) *http.Request {
	path := s.Config.Directory + request.Path
	imageURLPathComponents := strings.Split(path, "/")

	for index, component := range imageURLPathComponents {
		component = url.QueryEscape(component)
		imageURLPathComponents[index] = component
	}
	requestURL := &url.URL{
		Opaque: strings.Join(imageURLPathComponents, "/"),
		Scheme: "http",
		Host:   s.Config.Host,
	}

	httpRequest, _ := http.NewRequest("GET", requestURL.RequestURI(), nil)
	httpRequest.URL = requestURL

	return httpRequest
}

func init() {
	RegisterSource(ImageSourceTypeHttp, NewHttpImageSourceWithConfig)
}
