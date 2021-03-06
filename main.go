// main
package main

import (
	"log"
	"runtime"

	axe "github.com/JamesDunne/axewitcher"

	"github.com/JamesDunne/golang-openvg/host"
	"github.com/JamesDunne/golang-openvg/vgui"
)

var controller *axe.Controller
var midi axe.Midi

var f *vgui.Font
var ui *vgui.UI

var amps [2]*vgui.PreparedText
var fxNames [5]*vgui.PreparedText
var (
	ptSong      *vgui.PreparedText
	ptAcoustic  *vgui.PreparedText
	ptDirty     *vgui.PreparedText
	ptClean     *vgui.PreparedText
	ptGain      *vgui.PreparedText
	ptVolume    *vgui.PreparedText
	ptGainStr   *vgui.PreparedText
	ptVolumeStr *vgui.PreparedText
)

func initVG(width, height int32) {
	ui = vgui.NewUI()

	ui.Init()
	ui.SetWindow(vgui.NewWindow(0, 0, float32(width), float32(height)))

	amps = [2]*vgui.PreparedText{
		ui.PrepareText("MG"),
		ui.PrepareText("JD"),
	}

	ptSong = ui.PrepareText("Trippin on a Hole in a Paper Heart")
	ptAcoustic = ui.PrepareText("acoustic")
	ptDirty = ui.PrepareText("dirty")
	ptClean = ui.PrepareText("clean")
	ptGain = ui.PrepareText("Gain")
	ptVolume = ui.PrepareText("Volume")
	ptGainStr = ui.PrepareText("0.68")
	ptVolumeStr = ui.PrepareText("0 dB")

	fxNames = [5]*vgui.PreparedText{
		ui.PrepareText("pit1"),
		ui.PrepareText("rtr1"),
		ui.PrepareText("phr1"),
		ui.PrepareText("cho1"),
		ui.PrepareText("dly1"),
	}
}

func drawAmp(w vgui.Window, ampNo int) {
	amp := &controller.Curr.Amp[ampNo]

	ui.StrokeWidth(1.0)
	ui.StrokeColor(ui.Palette(3))
	ui.Pane(w)

	// Amp label at top center:
	label, w := w.SplitH(ui.FontSize() + 8)
	ui.FillColor(ui.Palette(4))
	ui.Text(label, ui.FontSize(), vgui.AlignCenter|vgui.AlignTop, amps[ampNo])

	// Tri-state buttons:
	top, bottom := w.SplitH(ui.FontSize() + 16)
	btnHeight := top.W * 0.33333333
	btnDirty, top := top.SplitV(btnHeight)
	btnClean, btnAcoustic := top.SplitV(btnHeight)

	if t := ui.Button(btnDirty, amp.Mode == axe.AmpDirty, ptDirty); t != nil {

	}
	ui.Button(btnClean, amp.Mode == axe.AmpClean, ptClean)
	ui.Button(btnAcoustic, amp.Mode == axe.AmpAcoustic, ptAcoustic)

	// FX toggles:
	fxWidth := bottom.W / 5.0
	mid, bottom := bottom.SplitH(bottom.H - (ui.FontSize() + 16))
	for i := 0; i < 5; i++ {
		var btnFX vgui.Window
		btnFX, bottom = bottom.SplitV(fxWidth)
		ui.Button(btnFX, amp.Fx[i].Enabled, fxNames[i])
	}

	ui.StrokeColor(ui.Palette(3))
	ui.Pane(mid)

	gain, volume := mid.SplitV(mid.W * 0.5)
	g := float32(amp.DirtyGain) / 127.0
	ui.Dial(gain, ptGain, g, ptGainStr)
	_, _ = gain, g
	v := float32(amp.Volume) / 127.0
	ui.Dial(volume, ptVolume, v, ptVolumeStr)
	_, _ = volume, v
}

// Rendering is done on main thread managed by host:
func drawVG(width, height int32) {
	// Update display window:
	ui.SetWindow(vgui.NewWindow(0, 0, float32(width), float32(height)))
	// Scale font based on window size:
	ui.SetFontSize(14.0 * float32(height) / 480.0)
	w := ui.Window()

	ui.BeginFrame()

	top, bottom := w.SplitH(ui.FontSize() + 8)

	ui.Label(top, ptSong, vgui.AlignLeft|vgui.AlignTop)
	_, _ = top, bottom

	// Split screen for MG v JD:
	mg, jd := bottom.SplitV(bottom.W * 0.5)

	drawAmp(mg, 0)
	drawAmp(jd, 1)
	_, _ = mg, jd

	// Draw touch points:
	for _, tp := range ui.Touches {
		if tp.ID <= 0 {
			continue
		}

		ui.FillColor(vgui.RGBA(255, 255, 255, 160))
		ui.BeginPath()
		ui.Circle(tp.Point, 15.0)
		ui.Fill()
		ui.EndPath()
	}

	ui.EndFrame()
}

func main() {
	// Keep all C calls on this thread:
	runtime.LockOSThread()

	// Create MIDI interface:
	var err error
	midi, err = axe.NewMidi()
	if err != nil {
		log.Println(err)
		// Use null driver:
		midi, err = axe.NewNullMidi()
	}
	defer midi.Close()

	// Initialize controller:
	controller = axe.NewController(midi)
	err = controller.Load()
	if err != nil {
		log.Fatal("Unable to load programs: ", err)
	}
	controller.Init()

	// Supply Go function callbacks:
	host.InitFunc = initVG
	host.DrawFunc = drawVG

	// Initialize the host to display OpenVG graphics at 800x480 resolution:
	if !host.Init(800, 480) {
		return
	}

	// Event polling with idle loop:
	for host.PollEvent() {
		host.Draw()
	}

	host.Destroy()
}
