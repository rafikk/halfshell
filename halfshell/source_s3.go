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
	"github.com/oysterbooks/s3"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	IMAGE_SOURCE_TYPE_S3 ImageSourceType = "s3"
)

type S3ImageSource struct {
	Config *SourceConfig
	Logger *Logger
}

func NewS3ImageSourceWithConfig(config *SourceConfig) ImageSource {
	return &S3ImageSource{
		Config: config,
		Logger: NewLogger("source.s3.%s", config.Name),
	}
}

func (s *S3ImageSource) GetImage(request *ImageSourceOptions) *Image {
	httpRequest := s.signedHTTPRequestForRequest(request)
	httpResponse, err := http.DefaultClient.Do(httpRequest)
	if err != nil {
		s.Logger.Warn("Error downlading image: %v", err)
		return nil
	}
	if httpResponse.StatusCode != 200 {
		s.Logger.Warn("Error downlading image (url=%v)", httpRequest.URL)
		return nil
	}
	image, err := NewImageFromHTTPResponse(httpResponse)
	if err != nil {
		responseBody, _ := ioutil.ReadAll(httpResponse.Body)
		s.Logger.Warn("Unable to create image from response body: %v (url=%v)", string(responseBody), httpRequest.URL)
	}
	s.Logger.Info("Successfully retrieved image from S3: %v", httpRequest.URL)
	return image
}

func (s *S3ImageSource) signedHTTPRequestForRequest(request *ImageSourceOptions) *http.Request {
	httpRequest, _ := http.NewRequest("GET", s.imageURLForRequest(request), nil)
	httpRequest.Header.Set("Date", time.Now().UTC().Format(http.TimeFormat))
	s3.Sign(httpRequest, s3.Keys{
		AccessKey: s.Config.S3AccessKey,
		SecretKey: s.Config.S3SecretKey,
	})

	return httpRequest
}

func (s *S3ImageSource) imageURLForRequest(request *ImageSourceOptions) string {
	imageURLPathComponents := strings.Split(request.Path, "/")
	for index, component := range imageURLPathComponents {
		component = url.QueryEscape(component)
		imageURLPathComponents[index] = component
	}

	url := &url.URL{
		Scheme: "http",
		Host:   fmt.Sprintf("%s.s3.amazonaws.com", s.Config.S3Bucket),
		Path:   strings.Join(imageURLPathComponents, "/"),
	}

	return url.String()
}

func init() {
	RegisterSource(IMAGE_SOURCE_TYPE_S3, NewS3ImageSourceWithConfig)
}
