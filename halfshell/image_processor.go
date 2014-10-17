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
	"math"
	"strings"

	"github.com/oysterbooks/halfshell/halfshell/util"
	"github.com/rafikk/imagick/imagick"
)

// ImageProcessor is the public interface for the image processor. It exposes a
// single method to process an image with options.
type ImageProcessor interface {
	ProcessImage(*Image, *ImageProcessorOptions) *Image
}

// ImageProcessorOptions specify the request parameters for the processing
// operation.
type ImageProcessorOptions struct {
	Dimensions   ImageDimensions
	BlurRadius   float64
	CropMode     string
	BorderRadius uint64
}

type imageProcessor struct {
	Config *ProcessorConfig
	Logger *Logger
}

// NewImageProcessorWithConfig creates a new ImageProcessor instance using
// configuration settings.
func NewImageProcessorWithConfig(config *ProcessorConfig) ImageProcessor {
	return &imageProcessor{
		Config: config,
		Logger: NewLogger("image_processor.%s", config.Name),
	}
}

// The public method for processing an image. The method receives an original
// image and options and returns the processed image.
func (ip *imageProcessor) ProcessImage(image *Image, request *ImageProcessorOptions) *Image {
	processedImage := Image{}
	wand := imagick.NewMagickWand()
	defer wand.Destroy()

	wand.ReadImageBlob(image.Bytes)

	var orientModified bool
	var scaleModified bool
	var blurModified bool
	var radiusModified bool
	var err error

	if ip.Config.AutoOrient {
		orientModified, err = ip.orientWand(wand, request)
		if err != nil {
			ip.Logger.Warnf("Error orienting image: %s", err)
			return nil
		}
	}

	scaleModified, err = ip.scaleWand(wand, request)
	if err != nil {
		ip.Logger.Warnf("Error scaling image: %s", err)
		return nil
	}

	blurModified, err = ip.blurWand(wand, request)
	if err != nil {
		ip.Logger.Warnf("Error blurring image: %s", err)
		return nil
	}

	radiusModified, err = ip.radiusWand(wand, request)
	if err != nil {
		ip.Logger.Warnf("Error applying radius: %s", err)
		return nil
	}

	if !scaleModified && !blurModified && !radiusModified && !orientModified {
		processedImage.Bytes = image.Bytes
	} else {
		processedImage.Bytes = wand.GetImageBlob()
	}

	processedImage.MimeType = fmt.Sprintf("image/%s", strings.ToLower(wand.GetImageFormat()))

	return &processedImage
}

func (ip *imageProcessor) scaleWand(wand *imagick.MagickWand, request *ImageProcessorOptions) (modified bool, err error) {
	currentDimensions := ImageDimensions{uint64(wand.GetImageWidth()), uint64(wand.GetImageHeight())}
	newDimensions := ip.getScaledDimensions(currentDimensions, request)
	requestedDimensions := request.Dimensions

	if newDimensions == currentDimensions && newDimensions == requestedDimensions {
		return false, nil
	}

	if err = wand.ResizeImage(uint(newDimensions.Width), uint(newDimensions.Height), imagick.FILTER_LANCZOS, 1); err != nil {
		ip.Logger.Warnf("ImageMagick error resizing image: %s", err)
		return true, err
	}

	if request.CropMode == "fill" {
		if err = ip.cropImage(newDimensions, request.Dimensions, wand); err != nil {
			ip.Logger.Warnf("ImageMagick error cropping image: %s", err)
			return true, err
		}
	}

	if err = wand.SetImageInterpolateMethod(imagick.INTERPOLATE_PIXEL_BICUBIC); err != nil {
		ip.Logger.Warnf("ImageMagick error setting interpoliation method: %s", err)
		return true, err
	}

	if err = wand.StripImage(); err != nil {
		ip.Logger.Warnf("ImageMagick error stripping image routes and metadata")
		return true, err
	}

	if "JPEG" == wand.GetImageFormat() {
		if err = wand.SetImageInterlaceScheme(imagick.INTERLACE_PLANE); err != nil {
			ip.Logger.Warnf("ImageMagick error setting the image interlace scheme")
			return true, err
		}

		if err = wand.SetImageCompression(imagick.COMPRESSION_JPEG); err != nil {
			ip.Logger.Warnf("ImageMagick error setting the image compression type")
			return true, err
		}

		if err = wand.SetImageCompressionQuality(uint(ip.Config.ImageCompressionQuality)); err != nil {
			ip.Logger.Warnf("sImageMagick error setting compression quality: %s", err)
			return true, err
		}
	}

	return true, nil
}

