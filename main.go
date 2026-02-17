package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"

	"github.com/nfnt/resize"

	"github.com/pion/mediadevices"
	"github.com/pion/mediadevices/pkg/driver"
	_ "github.com/pion/mediadevices/pkg/driver/camera"
	"github.com/pion/mediadevices/pkg/prop"
)

// --- Constants & Styles ---

const (
	asciiStandard = " .:-=+*#%@"
	asciiDetailed = " .'`^\",:;Il!i><~+_-?][}{1)(|\\/tfjrxnuvczXYUJCLQ0OZmwqpdbkhao*#MW&8%B@$"
    asciiBlock    = "█"
)

var (
	titleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#D4AF37")).MarginBottom(1)
	statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	errorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000"))
)

// --- Image Processing ---

func imageToAscii(img image.Image, width, height int, chars string, center bool) string {
	if width <= 0 || height <= 0 { return "" }
	imgW := img.Bounds().Dx()
	imgH := img.Bounds().Dy()
	ratio := float64(imgW) / float64(imgH)
	
	termRatio := ratio / 0.5 
	
	finalW := width
	finalH := int(float64(width) / termRatio)
	
	if finalH > height {
		finalH = height
		finalW = int(float64(height) * termRatio)
	}
	
	if finalW <= 0 { finalW = 1 }
	if finalH <= 0 { finalH = 1 }

	resized := resize.Resize(uint(finalW), uint(finalH), img, resize.NearestNeighbor)
	
    bounds := resized.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	
    var sb strings.Builder
    
    for y := 0; y < h; y++ {
		if center {
			padding := (width - w) / 2
			if padding > 0 {
				sb.WriteString(strings.Repeat(" ", padding))
			}
		}
		
        for x := 0; x < w; x++ {
            c := resized.At(x, y)
            r, g, b, _ := c.RGBA()
            gray := (r*299 + g*587 + b*114) / 1000 
            
            idx := int(gray) * len(chars) / 65536
            if idx >= len(chars) { idx = len(chars) - 1 }
			if idx < 0 { idx = 0 }
            sb.WriteByte(chars[idx])
        }
        sb.WriteByte('\n')
    }
    return sb.String()
}

