package sitegen

import (
	_ "embed"
)

//go:embed assets/style.css
var styleCSS string

//go:embed assets/script.js
var scriptJS string