func (ip *imageProcessor) blurWand(wand *imagick.MagickWand, request *ImageProcessorOptions) (modified bool, err error) {
	if request.BlurRadius != 0 {
		blurRadius := float64(wand.GetImageWidth()) * request.BlurRadius * ip.Config.MaxBlurRadiusPercentage
		if err = wand.GaussianBlurImage(blurRadius, blurRadius); err != nil {
			ip.Logger.Warnf("ImageMagick error setting blur radius: %s", err)
		}
		return true, err
	}
	return false, nil
}

func (ip *imageProcessor) radiusWand(wand *imagick.MagickWand, request *ImageProcessorOptions) (modified bool, err error) {
	radiusInt := util.FirstUInt(request.BorderRadius, ip.Config.DefaultBorderRadius, 0)
	if radiusInt == 0 {
		return
	}
	radius := float64(radiusInt)

	widthI := wand.GetImageWidth()
	heightI := wand.GetImageHeight()
	widthF := float64(widthI)
	heightF := float64(heightI)

	canvas := imagick.NewMagickWand()
	defer canvas.Destroy()

	transparent := imagick.NewPixelWand()
	defer transparent.Destroy()

	white := imagick.NewPixelWand()
	defer white.Destroy()

	mask := imagick.NewDrawingWand()
	defer mask.Destroy()

	border := imagick.NewDrawingWand()
	defer border.Destroy()

	transparent.SetColor("none")
	white.SetColor("white")

	canvas.NewImage(widthI, heightI, transparent)

	mask.SetFillColor(white)
	mask.RoundRectangle(2, 2, widthF, heightF, radius, radius)
	canvas.DrawImage(mask)

	canvas.CompositeImage(wand, imagick.COMPOSITE_OP_SRC_IN, 0, 0)
	canvas.OpaquePaintImage(transparent, white, 0, false)

	border.SetFillColor(transparent)
	border.SetStrokeColor(white)
	border.SetStrokeWidth(1)
	border.RoundRectangle(2, 2, widthF, heightF, radius, radius)
	canvas.DrawImage(border)

	canvas.SetImageFormat(wand.GetImageFormat())

	err = wand.SetImage(canvas)
	modified = true
	return
}

func (ip *imageProcessor) getScaledDimensions(currentDimensions ImageDimensions, request *ImageProcessorOptions) ImageDimensions {
	requestDimensions := request.Dimensions
	if requestDimensions.Width == 0 && requestDimensions.Height == 0 {
		requestDimensions = ImageDimensions{Width: ip.Config.DefaultImageWidth, Height: ip.Config.DefaultImageHeight}
	}

	dimensions := ip.scaleToRequestedDimensions(currentDimensions, requestDimensions, request)
	return ip.clampDimensionsToMaxima(dimensions, request)
}

func (ip *imageProcessor) orientWand(wand *imagick.MagickWand, request *ImageProcessorOptions) (modified bool, err error) {
	orientation := wand.GetImageOrientation()

	switch orientation {
	case imagick.ORIENTATION_UNDEFINED:
	case imagick.ORIENTATION_TOP_LEFT:
		return
	}

	transparent := imagick.NewPixelWand()
	defer transparent.Destroy()
	transparent.SetColor("none")

	switch orientation {
	case imagick.ORIENTATION_TOP_RIGHT:
		err = wand.FlopImage()
	case imagick.ORIENTATION_BOTTOM_RIGHT:
		err = wand.RotateImage(transparent, 180)
	case imagick.ORIENTATION_BOTTOM_LEFT:
		err = wand.FlipImage()
	case imagick.ORIENTATION_LEFT_TOP:
		err = wand.TransposeImage()
	case imagick.ORIENTATION_RIGHT_TOP:
		err = wand.RotateImage(transparent, 90)
	case imagick.ORIENTATION_RIGHT_BOTTOM:
		err = wand.TransverseImage()
		break
	case imagick.ORIENTATION_LEFT_BOTTOM:
		err = wand.RotateImage(transparent, 270)
	}

	wand.SetImageOrientation(imagick.ORIENTATION_TOP_LEFT)
	modified = true
	return
}

