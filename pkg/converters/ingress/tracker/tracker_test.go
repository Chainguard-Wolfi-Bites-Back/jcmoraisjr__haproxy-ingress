/*
Copyright 2020 The HAProxy Ingress Controller Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package tracker

import (
	"reflect"
	"sort"
	"testing"

	convtypes "github.com/jcmoraisjr/haproxy-ingress/pkg/converters/types"
	hatypes "github.com/jcmoraisjr/haproxy-ingress/pkg/haproxy/types"
)

type hostTracking struct {
	rtype    convtypes.ResourceType
	name     string
	hostname string
}

type backTracking struct {
	rtype   convtypes.ResourceType
	name    string
	backend hatypes.BackendID
}

type userTracking struct {
	rtype    convtypes.ResourceType
	name     string
	userlist string
}

type storageTracking struct {
	rtype   convtypes.ResourceType
	name    string
	storage string
}

var (
	back1a = hatypes.BackendID{
		Namespace: "default",
		Name:      "svc1",
		Port:      "8080",
	}
	back1b = hatypes.BackendID{
		Namespace: "default",
		Name:      "svc1",
		Port:      "8080",
	}
	back2a = hatypes.BackendID{
		Namespace: "default",
		Name:      "svc2",
		Port:      "8080",
	}
	back2b = hatypes.BackendID{
		Namespace: "default",
		Name:      "svc2",
		Port:      "8080",
	}
)

func TestGetDirtyLinks(t *testing.T) {
	testCases := []struct {
		trackedHosts    []hostTracking
		trackedBacks    []backTracking
		trackedUsers    []userTracking
		trackedStorages []storageTracking
		//
		trackedMissingHosts []hostTracking
		trackedMissingBacks []backTracking
		//
		oldIngressList      []string
		addIngressList      []string
		oldIngressClassList []string
		addIngressClassList []string
		oldConfigMapList    []string
		addConfigMapList    []string
		oldServiceList      []string
		addServiceList      []string
		oldSecretList       []string
		addSecretList       []string
		addPodList          []string
		//
		expDirtyIngs     []string
		expDirtyHosts    []string
		expDirtyBacks    []hatypes.BackendID
		expDirtyUsers    []string
		expDirtyStorages []string
	}{
		// 0
		{},
		// 1
		{
			oldIngressList: []string{"default/ing1"},
			expDirtyIngs:   []string{"default/ing1"},
		},
		// 2
		{
			oldServiceList: []string{"default/svc1"},
		},
		// 3
		{
			trackedHosts: []hostTracking{
				{convtypes.IngressType, "default/ing1", "domain1.local"},
			},
		},
		// 4
		{
			trackedHosts: []hostTracking{
				{convtypes.IngressType, "default/ing1", "domain1.local"},
			},
			oldIngressList: []string{"default/ing1"},
			expDirtyIngs:   []string{"default/ing1"},
			expDirtyHosts:  []string{"domain1.local"},
		},
		// 5
		{
			trackedHosts: []hostTracking{
				{convtypes.IngressType, "default/ing1", "domain1.local"},
				{convtypes.ServiceType, "default/svc1", "domain1.local"},
			},
			oldServiceList: []string{"default/svc1"},
			expDirtyIngs:   []string{"default/ing1"},
			expDirtyHosts:  []string{"domain1.local"},
		},
		// 6
		{
			trackedHosts: []hostTracking{
				{convtypes.IngressType, "default/ing1", "domain1.local"},
				{convtypes.SecretType, "default/secret1", "domain1.local"},
			},
			oldSecretList: []string{"default/secret1"},
			expDirtyIngs:  []string{"default/ing1"},
			expDirtyHosts: []string{"domain1.local"},
		},
		// 7
		{
			trackedHosts: []hostTracking{
				{convtypes.IngressType, "default/ing1", "domain1.local"},
			},
			trackedMissingHosts: []hostTracking{
				{convtypes.ServiceType, "default/svc1", "domain1.local"},
			},
			addServiceList: []string{"default/svc1"},
			expDirtyIngs:   []string{"default/ing1"},
			expDirtyHosts:  []string{"domain1.local"},
		},
		// 8
		{
			trackedHosts: []hostTracking{
				{convtypes.IngressType, "default/ing1", "domain1.local"},
			},
			trackedMissingHosts: []hostTracking{
				{convtypes.SecretType, "default/secret1", "domain1.local"},
			},
			addSecretList: []string{"default/secret1"},
			expDirtyIngs:  []string{"default/ing1"},
			expDirtyHosts: []string{"domain1.local"},
		},
		// 9
		{
			trackedHosts: []hostTracking{
				{convtypes.IngressType, "default/ing1", "domain1.local"},
				{convtypes.IngressType, "default/ing2", "domain2.local"},
			},
			oldIngressList: []string{"default/ing1"},
			expDirtyIngs:   []string{"default/ing1"},
			expDirtyHosts:  []string{"domain1.local"},
		},
		// 10
		{
			trackedHosts: []hostTracking{
				{convtypes.IngressType, "default/ing1", "domain1.local"},
				{convtypes.IngressType, "default/ing2", "domain1.local"},
				{convtypes.IngressType, "default/ing3", "domain2.local"},
			},
			oldIngressList: []string{"default/ing1"},
			expDirtyIngs:   []string{"default/ing1", "default/ing2"},
			expDirtyHosts:  []string{"domain1.local"},
		},
		// 11
		{
			trackedHosts: []hostTracking{
				{convtypes.IngressType, "default/ing1", "domain1.local"},
				{convtypes.IngressType, "default/ing2", "domain1.local"},
				{convtypes.IngressType, "default/ing2", "domain2.local"},
				{convtypes.IngressType, "default/ing3", "domain2.local"},
			},
			oldIngressList: []string{"default/ing1"},
			expDirtyIngs:   []string{"default/ing1", "default/ing2", "default/ing3"},
			expDirtyHosts:  []string{"domain1.local", "domain2.local"},
		},
		// 12
		{
			trackedHosts: []hostTracking{
				{convtypes.IngressType, "default/ing1", "domain1.local"},
				{convtypes.IngressType, "default/ing2", "domain2.local"},
			},
			trackedBacks: []backTracking{
				{convtypes.IngressType, "default/ing1", back1a},
			},
			oldIngressList: []string{"default/ing1"},
			expDirtyIngs:   []string{"default/ing1"},
			expDirtyHosts:  []string{"domain1.local"},
			expDirtyBacks:  []hatypes.BackendID{back1b},
		},
		// 13
		{
			trackedHosts: []hostTracking{
				{convtypes.IngressType, "default/ing1", "domain1.local"},
				{convtypes.IngressType, "default/ing2", "domain2.local"},
				{convtypes.IngressType, "default/ing3", "domain3.local"},
			},
			trackedBacks: []backTracking{
				{convtypes.IngressType, "default/ing1", back1a},
				{convtypes.IngressType, "default/ing2", back2a},
				{convtypes.IngressType, "default/ing3", back1b},
			},
			oldIngressList: []string{"default/ing1"},
			expDirtyIngs:   []string{"default/ing1", "default/ing3"},
			expDirtyHosts:  []string{"domain1.local", "domain3.local"},
			expDirtyBacks:  []hatypes.BackendID{back1b},
		},
		// 14
		{
			trackedBacks: []backTracking{
				{convtypes.IngressType, "default/ing1", back1a},
				{convtypes.SecretType, "default/secret1", back1a},
			},
			oldSecretList: []string{"default/secret1"},
			expDirtyIngs:  []string{"default/ing1"},
			expDirtyBacks: []hatypes.BackendID{back1b},
		},
		// 15
		{
			trackedMissingBacks: []backTracking{
				{convtypes.SecretType, "default/secret1", back1a},
			},
			addSecretList: []string{"default/secret1"},
			expDirtyBacks: []hatypes.BackendID{back1b},
		},
		// 16
		{
			trackedUsers: []userTracking{
				{convtypes.SecretType, "default/secret1", "usr1"},
				{convtypes.SecretType, "default/secret2", "usr2"},
			},
			oldSecretList: []string{"default/secret1"},
			expDirtyUsers: []string{"usr1"},
		},
		// 17
		{
			trackedBacks: []backTracking{
				{convtypes.PodType, "default/pod1", back1a},
				{convtypes.PodType, "default/pod2", back1a},
				{convtypes.PodType, "default/pod3", back2a},
				{convtypes.PodType, "default/pod4", back2a},
			},
			addPodList:    []string{"default/pod3"},
			expDirtyBacks: []hatypes.BackendID{back2b},
		},
		// 18
		{
			trackedStorages: []storageTracking{
				{convtypes.IngressType, "default/ing1", "crt1"},
				{convtypes.IngressType, "default/ing2", "crt2"},
				{convtypes.IngressType, "default/ing2", "crt3"},
				{convtypes.IngressType, "default/ing3", "crt4"},
			},
			oldIngressList:   []string{"default/ing2"},
			expDirtyIngs:     []string{"default/ing2"},
			expDirtyStorages: []string{"crt2", "crt3"},
		},
		// 19
		{
			trackedHosts: []hostTracking{
				{convtypes.IngressClassType, "haproxy", "app1.local"},
			},
			oldIngressClassList: []string{"haproxy"},
			expDirtyHosts:       []string{"app1.local"},
		},
		// 20
		{
			trackedMissingHosts: []hostTracking{
				{convtypes.IngressClassType, "haproxy", "app1.local"},
			},
			addIngressClassList: []string{"haproxy"},
			expDirtyHosts:       []string{"app1.local"},
		},
		// 21
		{
			trackedHosts: []hostTracking{
				{convtypes.ConfigMapType, "ingress/config", "app1.local"},
			},
			oldConfigMapList: []string{"ingress/config"},
			expDirtyHosts:    []string{"app1.local"},
		},
		// 22
		{
			trackedMissingHosts: []hostTracking{
				{convtypes.ConfigMapType, "ingress/config", "app1.local"},
			},
			addConfigMapList: []string{"ingress/config"},
			expDirtyHosts:    []string{"app1.local"},
		},
	}
	for i, test := range testCases {
		c := setup(t)
		for _, trackedHost := range test.trackedHosts {
			c.tracker.TrackHostname(trackedHost.rtype, trackedHost.name, trackedHost.hostname)
		}
		for _, trackedBack := range test.trackedBacks {
			c.tracker.TrackBackend(trackedBack.rtype, trackedBack.name, trackedBack.backend)
		}
		for _, trackedUser := range test.trackedUsers {
			c.tracker.TrackUserlist(trackedUser.rtype, trackedUser.name, trackedUser.userlist)
		}
		for _, trackedStorage := range test.trackedStorages {
			c.tracker.TrackStorage(trackedStorage.rtype, trackedStorage.name, trackedStorage.storage)
		}
		for _, trackedMissingHost := range test.trackedMissingHosts {
			c.tracker.TrackMissingOnHostname(trackedMissingHost.rtype, trackedMissingHost.name, trackedMissingHost.hostname)
		}
		for _, trackedMissingBack := range test.trackedMissingBacks {
			c.tracker.TrackMissingOnBackend(trackedMissingBack.rtype, trackedMissingBack.name, trackedMissingBack.backend)
		}
		dirtyIngs, dirtyHosts, dirtyBacks, dirtyUsers, dirtyStorages :=
			c.tracker.GetDirtyLinks(
				test.oldIngressList,
				test.addIngressList,
				test.oldIngressClassList,
				test.addIngressClassList,
				test.oldConfigMapList,
				test.addConfigMapList,
				test.oldServiceList,
				test.addServiceList,
				test.oldSecretList,
				test.addSecretList,
				test.addPodList,
			)
		sort.Strings(dirtyIngs)
		sort.Strings(dirtyHosts)
		sort.Slice(dirtyBacks, func(i, j int) bool {
			return dirtyBacks[i].String() < dirtyBacks[j].String()
		})
		sort.Strings(dirtyUsers)
		sort.Strings(dirtyStorages)
		c.compareObjects("dirty ingress", i, dirtyIngs, test.expDirtyIngs)
		c.compareObjects("dirty hosts", i, dirtyHosts, test.expDirtyHosts)
		c.compareObjects("dirty backs", i, dirtyBacks, test.expDirtyBacks)
		c.compareObjects("dirty users", i, dirtyUsers, test.expDirtyUsers)
		c.compareObjects("dirty storages", i, dirtyStorages, test.expDirtyStorages)
		c.teardown()
	}
}

func TestDeleteHostnames(t *testing.T) {
	testCases := []struct {
		trackedHosts []hostTracking
		//
		trackedMissingHosts []hostTracking
		//
		deleteHostnames []string
		//
		expIngressHostname      stringStringMap
		expHostnameIngress      stringStringMap
		expIngressClassHostname stringStringMap
		expHostnameIngressClass stringStringMap
		expConfigMapHostname    stringStringMap
		expHostnameConfigMap    stringStringMap
		expServiceHostname      stringStringMap
		expHostnameService      stringStringMap
		expSecretHostname       stringStringMap
		expHostnameSecret       stringStringMap
		//
		expIngressClassHostnameMissing stringStringMap
		expHostnameIngressClassMissing stringStringMap
		expConfigMapHostnameMissing    stringStringMap
		expHostnameConfigMapMissing    stringStringMap
		expServiceHostnameMissing      stringStringMap
		expHostnameServiceMissing      stringStringMap
		expSecretHostnameMissing       stringStringMap
		expHostnameSecretMissing       stringStringMap
	}{
		// 0
		{},
		// 1
		{
			deleteHostnames: []string{"domain1.local"},
		},
		// 2
		{
			trackedHosts: []hostTracking{
				{convtypes.IngressType, "default/ing1", "domain1.local"},
			},
			expIngressHostname: stringStringMap{"default/ing1": {"domain1.local": empty{}}},
			expHostnameIngress: stringStringMap{"domain1.local": {"default/ing1": empty{}}},
		},
		// 3
		{
			trackedHosts: []hostTracking{
				{convtypes.IngressType, "default/ing1", "domain1.local"},
			},
			deleteHostnames: []string{"domain1.local"},
		},
		// 4
		{
			trackedHosts: []hostTracking{
				{convtypes.ServiceType, "default/svc1", "domain1.local"},
			},
			deleteHostnames: []string{"domain1.local"},
		},
		// 5
		{
			trackedMissingHosts: []hostTracking{
				{convtypes.ServiceType, "default/svc1", "domain1.local"},
			},
			deleteHostnames: []string{"domain1.local"},
		},
		// 6
		{
			trackedHosts: []hostTracking{
				{convtypes.SecretType, "default/secret1", "domain1.local"},
			},
			deleteHostnames: []string{"domain1.local"},
		},
		// 7
		{
			trackedMissingHosts: []hostTracking{
				{convtypes.SecretType, "default/secret1", "domain1.local"},
			},
			deleteHostnames: []string{"domain1.local"},
		},
		// 8
		{
			trackedHosts: []hostTracking{
				{convtypes.IngressType, "default/ing1", "domain1.local"},
				{convtypes.IngressType, "default/ing1", "domain2.local"},
			},
			deleteHostnames: []string{"domain1.local", "domain2.local"},
		},
		// 9
		{
			trackedHosts: []hostTracking{
				{convtypes.IngressType, "default/ing1", "domain1.local"},
				{convtypes.IngressType, "default/ing1", "domain2.local"},
			},
			deleteHostnames:    []string{"domain1.local"},
			expIngressHostname: stringStringMap{"default/ing1": {"domain2.local": empty{}}},
			expHostnameIngress: stringStringMap{"domain2.local": {"default/ing1": empty{}}},
		},
		// 10
		{
			trackedHosts: []hostTracking{
				{convtypes.IngressType, "default/ing1", "domain1.local"},
				{convtypes.IngressType, "default/ing2", "domain1.local"},
			},
			deleteHostnames: []string{"domain1.local"},
		},
		// 11
		{
			trackedHosts: []hostTracking{
				{convtypes.IngressType, "default/ing1", "domain1.local"},
				{convtypes.IngressType, "default/ing2", "domain1.local"},
			},
			deleteHostnames: []string{"domain1.local", "domain2.local"},
		},
		// 12
		{
			trackedHosts: []hostTracking{
				{convtypes.IngressClassType, "haproxy1", "domain1.local"},
				{convtypes.IngressClassType, "haproxy2", "domain1.local"},
			},
			deleteHostnames: []string{"domain1.local", "domain2.local"},
		},
		// 13
		{
			trackedMissingHosts: []hostTracking{
				{convtypes.IngressClassType, "haproxy1", "domain1.local"},
				{convtypes.IngressClassType, "haproxy2", "domain1.local"},
			},
			deleteHostnames: []string{"domain1.local", "domain2.local"},
		},
		// 14
		{
			trackedHosts: []hostTracking{
				{convtypes.ConfigMapType, "ingress/config1", "domain1.local"},
				{convtypes.ConfigMapType, "ingress/config2", "domain1.local"},
			},
			deleteHostnames: []string{"domain1.local", "domain2.local"},
		},
		// 15
		{
			trackedMissingHosts: []hostTracking{
				{convtypes.ConfigMapType, "ingress/config1", "domain1.local"},
				{convtypes.ConfigMapType, "ingress/config2", "domain1.local"},
			},
			deleteHostnames: []string{"domain1.local", "domain2.local"},
		},
		// 16
		{
			trackedHosts: []hostTracking{
				{convtypes.IngressType, "default/ing1", "domain1.local"},
				{convtypes.IngressType, "default/ing1", "domain2.local"},
				{convtypes.IngressType, "default/ing1", "domain3.local"},
				{convtypes.IngressClassType, "haproxy", "domain1.local"},
				{convtypes.IngressClassType, "haproxy", "domain2.local"},
				{convtypes.IngressClassType, "haproxy", "domain3.local"},
				{convtypes.ConfigMapType, "ingress/config", "domain1.local"},
				{convtypes.ConfigMapType, "ingress/config", "domain2.local"},
				{convtypes.ConfigMapType, "ingress/config", "domain3.local"},
				{convtypes.ServiceType, "default/svc1", "domain1.local"},
				{convtypes.ServiceType, "default/svc1", "domain2.local"},
				{convtypes.ServiceType, "default/svc1", "domain3.local"},
				{convtypes.SecretType, "default/secret1", "domain1.local"},
				{convtypes.SecretType, "default/secret1", "domain2.local"},
				{convtypes.SecretType, "default/secret1", "domain3.local"},
			},
			deleteHostnames:         []string{"domain1.local", "domain2.local"},
			expIngressHostname:      stringStringMap{"default/ing1": {"domain3.local": empty{}}},
			expHostnameIngress:      stringStringMap{"domain3.local": {"default/ing1": empty{}}},
			expIngressClassHostname: stringStringMap{"haproxy": {"domain3.local": empty{}}},
			expHostnameIngressClass: stringStringMap{"domain3.local": {"haproxy": empty{}}},
			expConfigMapHostname:    stringStringMap{"ingress/config": {"domain3.local": empty{}}},
			expHostnameConfigMap:    stringStringMap{"domain3.local": {"ingress/config": empty{}}},
			expServiceHostname:      stringStringMap{"default/svc1": {"domain3.local": empty{}}},
			expHostnameService:      stringStringMap{"domain3.local": {"default/svc1": empty{}}},
			expSecretHostname:       stringStringMap{"default/secret1": {"domain3.local": empty{}}},
			expHostnameSecret:       stringStringMap{"domain3.local": {"default/secret1": empty{}}},
		},
	}
	for i, test := range testCases {
		c := setup(t)
		for _, trackedHost := range test.trackedHosts {
			c.tracker.TrackHostname(trackedHost.rtype, trackedHost.name, trackedHost.hostname)
		}
		for _, trackedMissingHost := range test.trackedMissingHosts {
			c.tracker.TrackMissingOnHostname(trackedMissingHost.rtype, trackedMissingHost.name, trackedMissingHost.hostname)
		}
		c.tracker.DeleteHostnames(test.deleteHostnames)
		c.compareObjects("ingressHostname", i, c.tracker.ingressHostname, test.expIngressHostname)
		c.compareObjects("hostnameIngress", i, c.tracker.hostnameIngress, test.expHostnameIngress)
		c.compareObjects("ingressClassHostname", i, c.tracker.ingressClassHostname, test.expIngressClassHostname)
		c.compareObjects("hostnameIngressClass", i, c.tracker.hostnameIngressClass, test.expHostnameIngressClass)
		c.compareObjects("configMapHostname", i, c.tracker.configMapHostname, test.expConfigMapHostname)
		c.compareObjects("hostnameConfigMap", i, c.tracker.hostnameConfigMap, test.expHostnameConfigMap)
		c.compareObjects("serviceHostname", i, c.tracker.serviceHostname, test.expServiceHostname)
		c.compareObjects("hostnameService", i, c.tracker.hostnameService, test.expHostnameService)
		c.compareObjects("secretHostname", i, c.tracker.secretHostname, test.expSecretHostname)
		c.compareObjects("hostnameSecret", i, c.tracker.hostnameSecret, test.expHostnameSecret)
		c.compareObjects("ingressClassHostnameMissing", i, c.tracker.ingressClassHostnameMissing, test.expIngressClassHostnameMissing)
		c.compareObjects("hostnameIngressClassMissing", i, c.tracker.hostnameIngressClassMissing, test.expHostnameIngressClassMissing)
		c.compareObjects("configMapHostnameMissing", i, c.tracker.configMapHostnameMissing, test.expConfigMapHostnameMissing)
		c.compareObjects("hostnameConfigMapMissing", i, c.tracker.hostnameConfigMapMissing, test.expHostnameConfigMapMissing)
		c.compareObjects("serviceHostnameMissing", i, c.tracker.serviceHostnameMissing, test.expServiceHostnameMissing)
		c.compareObjects("hostnameServiceMissing", i, c.tracker.hostnameServiceMissing, test.expHostnameServiceMissing)
		c.compareObjects("secretHostnameMissing", i, c.tracker.secretHostnameMissing, test.expSecretHostnameMissing)
		c.compareObjects("hostnameSecretMissing", i, c.tracker.hostnameSecretMissing, test.expHostnameSecretMissing)
		c.teardown()
	}
}

func TestDeleteBackends(t *testing.T) {
	testCases := []struct {
		trackedBacks []backTracking
		//
		trackedMissingBacks []backTracking
		//
		deleteBackends []hatypes.BackendID
		//
		expIngressBackend stringBackendMap
		expBackendIngress backendStringMap
		expSecretBackend  stringBackendMap
		expBackendSecret  backendStringMap
		//
		expSecretBackendMissing stringBackendMap
		expBackendSecretMissing backendStringMap
	}{
		// 0
		{},
		// 1
		{
			deleteBackends: []hatypes.BackendID{back1b},
		},
		// 2
		{
			trackedBacks: []backTracking{
				{convtypes.IngressType, "default/ing1", back1a},
			},
			expBackendIngress: backendStringMap{back1b: {"default/ing1": empty{}}},
			expIngressBackend: stringBackendMap{"default/ing1": {back1b: empty{}}},
		},
		// 3
		{
			trackedBacks: []backTracking{
				{convtypes.IngressType, "default/ing1", back1a},
			},
			deleteBackends: []hatypes.BackendID{back1b},
		},
		// 4
		{
			trackedBacks: []backTracking{
				{convtypes.IngressType, "default/ing1", back1a},
				{convtypes.IngressType, "default/ing2", back1a},
				{convtypes.IngressType, "default/ing2", back2a},
			},
			deleteBackends:    []hatypes.BackendID{back1b},
			expBackendIngress: backendStringMap{back2b: {"default/ing2": empty{}}},
			expIngressBackend: stringBackendMap{"default/ing2": {back2b: empty{}}},
		},
		// 5
		{
			trackedBacks: []backTracking{
				{convtypes.SecretType, "default/secret1", back1a},
				{convtypes.SecretType, "default/secret2", back1a},
				{convtypes.SecretType, "default/secret2", back2a},
			},
			trackedMissingBacks: []backTracking{
				{convtypes.SecretType, "default/secret1", back1a},
				{convtypes.SecretType, "default/secret2", back1a},
				{convtypes.SecretType, "default/secret2", back2a},
			},
			deleteBackends:          []hatypes.BackendID{back1b},
			expSecretBackend:        stringBackendMap{"default/secret2": {back2b: empty{}}},
			expBackendSecret:        backendStringMap{back2b: {"default/secret2": empty{}}},
			expSecretBackendMissing: stringBackendMap{"default/secret2": {back2b: empty{}}},
			expBackendSecretMissing: backendStringMap{back2b: {"default/secret2": empty{}}},
		},
	}
	for i, test := range testCases {
		c := setup(t)
		for _, trackedBack := range test.trackedBacks {
			c.tracker.TrackBackend(trackedBack.rtype, trackedBack.name, trackedBack.backend)
		}
		for _, trackedMissingBack := range test.trackedMissingBacks {
			c.tracker.TrackMissingOnBackend(trackedMissingBack.rtype, trackedMissingBack.name, trackedMissingBack.backend)
		}
		c.tracker.DeleteBackends(test.deleteBackends)
		c.compareObjects("ingressBackend", i, c.tracker.ingressBackend, test.expIngressBackend)
		c.compareObjects("backendIngress", i, c.tracker.backendIngress, test.expBackendIngress)
		c.compareObjects("secretBackend", i, c.tracker.secretBackend, test.expSecretBackend)
		c.compareObjects("backendSecret", i, c.tracker.backendSecret, test.expBackendSecret)
		c.compareObjects("secretBackendMissing", i, c.tracker.secretBackendMissing, test.expSecretBackendMissing)
		c.compareObjects("backendSecretMissing", i, c.tracker.backendSecretMissing, test.expBackendSecretMissing)
		c.teardown()
	}
}

func TestDeleteUserlists(t *testing.T) {
	testCases := []struct {
		trackedUsers []userTracking
		//
		deleteUserlists []string
		//
		expSecretUserlist stringStringMap
		expUserlistSecret stringStringMap
	}{
		// 0
		{},
		// 1
		{
			deleteUserlists: []string{"usr1"},
		},
		// 2
		{
			trackedUsers: []userTracking{
				{convtypes.SecretType, "default/secret1", "usr1"},
			},
			expUserlistSecret: stringStringMap{"usr1": {"default/secret1": empty{}}},
			expSecretUserlist: stringStringMap{"default/secret1": {"usr1": empty{}}},
		},
		// 3
		{
			trackedUsers: []userTracking{
				{convtypes.SecretType, "default/secret1", "usr1"},
			},
			deleteUserlists:   []string{"usr2"},
			expUserlistSecret: stringStringMap{"usr1": {"default/secret1": empty{}}},
			expSecretUserlist: stringStringMap{"default/secret1": {"usr1": empty{}}},
		},
		// 4
		{
			trackedUsers: []userTracking{
				{convtypes.SecretType, "default/secret1", "usr1"},
			},
			deleteUserlists: []string{"usr1"},
		},
		// 5
		{
			trackedUsers: []userTracking{
				{convtypes.SecretType, "default/secret1", "usr1"},
				{convtypes.SecretType, "default/secret2", "usr2"},
			},
			deleteUserlists:   []string{"usr2"},
			expUserlistSecret: stringStringMap{"usr1": {"default/secret1": empty{}}},
			expSecretUserlist: stringStringMap{"default/secret1": {"usr1": empty{}}},
		},
	}
	for i, test := range testCases {
		c := setup(t)
		for _, trackedUser := range test.trackedUsers {
			c.tracker.TrackUserlist(trackedUser.rtype, trackedUser.name, trackedUser.userlist)
		}
		c.tracker.DeleteUserlists(test.deleteUserlists)
		c.compareObjects("secretUserlist", i, c.tracker.secretUserlist, test.expSecretUserlist)
		c.compareObjects("userlistSecret", i, c.tracker.userlistSecret, test.expUserlistSecret)
		c.teardown()
	}
}

func TestDeleteStorages(t *testing.T) {
	testCases := []struct {
		trackedStorages []storageTracking
		//
		deleteStorages []string
		//
		expIngressStorages stringStringMap
		expStoragesIngress stringStringMap
	}{
		// 0
		{},
		// 1
		{
			deleteStorages: []string{"crt1"},
		},
		// 2
		{
			trackedStorages: []storageTracking{
				{convtypes.IngressType, "default/ingress1", "crt1"},
			},
			expIngressStorages: stringStringMap{"default/ingress1": {"crt1": empty{}}},
			expStoragesIngress: stringStringMap{"crt1": {"default/ingress1": empty{}}},
		},
		// 3
		{
			trackedStorages: []storageTracking{
				{convtypes.IngressType, "default/ingress1", "crt1"},
			},
			deleteStorages:     []string{"crt2"},
			expIngressStorages: stringStringMap{"default/ingress1": {"crt1": empty{}}},
			expStoragesIngress: stringStringMap{"crt1": {"default/ingress1": empty{}}},
		},
		// 4
		{
			trackedStorages: []storageTracking{
				{convtypes.IngressType, "default/ingress1", "crt1"},
			},
			deleteStorages: []string{"crt1"},
		},
		// 5
		{
			trackedStorages: []storageTracking{
				{convtypes.IngressType, "default/ingress1", "crt1"},
				{convtypes.IngressType, "default/ingress2", "crt2"},
			},
			deleteStorages:     []string{"crt2"},
			expIngressStorages: stringStringMap{"default/ingress1": {"crt1": empty{}}},
			expStoragesIngress: stringStringMap{"crt1": {"default/ingress1": empty{}}},
		},
	}
	for i, test := range testCases {
		c := setup(t)
		for _, trackedStorage := range test.trackedStorages {
			c.tracker.TrackStorage(trackedStorage.rtype, trackedStorage.name, trackedStorage.storage)
		}
		c.tracker.DeleteStorages(test.deleteStorages)
		c.compareObjects("ingressStorages", i, c.tracker.ingressStorages, test.expIngressStorages)
		c.compareObjects("storagesIngress", i, c.tracker.storagesIngress, test.expStoragesIngress)
		c.teardown()
	}
}

type testConfig struct {
	t       *testing.T
	tracker *tracker
}

func setup(t *testing.T) *testConfig {
	return &testConfig{
		t:       t,
		tracker: NewTracker().(*tracker),
	}
}

func (c *testConfig) teardown() {}

func (c *testConfig) compareObjects(name string, index int, actual, expected interface{}) {
	if !reflect.DeepEqual(actual, expected) {
		c.t.Errorf("%s on %d differs - expected: %v - actual: %v", name, index, expected, actual)
	}
}