func imageToANSI(img image.Image, width, height int) string {
	if width <= 0 || height <= 0 { return "" }
	
	// Resize logic similar to ASCII but for blocks
	// Blocks are roughly 1:2, same as chars usually.
	imgW := img.Bounds().Dx()
	imgH := img.Bounds().Dy()
	ratio := float64(imgW) / float64(imgH)
	termRatio := ratio / 0.5 
	finalW := width
	finalH := int(float64(width) / termRatio)
	if finalH > height {
		finalH = height
		finalW = int(float64(height) * termRatio)
	}
	if finalW <= 0 { finalW = 1 }
	if finalH <= 0 { finalH = 1 }

	resized := resize.Resize(uint(finalW), uint(finalH), img, resize.NearestNeighbor)
	bounds := resized.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	
    var sb strings.Builder
    
    for y := 0; y < h; y++ {
		padding := (width - w) / 2
		if padding > 0 {
			sb.WriteString(strings.Repeat(" ", padding))
		}
        for x := 0; x < w; x++ {
            c := resized.At(x, y)
            r, g, b, _ := c.RGBA()
            // 8-bit color
            r8, g8, b8 := uint8(r>>8), uint8(g>>8), uint8(b>>8)
            
            // ANSI Truecolor foreground with block char
            fmt.Fprintf(&sb, "\x1b[38;2;%d;%d;%dm█", r8, g8, b8)
        }
        sb.WriteString("\x1b[0m\n")
    }
    	return sb.String()
    }
    
    func imageToStructureAscii(img image.Image, width, height int, center bool) string {
    	if width <= 0 || height <= 0 { return "" }
    	
    	// Resize first
    	imgW := img.Bounds().Dx()
    	imgH := img.Bounds().Dy()
    	ratio := float64(imgW) / float64(imgH)
    	termRatio := ratio / 0.5 
    	finalW := width
    	finalH := int(float64(width) / termRatio)
    	if finalH > height {
    		finalH = height
    		finalW = int(float64(height) * termRatio)
    	}
    	if finalW <= 0 { finalW = 1 }
    	if finalH <= 0 { finalH = 1 }
    
    	resized := resize.Resize(uint(finalW), uint(finalH), img, resize.Bilinear) // Bilinear for smoother gradients
    	bounds := resized.Bounds()
    	w, h := bounds.Dx(), bounds.Dy()
    	
        var sb strings.Builder
        
    	// Simple Sobel-like kernels
    	// Gx = [-1 0 1]
    	// Gy = [-1]
    	//      [ 0]
    	//      [ 1]
    	
        for y := 0; y < h; y++ {
    		if center {
    			padding := (width - w) / 2
    			if padding > 0 {
    				sb.WriteString(strings.Repeat(" ", padding))
    			}
    		}
    		
            for x := 0; x < w; x++ {
    			// Get neighbors for gradient
    			// Clamp coordinates
    			x0 := x-1; if x0 < 0 { x0 = 0 }
    			x2 := x+1; if x2 >= w { x2 = w-1 }
    			y0 := y-1; if y0 < 0 { y0 = 0 }
    			y2 := y+1; if y2 >= h { y2 = h-1 }
    			
    			c_x0 := resized.At(x0, y)
    			c_x2 := resized.At(x2, y)
    			c_y0 := resized.At(x, y0)
    			c_y2 := resized.At(x, y2)
    			
    			// Brightness function
    			lum := func(c color.Color) float64 {
    				r, g, b, _ := c.RGBA()
    				return float64(r*299 + g*587 + b*114) / 65535.0 / 1000.0
    			}
    			
    			gx := lum(c_x2) - lum(c_x0)
    			gy := lum(c_y2) - lum(c_y0)
    			
    			mag := gx*gx + gy*gy
    			threshold := 0.02 // Sensitivity
    			
    			if mag > threshold {
    				// Determine direction
    				// Atan2(y, x) returns radians
    				// We care about the slope of the edge, which is perpendicular to gradient
    				// But for ASCII chars, we just map gradient direction directly to char slope
    				
    				// Using simple ratio for speed
    				absX := gx; if absX < 0 { absX = -absX }
    				absY := gy; if absY < 0 { absY = -absY }
    				
    				if absY > absX * 2.0 {
    					sb.WriteByte('-') // Vertical gradient -> Horizontal line (wait, gradient is change)
    					// Gradient vector points across the edge.
    					// If Gy is large (change in Y is high), we have a horizontal edge.
    					// Wait:
    					// Top is dark, Bottom is light -> Gy is large. Edge is horizontal line "---"
    					// Left is dark, Right is light -> Gx is large. Edge is vertical line "|"
    				} else if absX > absY * 2.0 {
    					sb.WriteByte('|') 
    				} else {
    					// Diagonal
    					// If signs match -> / (roughly)
    					// If signs differ -> \ (roughly)
    					if (gx > 0 && gy > 0) || (gx < 0 && gy < 0) {
    						sb.WriteByte('\\') 
    					} else {
    						sb.WriteByte('/')
    					}
    				}
    			} else {
    				// Low gradient - use standard shading or whitespace
    				// Using standard chars for "texture"
    				val := lum(resized.At(x, y))
    				if val < 0.2 {
    					sb.WriteByte(' ')
    				} else if val < 0.5 {
    					sb.WriteByte('.')
    				} else {
    					sb.WriteByte(':') // Simple texture for non-edges
    				}
    			}
            }
            sb.WriteByte('\n')
        }
        return sb.String()
    }
    
    func textToImage(text string) image.Image {
    
	lines := strings.Split(text, "\n")
	if len(lines) == 0 { return image.NewRGBA(image.Rect(0,0,1,1)) }
	
	// Basic font is 7x13
	charW := 7
	charH := 13
	
	width := 0
	for _, line := range lines {
		if len(line) > width { width = len(line) }
	}
	width *= charW
	height := len(lines) * charH
	
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	// Fill black
	for i := 0; i < len(img.Pix); i+=4 {
		img.Pix[i] = 0
		img.Pix[i+1] = 0
		img.Pix[i+2] = 0
		img.Pix[i+3] = 255
	}
	
	d := &font.Drawer{
		Dst:  img,
		Src:  image.White,
		Face: basicfont.Face7x13,
		Dot:  fixed.P(0, charH),
	}
	
	for _, line := range lines {
		d.Dot.X = 0
		d.DrawString(line)
		d.Dot.Y += fixed.I(charH)
	}
	
	return img
}

