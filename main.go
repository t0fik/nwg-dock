package main

import (
    "fmt"
    "log"
    "path/filepath"

    "github.com/dlasky/gotk3-layershell/layershell"
    "github.com/gotk3/gotk3/gdk"
    "github.com/gotk3/gotk3/glib"
    "github.com/gotk3/gotk3/gtk"
)

var (
    appDirs         []string
    configDirectory string
    pinnedFile      string
    pinned          []string
    oldTasks        []task
    mainBox         *gtk.Box
)

func buildMainBox(tasks []task, vbox *gtk.Box) {
    mainBox.Destroy()
    mainBox, _ = gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 0)
    vbox.PackStart(mainBox, false, false, 0)

    for _, pin := range pinned {
        button := pinnedButton(pin)
        mainBox.PackStart(button, false, false, 0)
    }

    var alreadyAdded []string
    for _, task := range tasks {
        if !inPinned(task.ID) {
            instances := taskInstances(task.ID, tasks)
            if len(instances) == 1 {
                button := taskButton(task, instances)
                /*button.Connect("clicked", func() {
                    focusCon(task.conID)
                })*/
                mainBox.PackStart(button, false, false, 0)
            } else if !isIn(alreadyAdded, task.ID) {
                button := taskButton(task, instances)
                mainBox.PackStart(button, false, false, 0)
                alreadyAdded = append(alreadyAdded, task.ID)
                taskMenu(task.ID, instances)
            } else {
                continue
            }
        }
    }
    mainBox.ShowAll()
}

func main() {
    configDirectory = configDir()
    // if doesn't exist:
    createDir(configDirectory)

    cacheDirectory := cacheDir()
    if cacheDirectory == "" {
        log.Panic("Couldn't determine cache directory location")
    }
    pinnedFile = filepath.Join(cacheDirectory, "nwg-dock-pinned")

    var err error
    pinned, err = loadTextFile(pinnedFile)
    if err != nil {
        pinned = nil
    }

    cssFile := filepath.Join(configDirectory, "style.css")

    appDirs = getAppDirs()

    gtk.Init(nil)

    cssProvider, _ := gtk.CssProviderNew()

    err = cssProvider.LoadFromPath(cssFile)
    if err != nil {
        fmt.Printf("%s file not found, using GTK styling\n", cssFile)
    } else {
        fmt.Printf("Using style: %s\n", cssFile)
        screen, _ := gdk.ScreenGetDefault()
        gtk.AddProviderForScreen(screen, cssProvider, gtk.STYLE_PROVIDER_PRIORITY_USER)
    }

    win, err := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
    if err != nil {
        log.Fatal("Unable to create window:", err)
    }
    layershell.InitForWindow(win)

    layershell.SetAnchor(win, layershell.LAYER_SHELL_EDGE_LEFT, false)
    layershell.SetAnchor(win, layershell.LAYER_SHELL_EDGE_BOTTOM, true)
    layershell.SetAnchor(win, layershell.LAYER_SHELL_EDGE_RIGHT, false)

    layershell.SetLayer(win, layershell.LAYER_SHELL_LAYER_TOP)
    layershell.SetMargin(win, layershell.LAYER_SHELL_EDGE_TOP, 0)
    layershell.SetMargin(win, layershell.LAYER_SHELL_EDGE_LEFT, 0)
    layershell.SetMargin(win, layershell.LAYER_SHELL_EDGE_RIGHT, 0)
    layershell.SetMargin(win, layershell.LAYER_SHELL_EDGE_BOTTOM, 0)

    win.Connect("destroy", func() {
        gtk.MainQuit()
    })

    vbox, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 0)
    win.Add(vbox)

    mainBox, _ = gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 0)
    vbox.PackStart(mainBox, true, true, 0)

    tasks, err := listTasks()
    if err != nil {
        log.Fatal("Couldn't list tasks:", err)
    }
    oldTasks = tasks

    buildMainBox(tasks, vbox)

    glib.TimeoutAdd(uint(250), func() bool {
        currentTasks, _ := listTasks()
        if len(currentTasks) != len(oldTasks) {
            fmt.Println("refreshing...")
            buildMainBox(currentTasks, vbox)
            oldTasks = currentTasks
        }
        return true
    })

    win.ShowAll()
    gtk.Main()
}