func (ip *imageProcessor) scaleToRequestedDimensions(currentDimensions, requestedDimensions ImageDimensions, request *ImageProcessorOptions) ImageDimensions {
	if requestedDimensions.Width == 0 && requestedDimensions.Height == 0 {
		return currentDimensions
	}

	imageAspectRatio := currentDimensions.AspectRatio()

	// No height was specified, thus image proportions should be retained.
	if requestedDimensions.Width > 0 && requestedDimensions.Height == 0 {
		height := ip.getAspectScaledHeight(imageAspectRatio, requestedDimensions.Width)
		return ImageDimensions{requestedDimensions.Width, height}
	}

	// No width was specified, thus image proportions should be retained.
	if requestedDimensions.Height > 0 && requestedDimensions.Width == 0 {
		width := ip.getAspectScaledWidth(imageAspectRatio, requestedDimensions.Height)
		return ImageDimensions{width, requestedDimensions.Height}
	}

	// The "stretch" crop mode is a NOOP, hence it's the default.
	cropMode := util.FirstString(request.CropMode, ip.Config.DefaultCropMode, "stretch")
	if cropMode == "stretch" {
		return requestedDimensions
	}

	// The "fit" crop mode retains the aspect ration while at least filling the
	// bounds requested. No cropping will occur.
	if cropMode == "fit" {
		requestedAspectRatio := requestedDimensions.AspectRatio()
		if requestedAspectRatio > imageAspectRatio {
			return ip.scaleToRequestedDimensions(currentDimensions, ImageDimensions{0, requestedDimensions.Height}, request)
		} else if requestedAspectRatio < imageAspectRatio {
			return ip.scaleToRequestedDimensions(currentDimensions, ImageDimensions{requestedDimensions.Width, 0}, request)
		}
		return requestedDimensions
	}

	// The "fill" crop mode will use the exact width/height and crop out the parts
	// that bleed out of the edges.
	//
	// Cropping does occur (handled elsewhere). The new dimensions defined here
	// ensure that clipping occurs on smallest edges possible. This is done by
	// bounding to the larger of the two axes.
	if cropMode == "fill" {
		requestedAspectRatio := requestedDimensions.AspectRatio()
		if requestedAspectRatio < imageAspectRatio {
			return ip.scaleToRequestedDimensions(currentDimensions, ImageDimensions{0, requestedDimensions.Height}, request)
		} else if requestedAspectRatio > imageAspectRatio {
			return ip.scaleToRequestedDimensions(currentDimensions, ImageDimensions{requestedDimensions.Width, 0}, request)
		}
		return requestedDimensions
	}

	// Unsupported crop modes are a NOOP.
	return requestedDimensions
}

func (ip *imageProcessor) clampDimensionsToMaxima(dimensions ImageDimensions, request *ImageProcessorOptions) ImageDimensions {
	if ip.Config.MaxImageWidth > 0 && dimensions.Width > ip.Config.MaxImageWidth {
		scaledHeight := ip.getAspectScaledHeight(dimensions.AspectRatio(), ip.Config.MaxImageWidth)
		return ip.clampDimensionsToMaxima(ImageDimensions{ip.Config.MaxImageWidth, scaledHeight}, request)
	}

	if ip.Config.MaxImageHeight > 0 && dimensions.Height > ip.Config.MaxImageHeight {
		scaledWidth := ip.getAspectScaledWidth(dimensions.AspectRatio(), ip.Config.MaxImageHeight)
		return ip.clampDimensionsToMaxima(ImageDimensions{scaledWidth, ip.Config.MaxImageHeight}, request)
	}

	return dimensions
}

func (ip *imageProcessor) cropImage(currentDimensions ImageDimensions, requestedDimensions ImageDimensions, wand *imagick.MagickWand) (err error) {
	err = wand.CropImage(
		uint(requestedDimensions.Width),
		uint(requestedDimensions.Height),
		int((currentDimensions.Width-requestedDimensions.Width)/2),
		int((currentDimensions.Height-requestedDimensions.Height)/2),
	)
	return
}

func (ip *imageProcessor) getAspectScaledHeight(aspectRatio float64, width uint64) uint64 {
	return uint64(math.Floor(float64(width)/aspectRatio + 0.5))
}

func (ip *imageProcessor) getAspectScaledWidth(aspectRatio float64, height uint64) uint64 {
	return uint64(math.Floor(float64(height)*aspectRatio + 0.5))
}
