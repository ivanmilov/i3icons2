package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"sort"
	"strings"

	"github.com/tidwall/gjson"
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

func contains(s []string, searchterm string) bool {
	i := sort.SearchStrings(s, searchterm)
	return i < len(s) && s[i] == searchterm
}

func main() {
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}

	// handle command line arguments
	var configFileName = flag.String("c", usr.HomeDir+"/.config/i3icons/i3icons2.json", "config file")
	var verbose = flag.Bool("v", false, "verbose")
	var default_icon = flag.String("d", "", "set default workspace icon")
	flag.Parse()

	var vprintf = get_verbose_print(*verbose)

	conf, err := ioutil.ReadFile(*configFileName)
	if err != nil {
		fmt.Println(err)
		flag.Usage()
		os.Exit(1)
	}

	jconf := gjson.Get(string(conf), "config")
	config := make(map[string]string)
	jconf.ForEach(func(key, val gjson.Result) bool {
		config[key.String()] = val.String()
		return true
	})

	jignore := gjson.Get(string(conf), "ignore")
	var ignore []string
	jignore.ForEach(func(key, val gjson.Result) bool {
		ignore = append(ignore, val.String())
		return true
	})
	sort.Strings(ignore)

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

				if contains(ignore, winname) {
					vprintf("\t%s is in ignore\n", winname)
					continue
				}
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
