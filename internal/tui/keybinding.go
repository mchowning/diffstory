package tui

// Keybinding represents a single keyboard shortcut
type Keybinding struct {
	Key         string
	Description string
	Context     string // "global", "navigation", "files"
}

// KeybindingRegistry holds all registered keybindings
type KeybindingRegistry struct {
	bindings []Keybinding
}

// NewKeybindingRegistry creates an empty registry
func NewKeybindingRegistry() *KeybindingRegistry {
	return &KeybindingRegistry{}
}

// Register adds a keybinding to the registry
func (r *KeybindingRegistry) Register(binding Keybinding) {
	r.bindings = append(r.bindings, binding)
}

// GetAll returns all registered keybindings
func (r *KeybindingRegistry) GetAll() []Keybinding {
	return r.bindings
}

// GetByContext returns keybindings for a specific context
func (r *KeybindingRegistry) GetByContext(ctx string) []Keybinding {
	var result []Keybinding
	for _, b := range r.bindings {
		if b.Context == ctx {
			result = append(result, b)
		}
	}
	return result
}
