package tarot

import (
	"image/color"

	dreams "github.com/dReam-dApps/dReams"
	"github.com/dReam-dApps/dReams/bundle"
	"github.com/dReam-dApps/dReams/dwidget"
	"github.com/dReam-dApps/dReams/rpc"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

var T dreams.ContainerStack

// Layout all objects for Iluma Tarot dApp
func LayoutAllItems(d *dreams.AppObject) fyne.CanvasObject {
	T.LeftLabel = widget.NewLabel("")
	T.RightLabel = widget.NewLabel("")
	T.LeftLabel.SetText("Total Readings: " + Iluma.Value.Readings + "      Click your card for Iluma reading")
	T.RightLabel.SetText("dReams Balance: " + rpc.DisplayBalance("dReams") + "      Dero Balance: " + rpc.DisplayBalance("Dero") + "      Height: " + rpc.Wallet.Display.Height)

	search_entry := widget.NewEntry()
	search_entry.SetPlaceHolder("TXID:")
	search_button := widget.NewButton("    Search   ", func() {
		txid := search_entry.Text
		if len(txid) == 64 {
			signer := rpc.VerifySigner(txid)
			if signer {
				Iluma.Value.Display = true
				Iluma.Label.SetText("")
				FetchReading(txid)
				if Iluma.Value.Card2 != 0 && Iluma.Value.Card3 != 0 {
					Iluma.Card1.Objects[1] = DisplayCard(Iluma.Value.Card1)
					Iluma.Card2.Objects[1] = DisplayCard(Iluma.Value.Card2)
					Iluma.Card3.Objects[1] = DisplayCard(Iluma.Value.Card3)
					Iluma.Value.Num = 3
				} else {
					Iluma.Card1.Objects[1] = DisplayCard(0)
					Iluma.Card2.Objects[1] = DisplayCard(Iluma.Value.Card1)
					Iluma.Card3.Objects[1] = DisplayCard(0)
					Iluma.Value.Num = 1
				}
				Iluma.Box.Refresh()
			} else {
				logger.Errorln("[Iluma] This is not your reading")
			}
		}
	})

	tarot_label := container.NewHBox(T.LeftLabel, layout.NewSpacer(), T.RightLabel)

	//  Clickable Tarot card objects
	Iluma.Label = widget.NewLabel("")
	Iluma.Label.Alignment = fyne.TextAlignCenter
	one := widget.NewButton("", func() {
		if Iluma.Value.Num == 3 && !Iluma.Open && Iluma.Value.Card1 > 0 {
			c := Iluma.Value.Card1
			reset := Iluma.Card1
			Iluma.Card1 = *ilumaDialog(1, ilumaDescription(c), reset)
		}
	})

	card_back := defaultCard

	spacer := canvas.NewRectangle(color.RGBA{0, 0, 0, 0})
	spacer.SetMinSize(fyne.NewSize(40, 0))

	Iluma.Card1 = *container.NewStack(one, card_back)
	pad1 := container.NewBorder(nil, nil, spacer, spacer, &Iluma.Card1)

	two := widget.NewButton("", func() {
		if !Iluma.Open {
			reset := Iluma.Card2
			if Iluma.Value.Num == 3 && Iluma.Value.Card2 > 0 {
				c := Iluma.Value.Card2
				Iluma.Card2 = *ilumaDialog(2, ilumaDescription(c), reset)
			}

			if Iluma.Value.Num == 1 && Iluma.Value.Card1 > 0 {
				c := Iluma.Value.Card1
				Iluma.Card2 = *ilumaDialog(2, ilumaDescription(c), reset)
			}
		}
	})

	Iluma.Card2 = *container.NewStack(two, card_back)
	pad2 := container.NewBorder(nil, nil, spacer, spacer, &Iluma.Card2)

	three := widget.NewButton("", func() {
		if Iluma.Value.Num == 3 && !Iluma.Open && Iluma.Value.Card3 > 0 {
			c := Iluma.Value.Card3
			reset := Iluma.Card3
			Iluma.Card3 = *ilumaDialog(3, ilumaDescription(c), reset)
		}
	})

	one.Importance = widget.LowImportance
	two.Importance = widget.LowImportance
	three.Importance = widget.LowImportance

	Iluma.Card3 = *container.NewStack(three, card_back)
	pad3 := container.NewBorder(nil, nil, spacer, spacer, &Iluma.Card3)

	actions := container.NewAdaptiveGrid(3,
		layout.NewSpacer(),
		Iluma.Label,
		layout.NewSpacer())

	Iluma.Box = container.NewAdaptiveGrid(3,
		pad1,
		pad2,
		pad3)

	pad := container.NewBorder(
		nil,
		nil,
		layout.NewSpacer(),
		layout.NewSpacer(),
		Iluma.Box)

	alpha150 := canvas.NewRectangle(color.RGBA{0, 0, 0, 150})
	if bundle.AppColor == color.White {
		alpha150 = canvas.NewRectangle(color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0x55})
	}

	card_box := container.NewBorder(
		nil,
		actions,
		nil,
		nil,
		pad)

	T.DApp = container.NewBorder(
		dwidget.LabelColor(tarot_label),
		nil,
		nil,
		nil,
		container.NewStack(alpha150, card_box))

	Iluma.Draw1 = widget.NewButton("Draw One", func() {
		if !Iluma.Open {
			Iluma.Draw1.Hide()
			Iluma.Draw3.Hide()
			drawConfirm(1, d)
		}
	})
	Iluma.Draw1.Importance = widget.HighImportance

	Iluma.Draw3 = widget.NewButton("Draw Three", func() {
		if !Iluma.Open {
			Iluma.Draw1.Hide()
			Iluma.Draw3.Hide()
			drawConfirm(3, d)
		}
	})
	Iluma.Draw3.Importance = widget.HighImportance

	Iluma.Draw1.Hide()
	Iluma.Draw3.Hide()

	draw_cont := container.NewAdaptiveGrid(5,
		layout.NewSpacer(),
		layout.NewSpacer(),
		Iluma.Draw1,
		Iluma.Draw3,
		layout.NewSpacer())

	Iluma.Search = container.NewBorder(nil, nil, nil, search_button, search_entry)

	Iluma.Actions = container.NewVBox(
		layout.NewSpacer(),
		container.NewAdaptiveGrid(2, draw_cont, Iluma.Search))

	Iluma.Search.Hide()
	Iluma.Actions.Hide()

	// Iluma tab objects, intro description and image scroll
	var display int
	var first, second, third bool
	img := canvas.NewImageFromResource(resourceIluma82Png)
	intro := widget.NewLabel(iluma_intro)
	scroll := container.NewScroll(intro)

	alpha120 := canvas.NewRectangle(color.RGBA{0, 0, 0, 120})
	if bundle.AppColor == color.White {
		alpha120 = canvas.NewRectangle(color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0x55})
	}

	iluma_cont := container.NewGridWithColumns(2, scroll, img)

	scroll.OnScrolled = func(p fyne.Position) {
		if p.Y <= 400 {
			second = false
			third = false
			display = 1
		} else if p.Y >= 400 && p.Y <= 800 {
			first = false
			third = false
			display = 2
		} else if p.Y >= 800 {
			first = false
			second = false
			display = 3
		}

		switch display {
		case 1:
			if !first {
				iluma_cont.Objects[1] = canvas.NewImageFromResource(resourceIluma82Png)
				iluma_cont.Refresh()
				first = true
			}
		case 2:
			if !second {
				iluma_cont.Objects[1] = canvas.NewImageFromResource(resourceIluma80Png)
				iluma_cont.Refresh()
				second = true
			}
		case 3:
			if !third {
				iluma_cont.Objects[1] = canvas.NewImageFromResource(resourceIluma83Png)
				iluma_cont.Refresh()
				third = true
			}
		default:

		}
	}

	tarot_tabs := container.NewAppTabs(
		container.NewTabItem("Iluma", container.NewStack(alpha120, iluma_cont)),
		container.NewTabItem("Reading", T.DApp))

	tarot_tabs.OnSelected = func(ti *container.TabItem) {
		switch ti.Text {
		case "Iluma":
			Iluma.Actions.Hide()
		case "Reading":
			Iluma.Actions.Show()
		default:

		}
	}

	tarot_tabs.SetTabLocation(container.TabLocationBottom)

	go fetch(d)

	return container.NewStack(tarot_tabs, Iluma.Actions)
}
