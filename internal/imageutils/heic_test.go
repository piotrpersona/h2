package imageutils_test

import (
	_ "embed"
	"h2img/internal/imageutils"
	"image/jpeg"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

//go:embed images/image.HEIC
var heicImageBytes []byte

func TestDecode(t *testing.T) {
	t.Parallel()

	t.Run("decode HEIC image", func(t *testing.T) {
		t.Parallel()

		img, err := imageutils.NewHeicDecoder().DecodeFromBytes(heicImageBytes)
		require.NoError(t, err)
		require.NotNil(t, img)

		fh, err := os.Create("images/image.jpeg")
		require.NoError(t, err)

		err = jpeg.Encode(fh, img, nil)
		require.NoError(t, err)
	})
}
