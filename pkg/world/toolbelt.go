package world

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/kettek/ebijam22/pkg/data"
)

// Toolbelt is the interface for containing user actions for placing turrets and similar.
type Toolbelt struct {
	items      []*ToolbeltItem
	activeItem *ToolbeltItem
}

// Update updates the toolbelt. This seems a bit silly, but oh well.
func (t *Toolbelt) Update() (request Request) {
	// This is a stupid check.
	if t.activeItem == nil && len(t.items) > 0 {
		t.activeItem = t.items[0]
		t.activeItem.active = true
	}

	// Might as well allow mousewheel for the plebs.
	wheelX, wheelY := ebiten.Wheel()
	if wheelX < 0 || wheelY < 0 {
		t.ScrollItem(-1)
	} else if wheelX > 0 || wheelY > 0 {
		t.ScrollItem(1)
	}

	// Update our individual slots.
	for _, item := range t.items {
		r := item.Update()
		if r != nil {
			switch r.(type) {
			case SelectToolbeltItemRequest:
				// If we're selecting the same item twice, cycle the toolbelt item
				if t.activeItem == item {
					t.activeItem.Cycle()
				} else {
					if t.activeItem != nil {
						t.activeItem.active = false
					}
					t.activeItem = item
					t.activeItem.active = true
				}
			}
			request = r
			break
		}
	}
	return request
}

func (t *Toolbelt) CheckHit(x, y int) bool {
	return false
}

// Position positions the toolbelt and all its tools.
func (t *Toolbelt) Position() {
	toolSlotImage, _ := data.GetImage("toolslot.png")
	x, y := 8, ScreenHeight-8-toolSlotImage.Bounds().Dy()+toolSlotImage.Bounds().Dy()/2

	for _, ti := range t.items {
		ti.Position(&x, &y)
	}
}

func (t *Toolbelt) ActivateItem(item *ToolbeltItem) {
	if t.activeItem != nil {
		t.activeItem.active = false
	}
	t.activeItem = item
	t.activeItem.active = true
}

func (t *Toolbelt) ScrollItem(dir int) {
	for i, item := range t.items {
		if item == t.activeItem {
			i += dir
			if i < 0 {
				i = len(t.items) - 1
			} else if i >= len(t.items) {
				i = 0
			}
			t.ActivateItem(t.items[i])
			break
		}
	}
}

func (t *Toolbelt) Draw(screen *ebiten.Image) {
	// Draw the belt slots.
	for _, ti := range t.items {
		ti.DrawSlot(screen)
	}
	// Then the slot items.
	for _, ti := range t.items {
		ti.Draw(screen)
	}
}

type ToolKind string

const (
	ToolNone    ToolKind = "none"
	ToolGun              = "gun"
	ToolTurret           = "turret"
	ToolWall             = "wall"
	ToolDestroy          = "destroy"
)

// ToolbeltItem is a toolbelt entry.
type ToolbeltItem struct {
	tool        ToolKind
	kind        data.EntityConfig
	polarity    data.Polarity
	x, y        int
	key         ebiten.Key // Key to check against for activation.
	active      bool
	description string
}

func (t *ToolbeltItem) Update() (request Request) {
	toolSlotImage, _ := data.GetImage("toolslot.png")
	// Does the cursor intersect us?
	if t.active && inpututil.IsKeyJustPressed(ebiten.KeyTab) {
		return SelectToolbeltItemRequest{t.tool}
	} else if inpututil.IsKeyJustPressed(t.key) {
		return SelectToolbeltItemRequest{t.tool}
	} else {
		x, y := ebiten.CursorPosition()
		x1, x2 := t.x-toolSlotImage.Bounds().Dx()/2, t.x+toolSlotImage.Bounds().Dx()/2
		y1, y2 := t.y-toolSlotImage.Bounds().Dy()/2, t.y+toolSlotImage.Bounds().Dy()/2

		if x >= x1 && x <= x2 && y >= y1 && y <= y2 {
			if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
				return SelectToolbeltItemRequest{t.tool}
			}
			// Do a dummy return to prevent click through.
			return DummyRequest{}
		}
	}
	return nil
}

// Position assigns the center position for the toolbelt item.
func (t *ToolbeltItem) Position(sx, sy *int) {
	toolSlotImage, _ := data.GetImage("toolslot.png")
	toolDestroyImage, _ := data.GetImage("tool-destroy.png")
	t.x = *sx + toolDestroyImage.Bounds().Dx()/2
	t.y = *sy

	// Move forward our cursor.
	*sx += toolSlotImage.Bounds().Dx() + 1
}

