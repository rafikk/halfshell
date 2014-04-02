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
	"net"
	"net/http"
	"time"
)

type Server struct {
	*http.Server
	Routes []*Route
	Logger *Logger
	Config *ServerConfig
}

func NewServerWithConfigAndRoutes(config *ServerConfig, routes []*Route) *Server {
	httpServer := &http.Server{
		Addr:           fmt.Sprintf(":%d", config.Port),
		ReadTimeout:    time.Duration(config.ReadTimeout) * time.Second,
		WriteTimeout:   time.Duration(config.WriteTimeout) * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	server := &Server{httpServer, routes, NewLogger("server"), config}
	httpServer.Handler = server
	return server
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	hw := s.NewHalfshellResponseWriter(w)
	hr := s.NewHalfshellRequest(r)
	defer s.LogRequest(hw, hr)
	switch {
	case "/healthcheck" == hr.URL.Path || "/health" == hr.URL.Path:
		hw.Write([]byte("OK"))
	default:
		s.ImageRequestHandler(hw, hr)
	}
}

func (s *Server) ImageRequestHandler(w *HalfshellResponseWriter, r *HalfshellRequest) {
	if r.Route == nil {
		w.WriteError(fmt.Sprintf("No route available to handle request: %v",
			r.URL.Path), http.StatusNotFound)
		return
	}

	if !s.Config.StatsdDisabled {
		defer func() { go r.Route.Statter.RegisterRequest(w, r) }()
	}

	s.Logger.Info("Handling request for image %s with dimensions %v",
		r.SourceOptions.Path, r.ProcessorOptions.Dimensions)

	image := r.Route.Source.GetImage(r.SourceOptions)
	if image == nil {
		w.WriteError("Not Found", http.StatusNotFound)
		return
	}

	processedImage := r.Route.Processor.ProcessImage(image, r.ProcessorOptions)
	if processedImage == nil {
		s.Logger.Warn("Error processing image data %s to dimensions: %v",
			r.ProcessorOptions.Dimensions)
		w.WriteError("Internal Server Error", http.StatusNotFound)
		return
	}

	s.Logger.Info("Returning resized image %s to dimensions %v",
		r.SourceOptions.Path, r.ProcessorOptions.Dimensions)
	w.WriteImage(processedImage)
}

func (s *Server) LogRequest(w *HalfshellResponseWriter, r *HalfshellRequest) {
	logFormat := "%s - - [%s] \"%s %s %s\" %d %d\n"
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}
	fmt.Printf(logFormat, host, r.Timestamp.Format("02/Jan/2006:15:04:05 -0700"),
		r.Method, r.URL.RequestURI(), r.Proto, w.Status, w.Size)
}

type HalfshellRequest struct {
	*http.Request
	Timestamp        time.Time
	Route            *Route
	SourceOptions    *ImageSourceOptions
	ProcessorOptions *ImageProcessorOptions
}

func (s *Server) NewHalfshellRequest(r *http.Request) *HalfshellRequest {
	request := &HalfshellRequest{r, time.Now(), nil, nil, nil}
	for _, route := range s.Routes {
		if route.ShouldHandleRequest(r) {
			request.Route = route
		}
	}

	if request.Route != nil {
		request.SourceOptions, request.ProcessorOptions =
			request.Route.SourceAndProcessorOptionsForRequest(r)
	}

	return request
}

// HalfshellResponseWriter is a wrapper around http.ResponseWriter that provides
// access to the response status and size after they have been set.
type HalfshellResponseWriter struct {
	w      http.ResponseWriter
	Status int
	Size   int
}

// Create a new HalfshellResponseWriter by wrapping http.ResponseWriter.
func (s *Server) NewHalfshellResponseWriter(w http.ResponseWriter) *HalfshellResponseWriter {
	return &HalfshellResponseWriter{w: w}
}

// Forwards to http.ResponseWriter's WriteHeader method.
func (hw *HalfshellResponseWriter) WriteHeader(status int) {
	hw.Status = status
	hw.w.WriteHeader(status)
}

// Sets the value for a response header.
func (hw *HalfshellResponseWriter) SetHeader(name, value string) {
	hw.w.Header().Set(name, value)
}

// Writes data the output stream.
func (hw *HalfshellResponseWriter) Write(data []byte) (int, error) {
	hw.Size += len(data)
	return hw.w.Write(data)
}

// Writes an error response.
func (hw *HalfshellResponseWriter) WriteError(message string, status int) {
	hw.SetHeader("Content-Type", "text/plain; charset=utf-8")
	hw.WriteHeader(status)
	hw.Write([]byte(message))
}

// Writes an image to the output stream and sets the appropriate headers.
func (hw *HalfshellResponseWriter) WriteImage(image *Image) {
	hw.SetHeader("Content-Type", image.MimeType)
	hw.SetHeader("Content-Length", fmt.Sprintf("%d", len(image.Bytes)))
	hw.SetHeader("Cache-Control", "no-transform,public,max-age=86400,s-maxage=2592000")
	hw.WriteHeader(http.StatusOK)
	hw.Write(image.Bytes)
}
