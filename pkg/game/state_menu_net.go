package game

import (
	"fmt"
	"image/color"
	"math"
	"strconv"

	"os/user"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/kettek/ebijam22/pkg/data"
	"github.com/kettek/ebijam22/pkg/data/assets/lang"
	"github.com/kettek/ebijam22/pkg/data/ui"
	"github.com/kettek/ebijam22/pkg/net"
	"github.com/kettek/ebijam22/pkg/world"
)

type NetworkMenuState struct {
	game        *Game
	title       string
	magnetImage *ebiten.Image
	magnetSpin  float64
	mapList     MapList

	tiledBackgroundImages  []*ebiten.Image
	tiledBackgroundElapsed int
	tiledBackgroundIndex   int
	backgroundImage        *ebiten.Image

	buttons               []*data.Button
	cancelButton          data.Button
	playerNameInput       *data.TextInput
	remotePlayerNameInput *data.TextInput
	addressInput          *data.TextInput
	portInput             *data.TextInput
	syncRateInput         *data.TextInput
	inputs                []*data.TextInput
	netResult             chan error
	networking            bool
}

func (s *NetworkMenuState) Init() error {
	// Oh boy.
	t, err := data.LoadTileSet("magnet")
	if err != nil {
		return err
	}
	s.tiledBackgroundImages = t.BackgroundImages

	// Load our background image.
	if img, err := data.ReadImage("/ui/multiplayer.png"); err == nil {
		s.backgroundImage = ebiten.NewImageFromImage(img)
	} else {
		return err
	}

	// Generate our maps list.
	if err := s.mapList.Init(); err != nil {
		return err
	}
	if s.game.Options.Map != "" {
		s.mapList.selectedMap = s.game.Options.Map
	}

	// Title Text
	s.title = lang.NetworkGame

	// Set up our network response channel.
	s.netResult = make(chan error)

	// Magnet Image
	if img, err := data.ReadImage("/ui/magnet.png"); err == nil {
		s.magnetImage = ebiten.NewImageFromImage(img)
	} else {
		panic(err)
	}

	centeredX := world.ScreenWidth / 2

	localPlayerX := int(float64(world.ScreenWidth) * 0.15)
	remotePlayerX := int(float64(world.ScreenWidth) * 0.8)

	inputY := int(float64(world.ScreenHeight) * 0.8)
	buttonY := int(float64(world.ScreenHeight) * 0.9)

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
	hostGameButton := data.NewButton(
		centeredX,
		buttonY,
		lang.HostGame,
		func() {
			s.Host()
		},
	)
	hostGameButton.Hover = true

	joinGameButton := data.NewButton(
		centeredX,
		buttonY+hostGameButton.Image().Bounds().Dy()*2,
		lang.JoinGame,
		func() {
			s.JoinByIP()
		},
	)
	joinGameButton.Hover = true
	waitGameButton := data.NewButton(
		localPlayerX,
		buttonY,
		lang.WaitForPlayer,
		func() {
			s.Await()
		},
	)
	waitGameButton.Hover = true
	waitLanGameButton := data.NewButton(
		localPlayerX,
		buttonY+waitGameButton.Image().Bounds().Dy()*2,
		lang.BroadcastGame,
		func() {
			s.AwaitLan(false)
		},
	)
	waitLanGameButton.Hover = true

	findGameButton := data.NewButton(
		remotePlayerX,
		buttonY,
		lang.FindPlayer,
		func() {
			s.Find()
		},
	)
	findGameButton.Hover = true
	findLanGameButton := data.NewButton(
		remotePlayerX,
		buttonY+findGameButton.Image().Bounds().Dy()*2,
		lang.FindBroadcast,
		func() {
			s.AwaitLan(true)
		},
	)
	findLanGameButton.Hover = true

	s.buttons = []*data.Button{
		backButton,
		hostGameButton,
		joinGameButton,
		findGameButton,
		findLanGameButton,
		waitGameButton,
		waitLanGameButton,
	}

	// Standalone cancel button, since it is conditional.
	s.cancelButton = *data.NewButton(
		world.ScreenWidth/2,
		world.ScreenHeight/2,
		lang.Cancel,
		func() {
			s.Cancel()
		},
	)

	playerName := "player"
	if s.game.Options.Name != "" {
		playerName = s.game.Options.Name
	} else {
		user, err := user.Current()
		if err == nil {
			playerName = user.Username
		}
	}
	remoteName := "friendo"
	if s.game.Options.Search != "" {
		remoteName = s.game.Options.Search
	}

	// Create the inputs
	// Player Name Input
	s.playerNameInput = data.NewTextInput(
		lang.LocalPlayerName,
		playerName,
		15,
		localPlayerX,
		inputY,
	)

	// Other player name input
	s.remotePlayerNameInput = data.NewTextInput(
		lang.RemotePlayerName,
		remoteName,
		15,
		remotePlayerX,
		inputY,
	)

	// Sync RateInput
	s.syncRateInput = data.NewTextInput(
		lang.HostSyncRate,
		fmt.Sprintf("%d", s.game.Options.SyncRate),
		15,
		centeredX,
		inputY-40,
	)
	s.syncRateInput.OnChange = func(text string) {
		if rate, err := strconv.Atoi(text); err == nil {
			if rate < 5 {
				rate = 5
			} else if rate >= 5000 {
				rate = 5000
			}
			s.game.Options.SyncRate = rate
		}
	}

	// Address Input
	s.addressInput = data.NewTextInput(
		lang.IPAddress,
		"",
		15,
		centeredX-30, // oops
		inputY,
	)

	// Port Input
	s.portInput = data.NewTextInput(
		lang.Port,
		"20220",
		6,
		centeredX-30+int(float64(s.addressInput.Image().Bounds().Dx())*0.75),
		inputY,
	)

	s.inputs = []*data.TextInput{
		s.playerNameInput,
		s.remotePlayerNameInput,
		s.syncRateInput,
		s.addressInput,
		s.portInput,
	}

	// Create the buttons with handlers
	// Start the tunes
	data.BGM.Set("menu.ogg")
	return nil
}

