package render

// Camera applies a 2D pan + zoom transform from world space to screen space.
type Camera struct {
	X, Y float64  // world position of the camera centre
	Zoom float64  // 1.0 = no zoom
	W, H float64  // screen dimensions (set once at startup)
}

// NewCamera creates a Camera centred on (0,0) with zoom=1.
func NewCamera(screenW, screenH float64) Camera {
	return Camera{Zoom: 1.0, W: screenW, H: screenH}
}

// WorldToScreen converts a world-space point to screen-space pixels.
func (c *Camera) WorldToScreen(wx, wy float64) (float64, float64) {
	if c.Zoom == 0 {
		c.Zoom = 1
	}
	sx := (wx-c.X)*c.Zoom + c.W/2
	sy := (wy-c.Y)*c.Zoom + c.H/2
	return sx, sy
}

// ScreenToWorld converts screen-space pixels back to world-space.
func (c *Camera) ScreenToWorld(sx, sy float64) (float64, float64) {
	if c.Zoom == 0 {
		c.Zoom = 1
	}
	wx := (sx-c.W/2)/c.Zoom + c.X
	wy := (sy-c.H/2)/c.Zoom + c.Y
	return wx, wy
}

// Follow smoothly moves the camera toward (tx, ty) using linear interpolation.
// speed is in world units per second (e.g. 5.0 for a gentle follow).
func (c *Camera) Follow(tx, ty, speed, dt float64) {
	t := speed * dt
	if t > 1 {
		t = 1
	}
	c.X += (tx - c.X) * t
	c.Y += (ty - c.Y) * t
}

// Shake applies a simple positional shake offset. Call with a decaying amplitude
// each frame: amplitude *= 0.9
func (c *Camera) Shake(offsetX, offsetY float64) {
	c.X += offsetX
	c.Y += offsetY
}
