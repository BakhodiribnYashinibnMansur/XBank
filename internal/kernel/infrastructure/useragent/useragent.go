package useragent

import "strings"

// Info holds parsed User-Agent information.
type Info struct {
	Raw      string `json:"raw"`
	Browser  string `json:"browser,omitempty"`
	OS       string `json:"os,omitempty"`
	Device   string `json:"device,omitempty"`
	IsMobile bool   `json:"is_mobile"`
	IsBot    bool   `json:"is_bot"`
}

// Parse extracts browser, OS, and device info from a User-Agent string.
func Parse(raw string) Info {
	ua := strings.ToLower(raw)
	info := Info{Raw: raw}

	info.Browser = detectBrowser(ua)
	info.OS = detectOS(ua)
	info.Device = detectDevice(ua)
	info.IsMobile = isMobile(ua)
	info.IsBot = isBot(ua)

	return info
}

func detectBrowser(ua string) string {
	switch {
	case strings.Contains(ua, "edg/"):
		return "Edge"
	case strings.Contains(ua, "opr/") || strings.Contains(ua, "opera"):
		return "Opera"
	case strings.Contains(ua, "chrome") && !strings.Contains(ua, "edg/"):
		return "Chrome"
	case strings.Contains(ua, "safari") && !strings.Contains(ua, "chrome"):
		return "Safari"
	case strings.Contains(ua, "firefox"):
		return "Firefox"
	default:
		return "Unknown"
	}
}

func detectOS(ua string) string {
	switch {
	case strings.Contains(ua, "windows"):
		return "Windows"
	case strings.Contains(ua, "mac os x") || strings.Contains(ua, "macintosh"):
		return "macOS"
	case strings.Contains(ua, "android"):
		return "Android"
	case strings.Contains(ua, "iphone") || strings.Contains(ua, "ipad"):
		return "iOS"
	case strings.Contains(ua, "linux"):
		return "Linux"
	default:
		return "Unknown"
	}
}

func detectDevice(ua string) string {
	switch {
	case strings.Contains(ua, "iphone"):
		return "iPhone"
	case strings.Contains(ua, "ipad"):
		return "iPad"
	case strings.Contains(ua, "android") && strings.Contains(ua, "mobile"):
		return "Android Phone"
	case strings.Contains(ua, "android"):
		return "Android Tablet"
	default:
		return "Desktop"
	}
}

func isMobile(ua string) bool {
	return strings.Contains(ua, "mobile") ||
		strings.Contains(ua, "android") ||
		strings.Contains(ua, "iphone")
}

func isBot(ua string) bool {
	bots := []string{"bot", "crawler", "spider", "scraper", "curl", "wget", "httpclient"}
	for _, b := range bots {
		if strings.Contains(ua, b) {
			return true
		}
	}
	return false
}
