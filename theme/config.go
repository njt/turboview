package theme

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"github.com/gdamore/tcell/v2"
)

type Config struct {
	Base      string            `json:"base"`
	Overrides map[string]string `json:"overrides"`
}

func DefaultConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "turboview", "theme.json")
}

func LoadConfig(path string) (*ColorScheme, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("invalid theme config: %w", err)
	}

	base := Get(cfg.Base)
	if base == nil {
		return nil, fmt.Errorf("unknown base theme: %q", cfg.Base)
	}

	scheme := *base

	rv := reflect.ValueOf(&scheme).Elem()
	rt := rv.Type()
	for fieldName, styleStr := range cfg.Overrides {
		found := false
		for i := 0; i < rt.NumField(); i++ {
			if rt.Field(i).Name == fieldName {
				style, err := ParseStyleString(styleStr)
				if err != nil {
					return nil, fmt.Errorf("invalid style for %s: %w", fieldName, err)
				}
				rv.Field(i).Set(reflect.ValueOf(style))
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("unknown color scheme field: %q", fieldName)
		}
	}

	return &scheme, nil
}

func ParseStyleString(s string) (tcell.Style, error) {
	if s == "" {
		return tcell.StyleDefault, nil
	}

	parts := strings.SplitN(s, ":", 3)
	style := tcell.StyleDefault

	if len(parts) >= 1 && parts[0] != "" {
		fg, err := parseHexColor(parts[0])
		if err != nil {
			return style, fmt.Errorf("foreground: %w", err)
		}
		style = style.Foreground(fg)
	}

	if len(parts) >= 2 && parts[1] != "" {
		bg, err := parseHexColor(parts[1])
		if err != nil {
			return style, fmt.Errorf("background: %w", err)
		}
		style = style.Background(bg)
	}

	if len(parts) >= 3 {
		attrs := strings.ToLower(parts[2])
		for _, attr := range strings.Split(attrs, ",") {
			switch strings.TrimSpace(attr) {
			case "bold":
				style = style.Bold(true)
			case "underline":
				style = style.Underline(true)
			case "italic":
				style = style.Italic(true)
			case "reverse":
				style = style.Reverse(true)
			}
		}
	}

	return style, nil
}

func parseHexColor(s string) (tcell.Color, error) {
	s = strings.TrimPrefix(s, "#")
	if len(s) != 6 {
		return tcell.ColorDefault, fmt.Errorf("invalid hex color %q: must be 6 hex digits", "#"+s)
	}
	s = strings.ToLower(s)
	r, err := strconv.ParseUint(s[0:2], 16, 8)
	if err != nil {
		return tcell.ColorDefault, fmt.Errorf("invalid hex color %q: %w", "#"+s, err)
	}
	g, err := strconv.ParseUint(s[2:4], 16, 8)
	if err != nil {
		return tcell.ColorDefault, fmt.Errorf("invalid hex color %q: %w", "#"+s, err)
	}
	b, err := strconv.ParseUint(s[4:6], 16, 8)
	if err != nil {
		return tcell.ColorDefault, fmt.Errorf("invalid hex color %q: %w", "#"+s, err)
	}
	return tcell.NewRGBColor(int32(r), int32(g), int32(b)), nil
}
