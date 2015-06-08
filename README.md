# Halfshell

Halfshell is a proxy server for processing images on the fly. It allows you to dynamically resize (and apply effects to) images hosted on S3, a local filesystem or an http source via query parameters. It supports creating “families” of images which can read from distinct image sources and enable different configuration values for image processing and retrieval. See the [introduction blog post](http://engineering.oysterbooks.com/post/79458380259/resizing-images-on-the-fly-with-go).

Current version: `0.1.1`

## Architecture

Halfshell was architected to be extensible from the beginning. The system is composed of a few components with their own configuration and simple interfaces.

### Sources

Sources are repositories from which an “original” image can be loaded. They return an image given a path. Currently, sources for downloading images from S3, a local filesystem and http are included.

### Processors

Processors perform all image manipulation. They accept an image and a set of options and return a modified image. Out of the box, the default processor supports resizing and blurring images. Each processor can be configured with maximum and default image dimensions and enable/disable certain features.

### Routes

Routes bind URL rules (regular expressions) with a source and a processor. Halfshell supports setting up an arbitrary number of routes, and sources and processors do not need to correspond 1-1 with routes.

When Halfshell receives a request, it determines the matching route, retrieves the image from its source, and processes the image using its processor.

This simple architecture has allowed us to serve images from multiple S3 buckets and maintain isolated configuration settings for each family of images.

## Usage and Configuration

Halfshell uses a JSON file for configuration. An example is shown below:

```json
{
    "server": {
        "port": 8080,
        "read_timeout": 5,
        "write_timeout": 30
    },
    "sources": {
        "default": {
            "type": "s3",
            "s3_access_key": "<S3_ACCESS_KEY>",
            "s3_secret_key": "<S3_SECRET_KEY>"
        },
        "blog-post-images": {
            "s3_bucket": "my-company-blog-post-images"
        },
        "profile-photos": {
            "s3_bucket": "my-company-profile-photos"
        }
    },
    "processors": {
        "default": {
            "image_compression_quality": 85,
            "default_scale_mode": "aspect_fit",
            "max_blur_radius_percentage": 0,
            "max_image_height": 0,
            "max_image_width": 1000

        },
        "profile-photos": {
            "default_image_width": 120
        }
    },
    "routes": {
        "^/blog(?P<image_path>/.*)$": {
            "name": "blog-post-images",
            "source": "blog-post-images",
            "processor": "default",
            "cache_control": "no-transform,public,max-age=2592000,s-maxage=31104000"
        },
        "^/users(?P<image_path>/.*)$": {
            "name": "profile-photos",
            "source": "profile-photos",
            "processor": "profile-photos"
        }
    }
}
```

To start the server, pass configuration file path as an argument.

```bash
$ ./bin/halfshell config.json
```

This will start the server on port 8080, and service requests whose path begins with /users/ or /blog/, e.g.:

    http://localhost:8080/users/joe/default.jpg?w=100&h=100
    http://localhost:8080/blog/posts/announcement.jpg?w=600&h=200

The image_host named group in the route pattern match (e.g., `^/users(?P<image_path>/.*)$`) gets extracted as the request path for the source. In this instance, the file “joe/default.jpg” is requested from the “my-company-profile-photos” S3 bucket. The processor resizes the image to a width and height of 100.

### Server

The `server` configuration block accepts the following settings:

##### port

The port to run the server on.

##### read_timeout

The timeout in seconds for reading the initial data from the connection.

##### write_timeout

The timeout in seconds for writing the image data backto the connection.

### Sources

The `sources` block is a mapping of source names to source configuration values.
Values from a source named `default` will be inherited by all other sources.

##### type

The type of image source. Currently `s3` or `filesystem`.

##### s3_access_key

For the S3 source type, the access key to read from S3.

##### s3_secret_key

For the S3 source type, the secret key to read from S3.

##### s3_bucket

For the S3 source type, the bucket to request images from.

##### directory

For the Filesystem source type, the local directory to request images from. Required.
For the S3 source type, `directory` corresponds to an optional base directory in the S3 bucket.

### Processors

The `processors` block is a mapping of processor names to processor configuration values.
Values from a processor named `default` will be inherited by all other processors.

##### image_compression_quality

The compression quality to use for JPEG images.

##### maintain_aspect_ratio

DEPRECATED: Use the `aspect_fit` `scale_mode` instead.

If this is set to true, the resized images will always maintain the original
aspect ratio. When set to false, the image will be stretched to fit the width
and height requested.

##### default_scale_mode

When changing the dimensions of an image, you may want to crop edges or
constrain proportions. Use the `default_scale_mode` setting to define these
rules (`scale_mode` as a URL query parameter).

A value of `aspect_fit` will change the image size to fit in the given
dimensions while retaining original proportions. No part of the image will be
cut away.

A value of `aspect_fill` will change the image size to at least fit the given
dimensions while retaining original proportions. No part of the image will be
cut away.

A value of `aspect_crop` will change the image size to fit in the given
dimensions while retaining original proportions. Edges that do not fit in the
given dimensions will be cut off.

The default behavior is to `fill`, which changes the image size to fit the given
dimensions and will NOT retain the original proportions.

Cheat Sheet:

    Image dimensions: 500x800
    Requested dimensions: 400x400

    Scale mode: fill
    New dimensions: 400x400
    Maintain aspect ratio: NO
    Cropping: NO

    Scale mode: aspect_fit
    New dimensions: 250x400
    Maintain aspect ratio: YES
    Cropping: NO

    Scale mode: aspect_fill
    New dimensions: 400x640
    Maintain aspect ratio: YES
    Cropping: NO

    Scale mode: aspect_crop
    New dimensions: 400x400
    Maintain aspect ratio: YES
    Cropping: YES

##### default_image_width

In the absence of a width parameter in the request, use this as image width. A
value of `0` sets no default.
##### default_image_height

In the absence of a height parameter in the request, use this as image height.
A value of `0` sets no default.

##### max_image_width

Set a maximum image width. A value of `0` specifies no maximum.

##### max_image_height

Set a maximum image height. A value of `0` specifies no maximum.

##### max_blur_radius_percentage

Set a maximum blur radius percentage. A value of `0` disables blurring images.
For Gaussian blur, the radius used is this value * the image width. This allows
you to use a blur parameter (from 0-1) which will apply the same proportion of
blurring to each image size.

##### auto_orient

If set to true, the image processor will respect EXIF rotation data. A common
case are photos taken with a camera (eg: iPhone, digital camera) in landscape
mode. The built-in gyroscope will embed rotation data in the image via EXIF.

Disabled by default.

##### formats

```
formats: {
    "large": { "width": 1280, "height": 768, "blur": 0 },
    "medium": { "width": 640, "height": 480, "blur": 0 }
}
```

If specified, the `w`, `h` and `blur` parameters will be ignored from the
request. Instead will only be read the `format` parameter.

### Routes

The `routes` block is a mapping of route patterns to route configuration values.

The route pattern is a regular expression with a captured group for `image_path`.
The subexpression match is the path that is requested from the image source.

##### name

The name to use for the route. This is currently used in logging and StatsD key
names.

##### source

The name of the source to use for the route.

##### processor

The name of the processor to use for the route.

##### cache_control

The Cache-Control response header to set. If left empty or unspecified, `no-transform,public,max-age=86400,s-maxage=2592000` will be set.

### Health Checks

You can check the server health at `/healthcheck` and `/health`. If the server
is up and running, the HTTP client will receive a response with status code
`200`.

## Adopters

- [Oyster](https://www.oysterbooks.com)
- [Storehouse](https://www.storehouse.co)

If your organization is using Halfshell, consider adding a link and sending us a pull request!

## Contributing

Contributions are welcome.

### Building

There's a Vagrant file set up to ease development. After you have the
Vagrant box set up, cd to the /vagrant directory and run `make`.

### Notes

Run `make format` before sending any pull requests.

### Questions?

File an issue or send an email to rafik@oysterbooks.com.
