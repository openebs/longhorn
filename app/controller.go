package app

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/gorilla/handlers"
	"github.com/openebs/longhorn/backend/dynamic"
	"github.com/openebs/longhorn/backend/file"
	"github.com/openebs/longhorn/backend/remote"
	"github.com/openebs/longhorn/controller"
	"github.com/openebs/longhorn/controller/rest"
	"github.com/openebs/longhorn/types"
	"github.com/openebs/longhorn/util"
)

var (
	frontends = map[string]types.Frontend{}
)

func ControllerCmd() cli.Command {
	return cli.Command{
		Name: "controller",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "listen",
				Value: "localhost:9501",
			},
			cli.StringFlag{
				Name:  "frontend",
				Value: "",
			},
			cli.StringSliceFlag{
				Name:  "enable-backend",
				Value: (*cli.StringSlice)(&[]string{"tcp"}),
			},
			cli.StringSliceFlag{
				Name: "replica",
			},
		},
		Action: func(c *cli.Context) {
			if err := startController(c); err != nil {
				logrus.Fatalf("Error running controller command: %v.", err)
			}
		},
	}
}

func startController(c *cli.Context) error {
	var controlIp string
	if c.NArg() == 0 {
		return errors.New("volume name is required")
	}
	name := c.Args()[0]

	listen := c.String("listen")
	backends := c.StringSlice("enable-backend")
	replicas := c.StringSlice("replica")
	frontendName := c.String("frontend")

	factories := map[string]types.BackendFactory{}
	for _, backend := range backends {
		switch backend {
		case "file":
			factories[backend] = file.New()
		case "tcp":
			factories[backend] = remote.New()
		default:
			logrus.Fatalf("Unsupported backend: %s", backend)
		}
	}

	var frontend types.Frontend
	if frontendName != "" {
		f, ok := frontends[frontendName]
		if !ok {
			return fmt.Errorf("Failed to find frontend: %s", frontendName)
		}
		frontend = f
	}
	if frontendName == "gotgt" {
		controlIp = c.Args()[1]
	}
	control := controller.NewController(name, controlIp, dynamic.New(factories), frontend)
	server := rest.NewServer(control)
	router := http.Handler(rest.NewRouter(server))

	router = util.FilteredLoggingHandler(map[string]struct{}{
		"/v1/volumes":  struct{}{},
		"/v1/replicas": struct{}{},
	}, os.Stdout, router)
	router = handlers.ProxyHeaders(router)

	if len(replicas) > 0 {
		logrus.Infof("Starting with replicas %q", replicas)
		if err := control.Start(replicas...); err != nil {
			log.Fatal(err)
		}
	}

	logrus.Infof("Listening on %s", listen)

	addShutdown(func() {
		control.Shutdown()
	})
	return http.ListenAndServe(listen, router)
}
