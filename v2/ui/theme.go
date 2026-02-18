package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// ModernDarkTheme is a modern dark theme with green accent
type ModernDarkTheme struct{}

var _ fyne.Theme = (*ModernDarkTheme)(nil)

// Colors
var (
	ColorPrimary       = color.RGBA{R: 0, G: 184, B: 148, A: 255}   // #00b894 - Green accent
	ColorPrimaryHover  = color.RGBA{R: 0, G: 210, B: 170, A: 255}   // Lighter green
	ColorPrimaryDark   = color.RGBA{R: 0, G: 150, B: 120, A: 255}   // Darker green
	ColorBackground    = color.RGBA{R: 18, G: 18, B: 24, A: 255}    // #121218 - Very dark blue-black
	ColorSurface       = color.RGBA{R: 28, G: 28, B: 36, A: 255}    // #1c1c24 - Card background
	ColorSurfaceHover  = color.RGBA{R: 38, G: 38, B: 48, A: 255}    // Hover state
	ColorBorder        = color.RGBA{R: 50, G: 50, B: 62, A: 255}    // #32323e - Subtle border
	ColorText          = color.RGBA{R: 245, G: 245, B: 250, A: 255} // Almost white
	ColorTextSecondary = color.RGBA{R: 160, G: 160, B: 175, A: 255} // Muted text
	ColorSuccess       = color.RGBA{R: 0, G: 184, B: 148, A: 255}   // Green
	ColorWarning       = color.RGBA{R: 255, G: 183, B: 77, A: 255}  // Orange
	ColorDanger        = color.RGBA{R: 255, G: 107, B: 107, A: 255} // Red
)

// Color returns the color for the specified name
func (t *ModernDarkTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNamePrimary:
		return ColorPrimary
	case theme.ColorNameBackground:
		return ColorBackground
	case theme.ColorNameButton:
		return ColorPrimary
	case theme.ColorNameDisabledButton:
		return color.RGBA{R: 50, G: 50, B: 60, A: 255}
	case theme.ColorNameDisabled:
		return ColorTextSecondary
	case theme.ColorNameError:
		return ColorDanger
	case theme.ColorNameFocus:
		return ColorPrimary
	case theme.ColorNameForeground:
		return ColorText
	case theme.ColorNameHover:
		return ColorPrimaryHover
	case theme.ColorNameInputBackground:
		return ColorSurface
	case theme.ColorNameInputBorder:
		return ColorBorder
	case theme.ColorNameMenuBackground:
		return ColorSurface
	case theme.ColorNameOverlayBackground:
		return color.RGBA{R: 18, G: 18, B: 24, A: 240}
	case theme.ColorNamePlaceHolder:
		return ColorTextSecondary
	case theme.ColorNamePressed:
		return ColorPrimaryDark
	case theme.ColorNameScrollBar:
		return ColorBorder
	case theme.ColorNameSelection:
		return color.RGBA{R: 0, G: 184, B: 148, A: 80}
	case theme.ColorNameSeparator:
		return ColorBorder
	case theme.ColorNameShadow:
		return color.RGBA{R: 0, G: 0, B: 0, A: 120}
	case theme.ColorNameSuccess:
		return ColorSuccess
	case theme.ColorNameWarning:
		return ColorWarning
	case theme.ColorNameHeaderBackground:
		return ColorSurface
	default:
		return theme.DefaultTheme().Color(name, theme.VariantDark)
	}
}

// Font returns the font for the specified name
func (t *ModernDarkTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

// Icon returns the icon for the specified name
func (t *ModernDarkTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

// Size returns the size for the specified name
func (t *ModernDarkTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNamePadding:
		return 12
	case theme.SizeNameInlineIcon:
		return 24
	case theme.SizeNameText:
		return 14
	case theme.SizeNameHeadingText:
		return 24
	case theme.SizeNameSubHeadingText:
		return 18
	case theme.SizeNameInputBorder:
		return 2
	default:
		return theme.DefaultTheme().Size(name)
	}
}
