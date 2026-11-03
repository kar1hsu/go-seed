package resources

import (
	_ "embed"
	"encoding/json"
)

//go:embed menus.generated.json
var menuCatalog []byte

type MenuRoute struct {
	Path, Title, Icon, Group, Permission string
	Hidden                               bool
}

func MenuRoutes() ([]MenuRoute, error) {
	var routes []MenuRoute
	err := json.Unmarshal(menuCatalog, &routes)
	return routes, err
}
