package main

var containersToRun = make(chan string)

func StartContainerBgWorker() {
	// listen to the channel
	for containerId := range containersToRun {
		go runContainer(containerId)
	}
}

func runContainer(containerId string) {
	container, err := FetchContainerByUUID(containerId)
	if err != nil {
		return
	}
	container.PullImage()
	if container.Status != ContainerStatusImagePulled {
		return
	}
	container.Run()
}