func (t *ToolbeltItem) DrawSlot(screen *ebiten.Image) {
	orbImage, _ := data.GetImage("orb-large.png")
	toolSlotImage, _ := data.GetImage("toolslot.png")
	toolSlotActiveImage, _ := data.GetImage("toolslot-active.png")
	op := ebiten.DrawImageOptions{}
	if t.active {
		op.GeoM.Translate(float64(t.x-toolSlotActiveImage.Bounds().Dx()/2), float64(t.y-toolSlotActiveImage.Bounds().Dy()/2))
		screen.DrawImage(toolSlotActiveImage, &op)

		// Let's draw the title for active item here too
		{
			// Create title label
			toolTitle := t.kind.Title
			if toolTitle == "" {
				toolTitle = string(t.tool)
			}
			label := data.GiveMeString(toolTitle)

			// Create polarity label
			polarity := ""
			if t.tool == ToolGun || t.tool == ToolTurret {
				polarity = "(N) "
				if t.polarity == data.NegativePolarity {
					polarity = "(-) "
				} else if t.polarity == data.PositivePolarity {
					polarity = "(+) "
				}
			}

			// Create cost label
			cost := ""
			if t.tool == ToolTurret {
				config := data.TurretConfigs[t.kind.Title]
				cost = fmt.Sprint(config.Points)
			} else if t.tool == ToolWall {
				cost = "3"
			}

			// Combine labels
			label = fmt.Sprintf("%s %s%s", label, polarity, cost)
			textBounds := text.BoundString(data.NormalFace, label)
			x := t.x - toolSlotActiveImage.Bounds().Dx()/2
			y := t.y - toolSlotActiveImage.Bounds().Dy() + 5
			data.DrawStaticText(
				label,
				data.NormalFace,
				x,
				y,
				color.White,
				screen,
				false,
			)
			x += textBounds.Dx() + 12
			if cost != "" {
				imageOp := ebiten.DrawImageOptions{}
				imageOp.GeoM.Translate(float64(t.x+textBounds.Dx()-5), float64(y-orbImage.Bounds().Dy()))
				screen.DrawImage(orbImage, &imageOp)
				x += orbImage.Bounds().Dx()
			}
			descKey := fmt.Sprintf("desc_%s", toolTitle)
			descTxt := data.GiveMeString(descKey)
			if descKey == descTxt {
				descTxt = t.description
			}
			data.DrawStaticText(descTxt, data.NormalFace, x, y, color.RGBA{255, 255, 255, 128}, screen, false)
		}
	} else {
		op.GeoM.Translate(float64(t.x-toolSlotImage.Bounds().Dx()/2), float64(t.y-toolSlotImage.Bounds().Dy()/2))
		screen.DrawImage(toolSlotImage, &op)
	}
}

func (t *ToolbeltItem) Draw(screen *ebiten.Image) {
	op := ebiten.DrawImageOptions{}

	// Move to the center of our item.
	op.GeoM.Translate(float64(t.x), float64(t.y))

	image := GetToolImage(t.tool, t.kind.Title)

	if image != nil {
		op.ColorM.Scale(data.GetPolarityColorScale(t.polarity))
		op.GeoM.Translate(-float64(image.Bounds().Dx()/2), -float64(image.Bounds().Dy()/2))
		screen.DrawImage(image, &op)
	}
}

// Cycles through available selections for the toolbelt item
func (t *ToolbeltItem) Cycle() {
	switch t.tool {
	// Abuse the fact that polarities have value
	case ToolGun:
		t.polarity++
		if t.polarity > data.PositivePolarity {
			t.polarity = data.NegativePolarity
		}
	case ToolTurret:
		t.polarity *= -1
	}
}

// Retrieves the image for a toolkind
// TODO: perhaps intialize toolbelt items with these instead?
func GetToolImage(t ToolKind, k string) *ebiten.Image {
	var image *ebiten.Image
	switch t {
	case ToolTurret:
		image = data.TurretConfigs[k].HeadImages[0]
	case ToolDestroy:
		image, _ = data.GetImage("tool-destroy.png")
	case ToolGun:
		image, _ = data.GetImage("tool-gun.png")
	case ToolWall:
		image, _ = data.GetImage("wall.png")
	}
	return image
}
