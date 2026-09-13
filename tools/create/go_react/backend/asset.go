//go:build ignore

package __PACKAGE_NAME__

import "embed"

// Dist is the built React app. Must live at module root: //go:embed cannot use "..".
//
//go:embed __PROJECT_NAME__-react/dist
var Dist embed.FS

//go:embed __PROJECT_NAME__-react/template.html
var Template string
