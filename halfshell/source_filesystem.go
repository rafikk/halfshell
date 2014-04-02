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
	"os"
	"path/filepath"
	"strings"
)

const (
	IMAGE_SOURCE_TYPE_FILESYSTEM ImageSourceType = "filesystem"
)

type FileSystemImageSource struct {
	Config *SourceConfig
	Logger *Logger
}

var BlankImage = &Image{
	Bytes:    make([]byte, 0),
	MimeType: "",
}

func NewFileSystemImageSourceWithConfig(config *SourceConfig) ImageSource {
	source := &FileSystemImageSource{
		Config: config,
		Logger: NewLogger("source.fs.%s", config.Name),
	}

	baseDirectory, err := os.Open(source.Config.Directory)
	if os.IsNotExist(err) {
		source.Logger.Info("%s does not exit. Creating.", source.Config.Directory)
		_ = os.MkdirAll(source.Config.Directory, 0700)
		baseDirectory, err = os.Open(source.Config.Directory)
	}

	if err != nil {
		source.Logger.Fatal(err)
	}

	fileInfo, err := baseDirectory.Stat()
	if err != nil || !fileInfo.IsDir() {
		source.Logger.Fatal("Directory %s not a directory", source.Config.Directory, err)
	}

	return source
}

func (s *FileSystemImageSource) GetImage(request *ImageSourceOptions) *Image {
	fileName := s.fileNameForRequest(request)

	file, err := os.Open(fileName)
	if err != nil {
		s.Logger.Warn("Failed to open file: %v", err)
		return nil
	}

	image, err := NewImageFromFile(file)
	if err != nil {
		s.Logger.Warn("Failed to read image: %v", err)
		return nil
	}
	return image
}

func (s *FileSystemImageSource) fileNameForRequest(request *ImageSourceOptions) string {
	// Remove the leading / from the file name and replace the
	// directory separator (/) with something safe for file names (_)
	return filepath.Join(s.Config.Directory, strings.Replace(strings.TrimLeft(request.Path, string(filepath.Separator)), string(filepath.Separator), "_", -1))
}

func init() {
	RegisterSource(IMAGE_SOURCE_TYPE_FILESYSTEM, NewFileSystemImageSourceWithConfig)
}
