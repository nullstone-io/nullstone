package k8s

import (
	"fmt"
	"gopkg.in/nullstone-io/nullstone.v0/docker"
	core_v1 "k8s.io/api/core/v1"
)

func SetContainerImageTag(spec core_v1.PodSpec, containerName string, imageTag string) (core_v1.PodSpec, error) {
	result := spec

	if len(result.Containers) == 0 {
		return result, fmt.Errorf("cannot deploy service with no containers")
	}

	containerIndex := 0
	if containerName != "" {
		containerIndex = findContainerIndexByName(result.Containers, containerName)
		if containerIndex == -1 {
			return result, fmt.Errorf("cannot find container in spec with name = %q", containerName)
		}
	} else if len(result.Containers) > 1 {
		return result, fmt.Errorf("cannot set image tag because pod contains multiple containers and no name was specified")
	}

	existingImageUrl := docker.ParseImageUrl(result.Containers[containerIndex].Image)
	existingImageUrl.Digest = ""
	existingImageUrl.Tag = imageTag
	result.Containers[containerIndex].Image = existingImageUrl.String()

	return result, nil
}

func findContainerIndexByName(containers []core_v1.Container, name string) int {
	for i, container := range containers {
		if container.Name == name {
			return i
		}
	}
	return -1
}
