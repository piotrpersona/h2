package imageutils

import (
	"image"

	"github.com/strukturag/libheif/go/heif"
)

type HeicDecoder struct{}

func NewHeicDecoder() *HeicDecoder {
	return &HeicDecoder{}
}

func (h *HeicDecoder) DecodeFromBytes(heicBytes []byte) (image.Image, error) {
	heifCtx, err := heif.NewContext()
	if err != nil {
		return nil, err
	}
	if err := heifCtx.ReadFromMemory(heicBytes); err != nil {
		return nil, err
	}
	return h.decodeFromContext(heifCtx)
}

func (h *HeicDecoder) DecodeFromFile(path string) (image.Image, error) {
	heifCtx, err := heif.NewContext()
	if err != nil {
		return nil, err
	}
	if err := heifCtx.ReadFromFile(path); err != nil {
		return nil, err
	}
	return h.decodeFromContext(heifCtx)
}

func (h *HeicDecoder) decodeFromContext(heifCtx *heif.Context) (image.Image, error) {
	handle, err := heifCtx.GetPrimaryImageHandle()
	if err != nil {
		return nil, err
	}
	heicImg, err := handle.DecodeImage(heif.ColorspaceRGB, heif.ChromaInterleavedRGB, nil)
	if err != nil {
		return nil, err
	}
	img, err := heicImg.GetImage()
	if err != nil {
		return nil, err
	}
	return img, nil
}