func applyFilter(img image.Image, f Filter) image.Image {
	if f == FilterNone { return img }
	
	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	newImg := image.NewRGBA(bounds)
	
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			c := img.At(x, y)
			r, g, b, a := c.RGBA()
			
			switch f {
			case FilterGrayscale:
				gray := uint8((r*299 + g*587 + b*114) / 256000)
				newImg.Set(x, y, color.RGBA{gray, gray, gray, uint8(a >> 8)})
			case FilterInvert:
				newImg.Set(x, y, color.RGBA{uint8(255 - r>>8), uint8(255 - g>>8), uint8(255 - b>>8), uint8(a >> 8)})
			case FilterSepia:
				rr := float64(r>>8)
				gg := float64(g>>8)
				bb := float64(b>>8)
				tr := 0.393*rr + 0.769*gg + 0.189*bb
				tg := 0.349*rr + 0.686*gg + 0.168*bb
				tb := 0.272*rr + 0.534*gg + 0.131*bb
				if tr > 255 { tr = 255 }
				if tg > 255 { tg = 255 }
				if tb > 255 { tb = 255 }
				newImg.Set(x, y, color.RGBA{uint8(tr), uint8(tg), uint8(tb), uint8(a >> 8)})
			case FilterRed:
				newImg.Set(x, y, color.RGBA{uint8(r >> 8), 0, 0, uint8(a >> 8)})
			case FilterGreen:
				newImg.Set(x, y, color.RGBA{0, uint8(g >> 8), 0, uint8(a >> 8)})
			case FilterBlue:
				newImg.Set(x, y, color.RGBA{0, 0, uint8(b >> 8), uint8(a >> 8)})
			}
		}
	}
	return newImg
}

// --- Types ---

type Mode int
const (
    ModeASCII Mode = iota
    ModeDetailed
    ModeColor
    ModeStructure
)
func (m Mode) String() string {
	switch m {
	case ModeASCII: return "Standard ASCII"
	case ModeDetailed: return "High Detail ASCII"
	case ModeColor: return "Color (Normal)"
	case ModeStructure: return "Structure (Edge)"
	default: return "Unknown"
	}
}

type Filter int
const (
    FilterNone Filter = iota
    FilterGrayscale
    FilterInvert
    FilterSepia
	FilterRed
	FilterGreen
	FilterBlue
)
func (f Filter) String() string {
	switch f {
	case FilterNone: return "None"
	case FilterGrayscale: return "Grayscale"
	case FilterInvert: return "Invert"
	case FilterSepia: return "Sepia"
	case FilterRed: return "Red Tint"
	case FilterGreen: return "Green Tint"
	case FilterBlue: return "Blue Tint"
	default: return "Unknown"
	}
}

// --- Messages ---

type frameMsg image.Image
type errorMsg error
type statusMsg string
type clearStatusMsg struct{}

type cameraReadyMsg struct {
	stream mediadevices.MediaStream
	reader VideoReader
	driverID string
}

type VideoReader interface {
	Read() (image.Image, func(), error)
}

// --- Model ---

type model struct {
    width, height int
    
    stream      mediadevices.MediaStream
    reader      VideoReader
    
    currentFrame image.Image
    
    mode        Mode
    filter      Filter
    
    statusText  string
    statusTimer *time.Timer
    
    help        help.Model
    keys        keyMap
	showHelp    bool
	
	recording   bool
	recFrames   []image.Image
	recStart    time.Time
	
	devices     []mediadevices.MediaDeviceInfo
	currentDev  int
	
	err error
}

type keyMap struct {
    Snap   key.Binding
    Switch key.Binding
    Filter key.Binding
    Mode   key.Binding
    Help   key.Binding
    Record key.Binding
    Quit   key.Binding
}

var keys = keyMap{
    Snap:   key.NewBinding(key.WithKeys(" "), key.WithHelp("space", "save photo")),
    Switch: key.NewBinding(key.WithKeys("c"), key.WithHelp("c", "next camera")),
    Filter: key.NewBinding(key.WithKeys("f"), key.WithHelp("f", "cycle filter")),
    Mode:   key.NewBinding(key.WithKeys("m"), key.WithHelp("m", "toggle mode")),
    Help:   key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "toggle help")),
    Record: key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "record gif")),
    Quit:   key.NewBinding(key.WithKeys("q", "esc", "ctrl+c"), key.WithHelp("q", "quit")),
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Snap, k.Record, k.Mode, k.Filter, k.Switch, k.Help, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Snap, k.Record, k.Mode},
		{k.Filter, k.Switch, k.Help, k.Quit},
	}
}

