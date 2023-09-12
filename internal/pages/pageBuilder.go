package pages

import (
	"html/template"
)

type Builder interface {
	Build() template.HTML
}
