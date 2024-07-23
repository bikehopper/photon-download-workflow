package photon_download_workflow

import (
	"log"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	app_config "github.com/bikehopper/photon-download-workflow/src/app_config"
)

func Worker() {
	conf := app_config.New()
	hostPort := conf.TemporalUrl
	// The client and worker are heavyweight objects that should be created once per process.
	c, err := client.Dial(client.Options{
		HostPort: hostPort,
	})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	w := worker.New(c, "photon-download", worker.Options{
		EnableSessionWorker: true,
	})

	var activities *PhotonDownloadActivityObject

	w.RegisterWorkflow(PhotonDownload)
	w.RegisterActivity(activities)

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}