// --- Init & Update ---

func initialModel() model {
	devices := mediadevices.EnumerateDevices()
	var videoDevs []mediadevices.MediaDeviceInfo
	for _, d := range devices {
		if d.DeviceType == driver.Camera {
			videoDevs = append(videoDevs, d)
		}
	}

	h := help.New()
	h.ShowAll = true // Always show full help when visible

	return model{
		mode: ModeASCII,
		filter: FilterNone,
		help: h,
		keys: keys,
		statusText: "Initializing...",
		devices: videoDevs,
	}
}

func (m model) Init() tea.Cmd {
    return tea.Batch(
		initCameraCmd,
		tea.EnterAltScreen,
	)
}

func initCameraCmd() tea.Msg {
	s, err := mediadevices.GetUserMedia(mediadevices.MediaStreamConstraints{
		Video: func(c *mediadevices.MediaTrackConstraints) {
			// Relaxed constraints to find any matching driver
		},
	})
	
	if err != nil {
		return errorMsg(fmt.Errorf("failed to open camera: %w", err))
	}
	
	if len(s.GetVideoTracks()) == 0 {
		for _, t := range s.GetTracks() { t.Close() }
		return errorMsg(fmt.Errorf("no video tracks found"))
	}
	
	track := s.GetVideoTracks()[0]
	videoTrack := track.(*mediadevices.VideoTrack)
	reader := videoTrack.NewReader(false)
	
	return cameraReadyMsg{stream: s, reader: reader}
}

func switchCameraCmd(driverID string) tea.Cmd {
	return func() tea.Msg {
		s, err := mediadevices.GetUserMedia(mediadevices.MediaStreamConstraints{
			Video: func(c *mediadevices.MediaTrackConstraints) {
				c.DeviceID = prop.String(driverID)
			},
		})
		
		if err != nil { return errorMsg(err) }
		
		if len(s.GetVideoTracks()) == 0 {
			for _, t := range s.GetTracks() { t.Close() }
			return errorMsg(fmt.Errorf("no video tracks found on new device"))
		}
		
		track := s.GetVideoTracks()[0]
		videoTrack := track.(*mediadevices.VideoTrack)
		reader := videoTrack.NewReader(false)
		
		return cameraReadyMsg{stream: s, reader: reader, driverID: driverID}
	}
}

func readFrameCmd(reader VideoReader) tea.Cmd {
	return func() tea.Msg {
		frame, release, err := reader.Read()
		if err != nil {
			return errorMsg(err)
		}
		
		bounds := frame.Bounds()
		clone := image.NewRGBA(bounds)
		
		for y := 0; y < bounds.Dy(); y++ {
			for x := 0; x < bounds.Dx(); x++ {
				clone.Set(x, y, frame.At(x, y))
			}
		}
		
		release()
		return frameMsg(clone)
	}
}

func (m model) savePhoto() tea.Cmd {
	if m.currentFrame == nil {
		return func() tea.Msg { return statusMsg("No frame to save!") }
	}
	
	// Capture the current frame in a closure to avoid race if m.currentFrame changes
	frameToSave := m.currentFrame
	currentFilter := m.filter
	currentMode := m.mode
	// Capture dimensions for ASCII text generation
	w, h := m.width, m.height
	
	return func() tea.Msg {
		home, _ := os.UserHomeDir()
		dir := filepath.Join(home, "Pictures", "AtlasCam")
		if err := os.MkdirAll(dir, 0755); err != nil {
			return errorMsg(err)
		}
		
		timestamp := time.Now().Unix()
		name := fmt.Sprintf("atlas_cam_%d", timestamp)
		
		// 1. Save High-Res Image (JPG)
		// If in Color Mode, we save the High Res filtered image.
		// If in ASCII Mode, we save the RENDERED ASCII IMAGE.
		
		pathJPG := filepath.Join(dir, name+".jpg")
		f, err := os.Create(pathJPG)
		if err != nil { return errorMsg(err) }
		defer f.Close()

		filteredFrame := applyFilter(frameToSave, currentFilter)
		
		var finalImage image.Image
		finalImage = filteredFrame
		
		if currentMode == ModeASCII || currentMode == ModeDetailed || currentMode == ModeStructure {
			chars := asciiStandard
			if currentMode == ModeDetailed { chars = asciiDetailed }
			
			var txt string
			if currentMode == ModeStructure {
				txt = imageToStructureAscii(filteredFrame, w, h-4, false)
			} else {
				txt = imageToAscii(filteredFrame, w, h-4, chars, false)
			}
			
			// Convert that text to an image
			finalImage = textToImage(txt)
			
			// Also save the text file since we have it
			pathTXT := filepath.Join(dir, name+".txt")
			os.WriteFile(pathTXT, []byte(txt), 0644)
		}

		if err := jpeg.Encode(f, finalImage, nil); err != nil {
			return errorMsg(err)
		}
		
		return statusMsg("Saved " + name + ".jpg")
	}
}

