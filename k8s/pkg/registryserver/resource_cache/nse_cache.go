package resource_cache

import (
	"github.com/networkservicemesh/networkservicemesh/k8s/pkg/apis/networkservice/v1"
	"github.com/networkservicemesh/networkservicemesh/k8s/pkg/networkservice/informers/externalversions"
	"github.com/sirupsen/logrus"
)

type NetworkServiceEndpointCache struct {
	cache                   abstractResourceCache
	nseByNs                 map[string][]*v1.NetworkServiceEndpoint
	networkServiceEndpoints map[string]*v1.NetworkServiceEndpoint
}

func NewNetworkServiceEndpointCache() *NetworkServiceEndpointCache {
	rv := &NetworkServiceEndpointCache{
		nseByNs:                 make(map[string][]*v1.NetworkServiceEndpoint),
		networkServiceEndpoints: make(map[string]*v1.NetworkServiceEndpoint),
	}
	config := cacheConfig{
		keyFunc:             getNseKey,
		resourceAddedFunc:   rv.resourceAdded,
		resourceDeletedFunc: rv.resourceDeleted,
		resourceType:        NseResource,
	}
	rv.cache = newAbstractResourceCache(config)
	return rv
}

func (c *NetworkServiceEndpointCache) Get(key string) *v1.NetworkServiceEndpoint {
	return c.networkServiceEndpoints[key]
}

func (c *NetworkServiceEndpointCache) GetByNetworkService(networkServiceName string) []*v1.NetworkServiceEndpoint {
	return c.nseByNs[networkServiceName]
}

func (c *NetworkServiceEndpointCache) GetByNetworkServiceManager(nsmName string) []*v1.NetworkServiceEndpoint {
	var rv []*v1.NetworkServiceEndpoint

	for _, endpoint := range c.networkServiceEndpoints {
		if endpoint.Spec.NsmName == nsmName {
			rv = append(rv, endpoint)
		}
	}

	return rv
}

func (c *NetworkServiceEndpointCache) Add(nse *v1.NetworkServiceEndpoint) {
	logrus.Infof("Adding NSE to cache: %v", *nse)
	c.cache.add(nse)
}

func (c *NetworkServiceEndpointCache) Delete(key string) {
	c.cache.delete(key)
}

func (c *NetworkServiceEndpointCache) Start(informerFactory externalversions.SharedInformerFactory) (func(), error) {
	return c.cache.start(informerFactory)
}

func (c *NetworkServiceEndpointCache) resourceAdded(obj interface{}) {
	nse := obj.(*v1.NetworkServiceEndpoint)
	endpoints := c.nseByNs[nse.Spec.NetworkServiceName]
	if _, exist := c.networkServiceEndpoints[getNseKey(nse)]; !exist {
		c.nseByNs[nse.Spec.NetworkServiceName] = append(endpoints, nse)
	} else {
		for i, e := range endpoints {
			if getNseKey(nse) == getNseKey(e) {
				endpoints[i] = nse
				break
			}
		}
	}
	c.networkServiceEndpoints[getNseKey(nse)] = nse
}

func (c *NetworkServiceEndpointCache) resourceDeleted(key string) {
	nse, exist := c.networkServiceEndpoints[key]
	if !exist {
		return
	}

	endpoints := c.nseByNs[nse.Spec.NetworkServiceName]
	var index int
	for i, e := range endpoints {
		if getNseKey(nse) == getNseKey(e) {
			index = i
			break
		}
	}
	endpoints = append(endpoints[:index], endpoints[index+1:]...)
	if len(endpoints) == 0 {
		delete(c.nseByNs, nse.Spec.NetworkServiceName)
	} else {
		c.nseByNs[nse.Spec.NetworkServiceName] = endpoints
	}
	delete(c.networkServiceEndpoints, key)
}

func getNseKey(obj interface{}) string {
	return obj.(*v1.NetworkServiceEndpoint).Name
}
