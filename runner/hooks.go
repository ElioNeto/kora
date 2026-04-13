package runner

// Hook is a function called every tick/frame.
type Hook func(dt float64)

// AddUpdateHook registers fn to be called at the start of every Update tick.
// Useful for global systems (particle emitters, achievement trackers, etc.).
func (g *Game) AddUpdateHook(fn Hook) {
	g.updateHooks = append(g.updateHooks, fn)
}

// AddDrawHook registers fn to be called at the end of every Draw frame.
func (g *Game) AddDrawHook(fn Hook) {
	g.drawHooks = append(g.drawHooks, fn)
}