func (m model) saveVideo(frames []image.Image) tea.Cmd {
	if len(frames) == 0 { return nil }
	
	return func() tea.Msg {
		home, _ := os.UserHomeDir()
		dir := filepath.Join(home, "Pictures", "AtlasCam")
		if err := os.MkdirAll(dir, 0755); err != nil { return errorMsg(err) }
		
		name := fmt.Sprintf("atlas_cam_clip_%d.gif", time.Now().Unix())
		path := filepath.Join(dir, name)
		f, err := os.Create(path)
		if err != nil { return errorMsg(err) }
		defer f.Close()
		
		// Convert frames to Paletted for GIF
		// This is slow, so we do it here in the goroutine
		outGIF := &gif.GIF{}
		
		// Use a standard palette. 
		// For ASCII (BW), we can use a small palette.
		// For Color, we need Plan9 or similar.
		// Let's use Plan9 for safety.
		// opts := gif.Options{NumColors: 256, Drawer: draw.FloydSteinberg} // Unused
		
		for _, src := range frames {
			bounds := src.Bounds()
			paletted := image.NewPaletted(bounds, 	color.Palette{
				color.Black, color.White, color.RGBA{255,0,0,255}, color.RGBA{0,255,0,255}, color.RGBA{0,0,255,255},
				// Add grays
				color.Gray{0x33}, color.Gray{0x66}, color.Gray{0x99}, color.Gray{0xCC},
			})
			
			// Actually, just use standard quantizer if possible, but manual palette is faster?
			// Let's use a simple quantization
			// Draw into paletted image
			draw.Draw(paletted, bounds, src, bounds.Min, draw.Src)
			
			outGIF.Image = append(outGIF.Image, paletted)
			outGIF.Delay = append(outGIF.Delay, 4) // 4x10ms = 40ms ~ 25fps
		}
		
		// EncodeAll does not take options? Wait.
		// gif.EncodeAll signature is (w io.Writer, g *GIF) error
		// Options are only for Encode (single frame) or EncodeAll does not support them?
		// Ah, EncodeAll just encodes the GIF struct. The Quantizer is used during paletted creation.
		// My mistake. The 'opts' variable I created earlier (with NumColors) is for 'draw.Quantizer' interface usage 
		// if we were using a Quantizer. But here I am manually creating Paletted image.
		// 
		// Actually, standard library image/draw doesn't have a simple "Quantize to this palette" function
		// except draw.Draw which just finds nearest color.
		
		// So 'opts' is unused. I should remove it.
		
		gif.EncodeAll(f, outGIF)
		
		return statusMsg("Saved GIF: " + name)
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		
	case cameraReadyMsg:
		// Clean up old stream if it exists
		if m.stream != nil && m.stream != msg.stream {
			for _, t := range m.stream.GetTracks() { t.Close() }
		}
		m.stream = msg.stream
		m.reader = msg.reader
		m.statusText = "Camera Ready"
		if msg.driverID != "" {
			m.statusText += fmt.Sprintf(" (%s)", msg.driverID)
		}
		return m, readFrameCmd(m.reader)
		
	case frameMsg:
		m.currentFrame = image.Image(msg)
		
		// Recording Logic
		if m.recording {
			// Process frame EXACTLY as we do for saving/viewing
			// This duplicates some logic but ensures consistency
			filtered := applyFilter(m.currentFrame, m.filter)
			
			var frameToRec image.Image
			
			if m.mode == ModeASCII || m.mode == ModeDetailed || m.mode == ModeStructure {
				chars := asciiStandard
				if m.mode == ModeDetailed { chars = asciiDetailed }
				
				var txt string
				// No margin for video
				if m.mode == ModeStructure {
					txt = imageToStructureAscii(filtered, m.width, m.height-4, false)
				} else {
					txt = imageToAscii(filtered, m.width, m.height-4, chars, false)
				}
				frameToRec = textToImage(txt)
			} else {
				frameToRec = filtered
			}
			
			// Append copy? textToImage creates new, applyFilter creates new.
			// m.currentFrame is reused? No, readFrameCmd creates new copy.
			// So we are safe to just append.
			m.recFrames = append(m.recFrames, frameToRec)
		}
		
		return m, readFrameCmd(m.reader) // Loop
		
	case errorMsg:
		m.err = msg
		m.statusText = "Error: " + msg.Error()
		return m, nil
		
	case statusMsg:
		m.statusText = string(msg)
		// Clear status after 3 seconds
		return m, tea.Tick(3*time.Second, func(_ time.Time) tea.Msg {
			return clearStatusMsg{}
		})

	case clearStatusMsg:
		m.statusText = ""
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Quit):
			if m.stream != nil {
				for _, t := range m.stream.GetTracks() { t.Close() }
			}
			return m, tea.Quit
			
		case key.Matches(msg, m.keys.Record):
			m.recording = !m.recording
			if m.recording {
				m.recFrames = []image.Image{}
				m.recStart = time.Now()
				m.statusText = "Recording..."
			} else {
				m.statusText = fmt.Sprintf("Encoding %d frames...", len(m.recFrames))
				return m, m.saveVideo(m.recFrames)
			}
			
		case key.Matches(msg, m.keys.Snap):
			return m, m.savePhoto()
			
		case key.Matches(msg, m.keys.Mode):
			m.mode = (m.mode + 1) % 4
			
		case key.Matches(msg, m.keys.Filter):
			m.filter = (m.filter + 1) % 7
		
		case key.Matches(msg, m.keys.Help):
			m.showHelp = !m.showHelp
			
		case key.Matches(msg, m.keys.Switch):
			if len(m.devices) > 1 {
				m.currentDev = (m.currentDev + 1) % len(m.devices)
				dev := m.devices[m.currentDev]
				
				m.statusText = "Switching to " + dev.Label
				return m, switchCameraCmd(dev.DeviceID)
			} else {
				m.statusText = "No other cameras found"
				return m, nil
			}
		}
	}
	return m, nil
}

