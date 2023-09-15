package pages

import (
	"html/template"
)

// todo: extract from main logic
type Builder interface {
	Build() template.HTML
}
