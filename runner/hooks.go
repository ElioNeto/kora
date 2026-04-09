package runner

// hooks storage — kept in a separate file to avoid cluttering game.go.
// The Game struct in game.go is extended by declaring these fields here;
// because Go does not allow split struct declarations across files,
// we instead store hooks in a companion map keyed by *Game.
// This is the idiomatic Go approach for optional extensibility.

var (
	updateHooks = map[*Game][]Hook{}
	drawHooks   = map[*Game][]Hook{}
)

// addUpdateHook stores a hook for g.
func (g *Game) addUpdateHook(fn Hook) {
	updateHooks[g] = append(updateHooks[g], fn)
}

// addDrawHook stores a draw hook for g.
func (g *Game) addDrawHook(fn Hook) {
	drawHooks[g] = append(drawHooks[g], fn)
}

// runUpdateHooks executes all registered update hooks for g.
func (g *Game) runUpdateHooks(dt float64) {
	for _, fn := range updateHooks[g] {
		fn(dt)
	}
}

// runDrawHooks executes all registered draw hooks for g.
func (g *Game) runDrawHooks(dt float64) {
	for _, fn := range drawHooks[g] {
		fn(dt)
	}
}
