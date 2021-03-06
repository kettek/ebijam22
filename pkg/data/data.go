package data

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

var (
	// Images for drawing lines.
	EmptyImage    *ebiten.Image
	EmptySubImage *ebiten.Image

	// I'm sorry for violating our principles.
	OrbTinyImages   []*ebiten.Image
	OrbSmallImages  []*ebiten.Image
	OrbMediumImages []*ebiten.Image
	OrbLargeImages  []*ebiten.Image

	CellWidth, CellHeight int
	NormalFace, BoldFace  font.Face
)

// Images is the map of all loaded images.
var Images map[string]*ebiten.Image = make(map[string]*ebiten.Image)

// GetImage returns the image matching the given file name. IT ALSO LOADS IT.
func GetImage(p string) (*ebiten.Image, error) {
	if v, ok := Images[p]; ok {
		return v, nil
	}
	if img, err := ReadImage(p); err != nil {
		return nil, err
	} else {
		eimg := ebiten.NewImageFromImage(img)
		Images[p] = eimg
		return eimg, nil
	}
}

// Sounds is the map of all loaded sounds.
var Sounds map[string]*Sound = make(map[string]*Sound)

// GetSound returns the sound matching the given file name. IT ALSO LOADS IT.
func GetSound(p string) (*Sound, error) {
	if v, ok := Sounds[p]; ok {
		return v, nil
	}
	if snd, err := ReadSound(p); err != nil {
		return nil, err
	} else {
		Sounds[p] = snd
		return snd, nil
	}
}
func GetMusic(p string) (*Sound, error) {
	if v, ok := Sounds[p]; ok {
		return v, nil
	}
	if snd, err := ReadMusic(p); err != nil {
		return nil, err
	} else {
		Sounds[p] = snd
		return snd, nil
	}
}

// LoadData loads some data.
func LoadData() error {
	//
	EmptyImage = ebiten.NewImage(3, 3)
	EmptySubImage = EmptyImage.SubImage(image.Rect(1, 1, 2, 2)).(*ebiten.Image)
	EmptyImage.Fill(color.White)

	// Load the fonts.
	d, err := ReadFile("fonts/x12y16pxMaruMonica.ttf")
	if err != nil {
		return err
	}
	tt, err := opentype.Parse(d)
	if err != nil {
		return err
	}
	if NormalFace, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    16,
		DPI:     72,
		Hinting: font.HintingFull,
	}); err != nil {
		return err
	}
	d, err = ReadFile("fonts/x12y16pxMaruMonica.ttf")
	if err != nil {
		return err
	}
	tt, err = opentype.Parse(d)
	if err != nil {
		return err
	}
	if BoldFace, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    16,
		DPI:     72,
		Hinting: font.HintingFull,
	}); err != nil {
		return err
	}

	// Load the orbs.
	imgs, err := ReadImagesByPrefix("orb-tiny")
	if err != nil {
		return err
	}
	for _, img := range imgs {
		OrbTinyImages = append(OrbTinyImages, ebiten.NewImageFromImage(img))
	}
	imgs, err = ReadImagesByPrefix("orb-small")
	if err != nil {
		return err
	}
	for _, img := range imgs {
		OrbSmallImages = append(OrbSmallImages, ebiten.NewImageFromImage(img))
	}
	imgs, err = ReadImagesByPrefix("orb-medium")
	if err != nil {
		return err
	}
	for _, img := range imgs {
		OrbMediumImages = append(OrbMediumImages, ebiten.NewImageFromImage(img))
	}
	imgs, err = ReadImagesByPrefix("orb-large")
	if err != nil {
		return err
	}
	for _, img := range imgs {
		OrbLargeImages = append(OrbLargeImages, ebiten.NewImageFromImage(img))
	}

	return nil
}
