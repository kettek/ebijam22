package game

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/kettek/ebijam22/pkg/data"
	"github.com/kettek/ebijam22/pkg/data/assets/lang"
	"github.com/kettek/ebijam22/pkg/data/ui"
	"github.com/kettek/ebijam22/pkg/world"
)

type SoloMenuState struct {
	game        *Game
	title       string
	magnetImage *ebiten.Image
	magnetSpin  float64
	mapList     MapList

	tiledBackgroundImages  []*ebiten.Image
	tiledBackgroundElapsed int
	tiledBackgroundIndex   int
	backgroundImage        *ebiten.Image

	buttons []*data.Button
}

func (s *SoloMenuState) Init() error {
	// Oh boy.
	t, err := data.LoadTileSet("nature")
	if err != nil {
		return err
	}
	s.tiledBackgroundImages = t.BackgroundImages

	// Load our background images.
	if img, err := data.ReadImage("/ui/singleplayer.png"); err == nil {
		s.backgroundImage = ebiten.NewImageFromImage(img)
	} else {
		panic(err)
	}

	if err := s.mapList.Init(); err != nil {
		return err
	}
	if s.game.Options.Map != "" {
		s.mapList.selectedMap = s.game.Options.Map
	}

	// Title Text
	s.title = lang.SoloGame

	// Magnet Image
	if img, err := data.ReadImage("/ui/magnet.png"); err == nil {
		s.magnetImage = ebiten.NewImageFromImage(img)
	} else {
		panic(err)
	}

	centeredX := world.ScreenWidth / 2

	buttonY := int(float64(world.ScreenHeight) * 0.85)

	// Create Buttons
	backButton := data.NewButton(
		15,
		10,
		lang.Back,
		func() {
			s.game.SetState(&MenuState{
				game: s.game,
			})
		},
	)
	backButton.Hover = true

	startGameButton := data.NewButton(
		centeredX,
		buttonY,
		lang.StartGame,
		func() {
			s.StartGame()
		},
	)
	startGameButton.Hover = true

	s.buttons = []*data.Button{
		backButton,
		startGameButton,
	}

	return nil
}

func (s *SoloMenuState) Dispose() error {
	return nil
}

func (s *SoloMenuState) Update() error {
	// Animate the background.
	s.tiledBackgroundElapsed++
	if s.tiledBackgroundElapsed >= 30 {
		s.tiledBackgroundElapsed = 0
		s.tiledBackgroundIndex++
		if s.tiledBackgroundIndex >= len(s.tiledBackgroundImages) {
			s.tiledBackgroundIndex = 0
		}
	}
	// Spin at 4 degrees per update.
	s.magnetSpin -= math.Pi / 180 * 4

	// Update buttons
	for _, button := range s.buttons {
		button.Update()
	}

	s.mapList.Update()

	return nil
}

func (s *SoloMenuState) Draw(screen *ebiten.Image) {
	// Draw our tiled background.
	bgOp := ebiten.DrawImageOptions{}
	ui.DrawTiled(screen, s.tiledBackgroundImages[s.tiledBackgroundIndex], &bgOp, world.ScreenWidth, world.ScreenHeight)

	// Draw our background.
	screenOp := &ebiten.DrawImageOptions{}
	screenOp.ColorM.Scale(0.5, 0.5, 0.5, 1)
	screen.DrawImage(s.backgroundImage, screenOp)

	// Draw our title
	titleBounds := data.DrawStaticTextByCode(
		s.title,
		data.BoldFace,
		world.ScreenWidth/2,
		world.ScreenHeight/8,
		color.White,
		screen,
		true,
	)

	// Rotate our magnet about its center.
	magnetOp := ebiten.DrawImageOptions{}
	magnetOp.GeoM.Translate(-float64(s.magnetImage.Bounds().Dx())/2, -float64(s.magnetImage.Bounds().Dy())/2)

	rightOp := ebiten.DrawImageOptions{}
	rightOp.GeoM.Concat(magnetOp.GeoM)
	rightOp.GeoM.Rotate(s.magnetSpin)
	rightOp.GeoM.Translate(float64(world.ScreenWidth/2)+float64(titleBounds.Dx())*0.7, float64(world.ScreenHeight/8))

	leftOp := ebiten.DrawImageOptions{}
	leftOp.GeoM.Concat(magnetOp.GeoM)
	leftOp.GeoM.Rotate(-s.magnetSpin)
	leftOp.GeoM.Translate(float64(world.ScreenWidth/2)-float64(titleBounds.Dx())*0.7, float64(world.ScreenHeight/8))

	// Render magnets on each side of title
	screen.DrawImage(s.magnetImage, &leftOp)
	screen.DrawImage(s.magnetImage, &rightOp)

	op := ebiten.DrawImageOptions{}

	// Draw game buttons
	for _, button := range s.buttons {
		button.Draw(screen, &op)
	}

	op.GeoM.Translate(8, 80)
	s.mapList.Draw(screen, &op)
}

func (s *SoloMenuState) StartGame() {
	s.game.SetState(&TravelState{
		game:        s.game,
		targetLevel: s.mapList.selectedMap,
	})
}
