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
	"math"

	"github.com/rafikk/imagick/imagick"
)

const (
	ScaleFill       = 10
	ScaleAspectFit  = 21
	ScaleAspectFill = 22
	ScaleAspectCrop = 23
)

var ScaleModes = map[string]uint{
	"fill":        ScaleFill,
	"aspect_fit":  ScaleAspectFit,
	"aspect_fill": ScaleAspectFill,
	"aspect_crop": ScaleAspectCrop,
}

type ImageProcessor interface {
	ProcessImage(*Image, *ImageProcessorOptions) error
}

type ImageProcessorOptions struct {
	Dimensions ImageDimensions
	BlurRadius float64
	ScaleMode  uint
	Focalpoint Focalpoint
}

type imageProcessor struct {
	Config *ProcessorConfig
	Logger *Logger
}

func NewImageProcessorWithConfig(config *ProcessorConfig) ImageProcessor {
	return &imageProcessor{
		Config: config,
		Logger: NewLogger("image_processor.%s", config.Name),
	}
}

func (ip *imageProcessor) ProcessImage(img *Image, req *ImageProcessorOptions) error {
	if req.Dimensions == EmptyImageDimensions {
		req.Dimensions.Width = uint(ip.Config.DefaultImageWidth)
		req.Dimensions.Height = uint(ip.Config.DefaultImageHeight)
	}

	var err error

	err = ip.orient(img, req)
	if err != nil {
		ip.Logger.Errorf("Error orienting image: %s", err)
		return err
	}

	err = ip.resize(img, req)
	if err != nil {
		ip.Logger.Errorf("Error resizing image: %s", err)
		return err
	}

	err = ip.blur(img, req)
	if err != nil {
		ip.Logger.Errorf("Error blurring image: %s", err)
		return err
	}

	return nil
}

func (ip *imageProcessor) orient(img *Image, req *ImageProcessorOptions) error {
	if !ip.Config.AutoOrient {
		return nil
	}

	orientation := img.Wand.GetImageOrientation()

	switch orientation {
	case imagick.ORIENTATION_UNDEFINED:
	case imagick.ORIENTATION_TOP_LEFT:
		return nil
	}

	transparent := imagick.NewPixelWand()
	defer transparent.Destroy()
	transparent.SetColor("none")

	var err error

	switch orientation {
	case imagick.ORIENTATION_TOP_RIGHT:
		err = img.Wand.FlopImage()
	case imagick.ORIENTATION_BOTTOM_RIGHT:
		err = img.Wand.RotateImage(transparent, 180)
	case imagick.ORIENTATION_BOTTOM_LEFT:
		err = img.Wand.FlipImage()
	case imagick.ORIENTATION_LEFT_TOP:
		err = img.Wand.TransposeImage()
	case imagick.ORIENTATION_RIGHT_TOP:
		err = img.Wand.RotateImage(transparent, 90)
	case imagick.ORIENTATION_RIGHT_BOTTOM:
		err = img.Wand.TransverseImage()
	case imagick.ORIENTATION_LEFT_BOTTOM:
		err = img.Wand.RotateImage(transparent, 270)
	}

	if err != nil {
		return err
	}

	return img.Wand.SetImageOrientation(imagick.ORIENTATION_TOP_LEFT)
}

func (ip *imageProcessor) resize(img *Image, req *ImageProcessorOptions) error {
	scaleMode := req.ScaleMode
	if scaleMode == 0 {
		scaleMode = ip.Config.DefaultScaleMode
	}

	resize, err := ip.resizePrepare(img.GetDimensions(), req.Dimensions, scaleMode)
	if err != nil {
		return err
	}

	if resize.Scale != EmptyImageDimensions {
		err = ip.resizeApply(img, resize.Scale)
		if err != nil {
			return err
		}
	}

	if resize.Crop != EmptyImageDimensions {
		err = ip.cropApply(img, resize.Crop, req.Focalpoint)
		if err != nil {
			return err
		}
	}

	return nil
}

