// internal/types/types.go

package types

type Recording struct {
	Title string `json:"title"`
	Steps []Step `json:"steps"`
}

type Step struct {
	Type              string `json:"type"`
	Width             int    `json:"width,omitempty"`
	Height            int    `json:"height,omitempty"`
	DeviceScaleFactor int    `json:"deviceScaleFactor,omitempty"`
	IsMobile          bool   `json:"isMobile,omitempty"`
	HasTouch          bool   `json:"hasTouch,omitempty"`
	IsLandscape       bool   `json:"isLandscape,omitempty"`
	URL               string `json:"url,omitempty"`
	AssertedEvents    []struct {
		Type  string `json:"type"`
		URL   string `json:"url"`
		Title string `json:"title"`
	} `json:"assertedEvents,omitempty"`
	Target    string     `json:"target,omitempty"`
	Selectors [][]string `json:"selectors,omitempty"`
	OffsetY   float64    `json:"offsetY,omitempty"`
	OffsetX   float64    `json:"offsetX,omitempty"`
	Value     string     `json:"value,omitempty"`
}