func (m model) View() string {
	if m.err != nil {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, errorStyle.Render(m.err.Error()))
	}
	
	if m.currentFrame == nil {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, "Waiting for camera... (" + m.statusText + ")")
	}
	
	// Process Frame
	filtered := applyFilter(m.currentFrame, m.filter)
	
	// Render ASCII or ANSI
	var art string
	
	// Header/Footer allowance
	h := m.height - 4
	if h < 1 { h = 1 }

	switch m.mode {
	case ModeColor:
		art = imageToANSI(filtered, m.width, h)
	case ModeDetailed:
		art = imageToAscii(filtered, m.width, h, asciiDetailed, true)
	case ModeStructure:
		art = imageToStructureAscii(filtered, m.width, h, true)
	default:
		art = imageToAscii(filtered, m.width, h, asciiStandard, true)
	}
	
	// UI Layout
	title := "ATLAS CAM"
	if m.recording {
		title += fmt.Sprintf(" [REC %ds]", int(time.Since(m.recStart).Seconds()))
		titleStyle = titleStyle.Foreground(lipgloss.Color("#FF0000"))
	} else {
		titleStyle = titleStyle.Foreground(lipgloss.Color("#D4AF37"))
	}
	
	header := titleStyle.Render(title)
	
	var footer string
	if m.showHelp {
		footer = m.help.View(m.keys)
	} else {
		footer = statusStyle.Render(fmt.Sprintf("%s | %s | %s | Press '?' for help", m.mode, m.filter, m.statusText))
	}
	
	return lipgloss.JoinVertical(lipgloss.Center,
		header,
		art,
		footer,
	)
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
