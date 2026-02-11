// internal/core/registry.go
package core

var engines = map[string]GameEngine{}

func Register(engine GameEngine) {
	engines[engine.GameType()] = engine
}

func Get(gameType string) GameEngine {
	return engines[gameType]
}

func All() []GameEngine {
	result := make([]GameEngine, 0, len(engines))
	for _, e := range engines {
		result = append(result, e)
	}
	return result
}
