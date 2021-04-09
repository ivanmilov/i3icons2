package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"strings"

	"go.i3wm.org/i3/v4"
)

type vPrinter func(format string, v ...interface{})

func get_verbose_print(verbose bool) vPrinter {
	return func(format string, v ...interface{}) {
		if verbose {
			fmt.Printf(format, v...)
		}
	}
}

func main() {
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}

	// handle command line arguments
	var configFileName = flag.String("c", usr.HomeDir+"/.config/i3icons/i3icons2.config", "config file")
	var verbose = flag.Bool("v", false, "verbose")
	var default_icon = flag.String("d", "", "set default workspace icon")
	flag.Parse()

	var vprintf = get_verbose_print(*verbose)

	// Open our configFile
	configFile, err := os.Open(*configFileName)
	if err != nil {
		fmt.Println(err)
		flag.Usage()
		os.Exit(1)
	}
	defer configFile.Close()

	// read our config File and write to hash map
	byteValue, _ := ioutil.ReadAll(configFile)
	configLines := strings.Split(string(byteValue), "\n")
	config := make(map[string]string)
	for _, ci := range configLines {
		p := strings.Split(string(ci), "=")
		if len(p) == 2 {
			config[p[0]] = p[1]
		}
	}

	// subscribe to window events
	recv := i3.Subscribe(i3.WindowEventType, i3.WorkspaceEventType)
	for recv.Next() {
		ev := recv.Event()
		if win_ev, ok := ev.(*i3.WindowEvent); ok && win_ev.Change != "new" && win_ev.Change != "close" && win_ev.Change != "move" {
			vprintf("skip win ev [%s]\n", win_ev.Change)
			continue
		}
		if ws_ev, ok := recv.Event().(*i3.WorkspaceEvent); ok && ws_ev.Change != "init" && ws_ev.Change != "empty" {
			vprintf("skip ws ev [%s]\n", ws_ev.Change)
			continue
		}

		tree, err := i3.GetTree()
		if err != nil {
			log.Fatal(err)
		}

		// i3-msg -t get_workspaces doesn't fill IDs for the workspaces
		// wss, err := i3.GetWorkspaces()
		// if err != nil {
		// 	log.Fatal(err)
		// }
		wss := GetWorkspaces(tree)

		for _, ws := range wss {
			name := ws.Name
			number := strings.Split(name, " ")[0]
			windows := Leaves(tree, int64(ws.ID))
			newname := number
			vprintf("{%s} has wins:\n", name)

			windownames := make([]string, len(windows))
			for i, win := range windows {
				winname := strings.ToLower(win.WindowProperties.Class)
				vprintf("\t%s\n", winname)

				// rename window to config item, if present
				if val, ok := config[winname]; ok {
					winname = val
				} else if len(winname) > 7 {
					winname = winname[:4] + ".." + winname[len(winname)-3:]
				}
				// check if workspace name already contains window title
				choose := true
				for _, n := range windownames {
					if strings.Compare(n, winname) == 0 {
						choose = false
					}
				}
				if choose {
					windownames[i] = winname
				}
			}
			// rename workspace
			for _, windowname := range windownames {
				if len(windowname) > 0 {
					newname = fmt.Sprintf("%s %s", newname, windowname)
				}
			}

			if newname == number {
				vprintf("\t\t\tset default for ws [%s]\n", newname)
				newname += " " + *default_icon
			}
			if name != newname /* && newname != number */ {
				vprintf("[%s] -> [%s]\n", name, newname)
				cmd := "rename workspace \"" + name + "\" to \"" + newname + "\""
				vprintf("%s\n", cmd)

				i3.RunCommand(cmd)
			}
		}
		vprintf("-----------------\n")
	}
}
