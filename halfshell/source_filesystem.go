package halfshell

import (
	"os"
	"path/filepath"
)

const (
	IMAGE_SOURCE_TYPE_FILE_SYSTEM ImageSourceType = "filesystem"
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
		source.Logger.Info(source.Config.Directory, " does not exit. Creating.")
		_ = os.MkdirAll(source.Config.Directory, 0700)
		baseDirectory, err = os.Open(source.Config.Directory)
	}

	if err != nil {
		source.Logger.Fatal(err)
	}

	fileInfo, err := baseDirectory.Stat()
	if err != nil || !fileInfo.IsDir() {
		source.Logger.Fatal("Directory ", source.Config.Directory, " not a directory", err)
	}

	return source
}

func (s *FileSystemImageSource) GetImage(request *ImageSourceOptions) *Image {
	fileName := filepath.Join(s.Config.Directory, fileNameForRequest(request))
	image, err := NewImageFromPath(fileName)
	if err != nil {
		s.Logger.Warn("Failed to read image", err)
	}
	return image
}

func fileNameForRequest(request *ImageSourceOptions) string {
	return request.Path
}

func init() {
	RegisterSource(IMAGE_SOURCE_TYPE_FILE_SYSTEM, NewFileSystemImageSourceWithConfig)
}
