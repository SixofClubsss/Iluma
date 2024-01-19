package tarot

import (
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"github.com/blang/semver/v4"
	dreams "github.com/dReam-dApps/dReams"
	"github.com/dReam-dApps/dReams/bundle"
	"github.com/dReam-dApps/dReams/dwidget"
	"github.com/dReam-dApps/dReams/gnomes"
	"github.com/dReam-dApps/dReams/menu"
	"github.com/dReam-dApps/dReams/rpc"
	"github.com/sirupsen/logrus"
)

const app_tag = "Iluma"

var version = semver.MustParse("0.3.0-dev.3")

// Check tarot package version
func Version() semver.Version {
	return version
}

// Run Iluma as a single dApp
func StartApp() {
	n := runtime.NumCPU()
	runtime.GOMAXPROCS(n)
	gnomes.InitLogrusLog(logrus.InfoLevel)
	config := menu.ReadDreamsConfig(app_tag)
	gnomon := gnomes.NewGnomes()

	// Initialize Fyne app and window
	a := app.NewWithID(fmt.Sprintf("%s Tarot Client", app_tag))
	a.Settings().SetTheme(bundle.DeroTheme(config.Skin))
	w := a.NewWindow(app_tag)
	w.SetIcon(resourceIlumaIconJpg)
	w.Resize(fyne.NewSize(1400, 800))
	w.SetMaster()
	done := make(chan struct{})

	// Initialize dReams AppObject and close func
	menu.Theme.Img = *canvas.NewImageFromResource(menu.DefaultThemeResource())
	d := dreams.AppObject{
		App:        a,
		Window:     w,
		Background: container.NewStack(&menu.Theme.Img),
	}
	d.SetChannels(1)

	closeFunc := func() {
		save := dreams.SaveData{
			Skin:   config.Skin,
			DBtype: gnomon.DBStorageType(),
			Theme:  menu.Theme.Name,
		}

		if rpc.Daemon.Rpc == "" {
			save.Daemon = config.Daemon
		} else {
			save.Daemon = []string{rpc.Daemon.Rpc}
		}

		menu.WriteDreamsConfig(save)
		gnomon.Stop(app_tag)
		d.StopProcess()
		w.Close()
	}

	w.SetCloseIntercept(closeFunc)

	// Handle ctrl-c close
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println()
		closeFunc()
	}()

	// Stand alone process
	go func() {
		time.Sleep(3 * time.Second)
		ticker := time.NewTicker(3 * time.Second)
		for {
			select {
			case <-ticker.C:
				rpc.Ping()
				rpc.EchoWallet(app_tag)
				go rpc.GetDreamsBalances(rpc.SCIDs)
				rpc.GetWalletHeight(app_tag)

				if rpc.Daemon.IsConnected() {
					rpc.Startup = false
				}

				d.SignalChannel()

			case <-d.Closing():
				logger.Printf("[%s] Closing...", app_tag)
				ticker.Stop()
				d.CloseAllDapps()
				time.Sleep(time.Second)
				done <- struct{}{}
				return
			}
		}
	}()

	// Create dwidget connection box with controls
	connect_box := dwidget.NewHorizontalEntries(app_tag, 1)
	connect_box.Button.OnTapped = func() {
		rpc.GetAddress(app_tag)
		rpc.Ping()
	}

	connect_box.AddDaemonOptions(config.Daemon)

	connect_box.Container.Objects[0].(*fyne.Container).Add(menu.StartIndicators())

	max := LayoutAllItems(&d)
	max.(*fyne.Container).Objects[0].(*container.AppTabs).Append(container.NewTabItem("Log", rpc.SessionLog(app_tag, version)))

	go func() {
		time.Sleep(450 * time.Millisecond)
		w.SetContent(container.NewBorder(nil, container.NewVBox(layout.NewSpacer(), connect_box.Container), nil, nil, container.NewStack(d.Background, max)))
	}()

	w.ShowAndRun()
	<-done
	logger.Printf("[%s] Closed", app_tag)
}
