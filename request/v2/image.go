package request

import (
	"image"
	"net/http"
)

// Make request and return [image.Image]
//
// Require import modules to [image.Decode], see [image] to more info
func Image(Url string, RequestOption *Options) (image.Image, string, *http.Response, error) {
	res, err := Request(Url, RequestOption)
	if err != nil {
		return nil, "", nil, err
	}
	defer res.Body.Close()

	imageDecode, imageType, err := image.Decode(res.Body)
	return imageDecode, imageType, res, err
}

// Make request and return [image.DecodeConfig]
//
// Require import modules to [image.DecodeConfig], see [image] to more info
func ImageConfig(Url string, RequestOption *Options) (image.Config, string, *http.Response, error) {
	res, err := Request(Url, RequestOption)
	if err != nil {
		return image.Config{}, "", nil, err
	}
	defer res.Body.Close()

	imageDecode, imageType, err := image.DecodeConfig(res.Body)
	return imageDecode, imageType, res, err
}