func (ip *imageProcessor) resizePrepare(oldDimensions, reqDimensions ImageDimensions, scaleMode uint) (*ResizeDimensions, error) {
	resize := &ResizeDimensions{
		Scale: ImageDimensions{},
		Crop:  ImageDimensions{},
	}

	if reqDimensions == EmptyImageDimensions {
		return resize, nil
	}
	if oldDimensions == reqDimensions {
		return resize, nil
	}

	reqDimensions = clampDimensionsToMaxima(oldDimensions, reqDimensions, ip.Config.MaxImageDimensions)
	oldAspectRatio := oldDimensions.AspectRatio()

	// Unspecified dimensions are automatically computed relative to the specified
	// dimension using the old image's aspect ratio.
	if reqDimensions.Width > 0 && reqDimensions.Height == 0 {
		reqDimensions.Height = aspectHeight(oldAspectRatio, reqDimensions.Width)
	} else if reqDimensions.Height > 0 && reqDimensions.Width == 0 {
		reqDimensions.Width = aspectWidth(oldAspectRatio, reqDimensions.Height)
	}

	// Retain the aspect ratio while at least filling the bounds requested. No
	// cropping will occur but the image will be resized.
	if scaleMode == ScaleAspectFit {
		newAspectRatio := reqDimensions.AspectRatio()
		if newAspectRatio > oldAspectRatio {
			resize.Scale.Width = aspectWidth(oldAspectRatio, reqDimensions.Height)
			resize.Scale.Height = reqDimensions.Height
		} else if newAspectRatio < oldAspectRatio {
			resize.Scale.Width = reqDimensions.Width
			resize.Scale.Height = aspectHeight(oldAspectRatio, reqDimensions.Width)
		} else {
			resize.Scale.Width = reqDimensions.Width
			resize.Scale.Height = reqDimensions.Height
		}
		return resize, nil
	}

	// Retain the aspect ratio while filling the bounds requested completely. New
	// dimensions are at least as large as the requested dimensions. No cropping
	// will occur but the image will be resized.
	if scaleMode == ScaleAspectFill {
		newAspectRatio := reqDimensions.AspectRatio()
		if newAspectRatio < oldAspectRatio {
			resize.Scale.Width = aspectWidth(oldAspectRatio, reqDimensions.Height)
			resize.Scale.Height = reqDimensions.Height
		} else if newAspectRatio > oldAspectRatio {
			resize.Scale.Width = reqDimensions.Width
			resize.Scale.Height = aspectHeight(oldAspectRatio, reqDimensions.Width)
		} else {
			resize.Scale.Width = reqDimensions.Width
			resize.Scale.Height = reqDimensions.Height
		}
		return resize, nil
	}

	// Use exact width/height and clip off the parts that bleed. The image is
	// first resized to ensure clipping occurs on the smallest edges possible.
	if scaleMode == ScaleAspectCrop {
		newAspectRatio := reqDimensions.AspectRatio()
		if newAspectRatio > oldAspectRatio {
			resize.Scale.Width = reqDimensions.Width
			resize.Scale.Height = aspectHeight(oldAspectRatio, reqDimensions.Width)
		} else if newAspectRatio < oldAspectRatio {
			resize.Scale.Width = aspectWidth(oldAspectRatio, reqDimensions.Height)
			resize.Scale.Height = reqDimensions.Height
		} else {
			resize.Scale.Width = reqDimensions.Width
			resize.Scale.Height = reqDimensions.Height
		}
		resize.Crop.Width = reqDimensions.Width
		resize.Crop.Height = reqDimensions.Height
		return resize, nil
	}

	// Use the new dimensions exactly as is. Don't correct for aspect ratio and
	// don't do any cropping. This is equivalent to ScaleFill.
	resize.Scale = reqDimensions
	return resize, nil
}

func (ip *imageProcessor) resizeApply(img *Image, dimensions ImageDimensions) error {
	if dimensions == EmptyImageDimensions {
		return nil
	}

	err := img.Wand.ResizeImage(dimensions.Width, dimensions.Height, imagick.FILTER_LANCZOS, 1)
	if err != nil {
		ip.Logger.Errorf("Failed resizing image: %s", err)
		return err
	}

	err = img.Wand.SetImageInterpolateMethod(imagick.INTERPOLATE_PIXEL_BICUBIC)
	if err != nil {
		ip.Logger.Errorf("Failed getting interpolation method: %s", err)
		return err
	}

	err = img.Wand.StripImage()
	if err != nil {
		ip.Logger.Errorf("Failed stripping image metadata: %s", err)
		return err
	}

	if img.Wand.GetImageFormat() == "JPEG" {
		err = img.Wand.SetImageInterlaceScheme(imagick.INTERLACE_PLANE)
		if err != nil {
			ip.Logger.Errorf("Failed setting image interlace scheme: %s", err)
			return err
		}

		err = img.Wand.SetImageCompression(imagick.COMPRESSION_JPEG)
		if err != nil {
			ip.Logger.Errorf("Failed setting image compression type: %s", err)
			return err
		}

		err = img.Wand.SetImageCompressionQuality(uint(ip.Config.ImageCompressionQuality))
		if err != nil {
			ip.Logger.Errorf("Failed setting compression quality: %s", err)
			return err
		}
	}

	return nil
}

func (ip *imageProcessor) cropApply(img *Image, reqDimensions ImageDimensions, focalpoint Focalpoint) error {
	oldDimensions := img.GetDimensions()
	x := int(focalpoint.X * (float64(oldDimensions.Width) - float64(reqDimensions.Width)))
	y := int(focalpoint.Y * (float64(oldDimensions.Height) - float64(reqDimensions.Height)))
	w := reqDimensions.Width
	h := reqDimensions.Height
	return img.Wand.CropImage(w, h, x, y)
}

func (ip *imageProcessor) blur(image *Image, request *ImageProcessorOptions) error {
	if request.BlurRadius == 0 {
		return nil
	}
	blurRadius := float64(image.GetWidth()) * request.BlurRadius * ip.Config.MaxBlurRadiusPercentage
	return image.Wand.GaussianBlurImage(blurRadius, blurRadius)
}

func aspectHeight(aspectRatio float64, width uint) uint {
	return uint(math.Floor(float64(width)/aspectRatio + 0.5))
}

func aspectWidth(aspectRatio float64, height uint) uint {
	return uint(math.Floor(float64(height)*aspectRatio + 0.5))
}

func clampDimensionsToMaxima(imgDimensions, reqDimensions, maxDimensions ImageDimensions) ImageDimensions {
	if maxDimensions.Width > 0 && reqDimensions.Width > maxDimensions.Width {
		reqDimensions.Width = maxDimensions.Width
		reqDimensions.Height = aspectHeight(imgDimensions.AspectRatio(), maxDimensions.Width)
		return clampDimensionsToMaxima(imgDimensions, reqDimensions, maxDimensions)
	}

	if maxDimensions.Height > 0 && reqDimensions.Height > maxDimensions.Height {
		reqDimensions.Width = aspectWidth(imgDimensions.AspectRatio(), maxDimensions.Height)
		reqDimensions.Height = maxDimensions.Height
		return clampDimensionsToMaxima(imgDimensions, reqDimensions, maxDimensions)
	}

	return reqDimensions
}
