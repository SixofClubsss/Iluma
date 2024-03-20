package tarot

import (
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"fyne.io/fyne/v2"
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

const (
	appName = "Iluma"
	appID   = "dreamdapps.io.tarot"
)

var version = semver.MustParse("0.3.1-dev.x")

// Check tarot package version
func Version() semver.Version {
	return version
}

// Run Iluma as a single dApp
func StartApp() {
	n := runtime.NumCPU()
	runtime.GOMAXPROCS(n)

	// Initialize logrus logger to stdout
	gnomes.InitLogrusLog(logrus.InfoLevel)

	// Read config.json file
	config := menu.GetSettings(appName)

	// Initialize gnomes instance for app
	gnomon := gnomes.NewGnomes()

	// If no default background is set Iluma will use 'Glass'
	if config.Theme == "" {
		dreams.Theme.Name = "Glass"
	}

	// Initialize Fyne app and window as dreams.AppObject
	d := dreams.NewFyneApp(
		appID,
		appName,
		"Tarot Readings by Iluma",
		bundle.DeroTheme(config.Skin),
		resourceIlumaIconJpg,
		menu.DefaultBackgroundResource(),
		true)

	// Set one channel for tarot routine
	d.SetChannels(1)

	// Initialize close func anc channel
	done := make(chan struct{})

	closeFunc := func() {
		save := dreams.SaveData{
			Skin:   config.Skin,
			DBtype: gnomon.DBStorageType(),
			Theme:  dreams.Theme.Name,
		}

		if rpc.Daemon.Rpc == "" {
			save.Daemon = config.Daemon
		} else {
			save.Daemon = []string{rpc.Daemon.Rpc}
		}

		menu.StoreSettings(save)
		gnomon.Stop(appName)
		d.StopProcess()
		d.Window.Close()
	}

	d.Window.SetCloseIntercept(closeFunc)

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
				rpc.Wallet.Sync()

				d.SignalChannel()

			case <-d.Closing():
				logger.Printf("[%s] Closing...", appName)
				ticker.Stop()
				d.CloseAllDapps()
				time.Sleep(time.Second)
				done <- struct{}{}
				return
			}
		}
	}()

	// Create dwidget connection box, using default OnTapped for RPC/XSWD connections
	connection := dwidget.NewHorizontalEntries(appName, 1, &d)

	// Set any saved daemon configs
	connection.AddDaemonOptions(config.Daemon)

	// Adding dReams indicator panel for wallet, daemon and Gnomon
	connection.AddIndicator(menu.StartIndicators(nil))

	// Layout all items, appending with rpc.SessionLog tab
	max := LayoutAll(&d)
	max.(*fyne.Container).Objects[0].(*container.AppTabs).Append(container.NewTabItem("Log", rpc.SessionLog(appName, version)))

	// Start app and set content
	go func() {
		time.Sleep(450 * time.Millisecond)
		d.Window.SetContent(container.NewStack(container.NewStack(d.Background, max), container.NewVBox(layout.NewSpacer(), connection.Container)))
	}()

	d.Window.ShowAndRun()
	<-done
	logger.Printf("[%s] Closed", appName)
}
