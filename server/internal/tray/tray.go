// Package tray provides a minimal cross-platform system-tray icon.
package tray

// MenuItem represents a clickable tray context-menu entry.
type MenuItem struct {
	Title     string
	ID        uint32
	clicked   chan struct{}
	ClickedCh <-chan struct{}
	disabled  bool
}

func newMenuItem(id uint32) *MenuItem {
	ch := make(chan struct{}, 1)
	return &MenuItem{
		ID:        id,
		clicked:   ch,
		ClickedCh: ch,
	}
}

// Tray manages the system-tray icon and context menu.
type Tray struct {
	icon    []byte
	tooltip string
	items   []*MenuItem
	nextID  uint32
	quit    chan struct{}
	hwnd    uintptr // hidden window handle (set by Windows Run)
}

// AddMenuItem appends a clickable menu item and returns it.
func (t *Tray) AddMenuItem(title, _ string) *MenuItem {
	t.nextID++
	item := newMenuItem(t.nextID)
	item.Title = title
	t.items = append(t.items, item)
	return item
}

// AddSeparator appends a menu separator.
func (t *Tray) AddSeparator() {}