func (s *NetworkMenuState) Dispose() error {
	return nil
}

func (s *NetworkMenuState) Update() error {
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
	s.magnetSpin += math.Pi / 180 * 4

	if s.networking {
		s.cancelButton.Update()
		// Get that network message if needed.
		select {
		case v := <-s.netResult:
			if v == nil {
				// success!
				s.StartGame()
				go s.game.net.Loop()
				return nil
			} else {
				s.networking = false
				fmt.Println("net error", v)
			}
		default:
		}
	}

	// Update buttons
	for _, button := range s.buttons {
		button.Update()
	}

	// Update inputs
	for i := range s.inputs {
		s.inputs[i].Update()
	}

	s.mapList.Update()

	return nil
}

func (s *NetworkMenuState) Draw(screen *ebiten.Image) {
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
	if s.networking {
		s.cancelButton.Draw(screen, &op)
	}
	// Draw inputs
	for i := range s.inputs {
		s.inputs[i].Draw(screen, &op)
	}

	op.GeoM.Translate(8, 80)
	s.mapList.Draw(screen, &op)
}

func (s *NetworkMenuState) StartGame() {
	s.game.SetState(&TravelState{
		game:        s.game,
		targetLevel: s.mapList.selectedMap,
	})
}

func (s *NetworkMenuState) CreateNet() {
	s.game.net = net.NewConnection(s.playerNameInput.GetInput())
}

func (s *NetworkMenuState) Host() {
	if s.networking {
		return
	}
	s.networking = true
	s.CreateNet()
	go func() {
		err := s.game.net.AwaitDirect(s.addressInput.GetInput()+":"+s.portInput.GetInput(), "")
		s.netResult <- err
	}()
}

func (s *NetworkMenuState) JoinByIP() {
	if s.networking {
		return
	}
	s.networking = true
	s.CreateNet()
	go func() {
		err := s.game.net.AwaitDirect("", s.addressInput.GetInput()+":"+s.portInput.GetInput())
		s.netResult <- err
	}()
}

func (s *NetworkMenuState) Await() {
	if s.networking {
		return
	}
	s.networking = true
	s.CreateNet()
	go func() {
		err := s.game.net.AwaitHandshake(s.game.Options.Handshaker, "", "")
		s.netResult <- err
	}()
}

func (s *NetworkMenuState) AwaitLan(joiner bool) {
	if s.networking {
		return
	}
	s.networking = true
	s.CreateNet()
	go func() {
		err := s.game.net.AwaitLAN(joiner)
		s.netResult <- err
	}()
}

func (s *NetworkMenuState) Find() {
	if s.networking {
		return
	}
	s.networking = true
	s.CreateNet()
	go func() {
		err := s.game.net.AwaitHandshake(s.game.Options.Handshaker, "", s.remotePlayerNameInput.GetInput())
		s.netResult <- err
	}()
}

func (s *NetworkMenuState) Cancel() {
	s.game.net.Close()
}
