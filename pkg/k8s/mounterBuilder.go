package k8s

import (
	"github.com/feedhenry/mcp-standalone/pkg/mobile"
	"github.com/pkg/errors"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
)

// MounterBuilder is a factory for MountManager
type MounterBuilder struct {
	k8s       kubernetes.Interface
	namespace string
}

// NewMounterBuilder creates a new MounterBuilder in the provided namespace
func NewMounterBuilder(namespace string) mobile.MounterBuilder {
	return &MounterBuilder{namespace: namespace}
}

// WithK8s will return a pointer to a MounterBuilder using the provided kubernetes client
func (mb *MounterBuilder) WithK8s(k8s kubernetes.Interface) mobile.MounterBuilder {
	return &MounterBuilder{k8s: k8s, namespace: mb.namespace}
}

// Build a new MountManager from the configured MounterBuilder
func (mb *MounterBuilder) Build() mobile.VolumeMounterUnmounter {
	return &MountManager{k8s: mb.k8s, namespace: mb.namespace}
}

// MountManager can mount and unmount into services
type MountManager struct {
	k8s       kubernetes.Interface
	namespace string
}

// Mount a secret named mount into the service
func (mm *MountManager) Mount(secret, clientService string) error {
	if _, err := mm.k8s.CoreV1().Secrets(mm.namespace).Get(secret, meta_v1.GetOptions{}); err != nil {
		return errors.Wrap(err, "k8s.mm.Mount -> could not find secret: "+secret)
	}
	deploy, err := mm.k8s.AppsV1beta1().Deployments(mm.namespace).Get(clientService, meta_v1.GetOptions{})
	if err != nil {
		return errors.Wrap(err, "k8s.mm.Mount -> could not find deployment named: "+clientService)
	}
	id := findContainerID(clientService, deploy.Spec.Template.Spec.Containers)
	if id < 0 {
		return errors.New("k8s.mm.Mount -> could not find container in deployment with name: " + clientService)
	}

	//only create the volume if it doesn't exist yet
	if _, vol := findVolumeByName(secret, deploy.Spec.Template.Spec.Volumes); vol.Name != secret {
		newVol := v1.Volume{
			Name: secret,
			VolumeSource: v1.VolumeSource{
				Secret: &v1.SecretVolumeSource{
					SecretName: secret,
				},
			},
		}
		deploy.Spec.Template.Spec.Volumes = append(deploy.Spec.Template.Spec.Volumes, newVol)
	}
	if _, volMount := findMountByName(secret, deploy.Spec.Template.Spec.Containers[id].VolumeMounts); volMount.Name != secret {
		newMount := v1.VolumeMount{Name: secret, ReadOnly: true, MountPath: "/etc/secrets/" + secret}
		deploy.Spec.Template.Spec.Containers[id].VolumeMounts = append(deploy.Spec.Template.Spec.Containers[id].VolumeMounts, newMount)
	}

	_, err = mm.k8s.AppsV1beta1().Deployments(mm.namespace).Update(deploy)
	if err != nil {
		return errors.Wrap(err, "k8s.mm.Mount -> could not update deploy config with new mount and volume")
	}

	return nil
}

// Unmount a secret named mount from the service
func (mm *MountManager) Unmount(secret, clientService string) error {
	if _, err := mm.k8s.CoreV1().Secrets(mm.namespace).Get(secret, meta_v1.GetOptions{}); err != nil {
		return errors.New("k8s.mm.Unmount -> could not find secret: " + secret)
	}
	deploy, err := mm.k8s.AppsV1beta1().Deployments(mm.namespace).Get(clientService, meta_v1.GetOptions{})
	if err != nil {
		return errors.Wrap(err, "k8s.mm.Mount -> could not find deployment named: "+clientService)
	}
	id := findContainerID(clientService, deploy.Spec.Template.Spec.Containers)
	if id < 0 {
		return errors.New("k8s.mm.Mount -> could not find container in deployment with name: " + clientService)
	}

	if i, _ := findVolumeByName(secret, deploy.Spec.Template.Spec.Volumes); i >= 0 {
		deploy.Spec.Template.Spec.Volumes = append(deploy.Spec.Template.Spec.Volumes[:i], deploy.Spec.Template.Spec.Volumes[i+1:]...)
	}
	if i, _ := findMountByName(secret, deploy.Spec.Template.Spec.Containers[id].VolumeMounts); i >= 0 {
		deploy.Spec.Template.Spec.Containers[id].VolumeMounts = append(deploy.Spec.Template.Spec.Containers[id].VolumeMounts[:i], deploy.Spec.Template.Spec.Containers[id].VolumeMounts[i+1:]...)
	}

	_, err = mm.k8s.AppsV1beta1().Deployments(mm.namespace).Update(deploy)
	if err != nil {
		return errors.Wrap(err, "k8s.mm.Unmount -> could not update deploy config with new mount and volume")
	}

	return nil
}

func findContainerID(name string, containers []v1.Container) int {
	for id, container := range containers {
		if container.Name == name {
			return id
		}
	}
	return -1
}

func findVolumeByName(name string, volumes []v1.Volume) (int, v1.Volume) {
	for i, vol := range volumes {
		if vol.Name == name {
			return i, vol
		}
	}

	return -1, v1.Volume{}
}

func findMountByName(name string, mounts []v1.VolumeMount) (int, v1.VolumeMount) {
	for i, mount := range mounts {
		if mount.Name == name {
			return i, mount
		}
	}

	return -1, v1.VolumeMount{}
}
