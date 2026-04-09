package input

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

// virtualPad maps on-screen regions to Actions.
// The game UI layer registers zones by calling RegisterZone.
type virtualPad struct {
	zones []padZone
}

type padZone struct {
	x, y, w, h float64
	action     Action
}

// RegisterZone registers a rectangular screen region (x,y,w,h) as a touch zone
// for the given action. Call once during scene setup.
func RegisterZone(x, y, w, h float64, action Action) {
	vpad.zones = append(vpad.zones, padZone{x: x, y: y, w: w, h: h, action: action})
}

// ClearZones removes all registered virtual pad zones.
func ClearZones() { vpad.zones = vpad.zones[:0] }

func (vp *virtualPad) sample() {
	activeIDs := ebiten.AppendTouchIDs(nil)
	for _, zone := range vp.zones {
		for _, id := range activeIDs {
			tx, ty := ebiten.TouchPosition(id)
			if inZone(float64(tx), float64(ty), zone) {
				state[zone.action] = true
			}
		}
	}
}

func inZone(tx, ty float64, z padZone) bool {
	return tx >= z.x && tx <= z.x+z.w && ty >= z.y && ty <= z.y+z.h
}

// ----------------------------------------------------------------------------
// Joystick (analogue stick simulation via touch drag)
// ----------------------------------------------------------------------------

// Joystick tracks a single-touch analogue stick.
type Joystick struct {
	CenterX, CenterY float64 // screen position of the stick centre
	Radius           float64 // dead zone + max radius in pixels
	activeID         ebiten.TouchID
	hasActive        bool
	DX, DY           float64 // normalised -1..+1 output
}

// Update samples the joystick. Call once per frame.
func (j *Joystick) Update() {
	j.DX, j.DY = 0, 0
	if !j.hasActive {
		// Look for a new touch inside the stick area.
		for _, id := range ebiten.AppendTouchIDs(nil) {
			tx, ty := ebiten.TouchPosition(id)
			dx := float64(tx) - j.CenterX
			dy := float64(ty) - j.CenterY
			if math.Sqrt(dx*dx+dy*dy) <= j.Radius {
				j.activeID = id
				j.hasActive = true
				break
			}
		}
	}
	if j.hasActive {
		if isTouchActive(j.activeID) {
			tx, ty := ebiten.TouchPosition(j.activeID)
			dx := float64(tx) - j.CenterX
			dy := float64(ty) - j.CenterY
			dist := math.Sqrt(dx*dx + dy*dy)
			if dist > j.Radius {
				dx = dx / dist * j.Radius
				dy = dy / dist * j.Radius
				dist = j.Radius
			}
			if dist > 0 {
				j.DX = dx / j.Radius
				j.DY = dy / j.Radius
			}
		} else {
			j.hasActive = false
		}
	}
}

func isTouchActive(id ebiten.TouchID) bool {
	for _, active := range ebiten.AppendTouchIDs(nil) {
		if active == id {
			return true
		}
	}
	return false
}
