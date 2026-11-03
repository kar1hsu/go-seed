package model

// Array order is the display order. A page's path identifies a compiled route.
type MenuPage struct {
	Path   string `json:"path"`
	Title  string `json:"title"`
	Icon   string `json:"icon"`
	Hidden bool   `json:"hidden"`
}

type MenuGroup struct {
	ID     string     `json:"id"`
	Title  string     `json:"title"`
	Icon   string     `json:"icon"`
	Hidden bool       `json:"hidden"`
	Pages  []MenuPage `json:"pages"`
}

type MenuLayout struct {
	ID       uint64
	Revision uint64
	Groups   []MenuGroup `gorm:"column:layout;serializer:json"`
}
