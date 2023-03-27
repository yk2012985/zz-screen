package models

type Allocatable struct {
	CPU              string `json:"cpu"`
	EphemeralStorage string `json:"ephemeral-storage"`
	Hugepages1Gi     string `json:"hugepages-1Gi"`
	Hugepages2Mi     string `json:"hugepages-2Mi"`
	Memory           string `json:"memory"`
	Pods             string `json:"pods"`
}

type Capacity struct {
	CoreWorker       string
	CPU              string
	EphemeralStorage string `json:"ephemeral-storage"`
	Hugepages1Gi     string
	Hugepages2Mi     string
	Memory           string
	Pods             string
	SocketWorker     string
}

type Status struct {
	Allocatable   Allocatable
	Capacity      Capacity
	clusterClaims interface{}
	conditions    interface{}
	version       interface{}
}

type Metadata struct {
	annotations       interface{}
	creationTimestamp string
	finalizers        interface{}
	generation        string
	labels            interface{}
	managedFields     interface{}
	Name              string
	ResourceVersion   string
	Uid               string
}

type ManagedCluster struct {
	ApiVersion string
	Kind       string
	Metadata   Metadata
	Spec       interface{}
	Status     Status
}

type ManagedClusters struct {
	ApiVersion string
	Items      []ManagedCluster
	Kind       string
	Metadata   interface{}
}

type ManagedClusterInfo struct {
	Name                     string
	CPUCapacity              string
	EphemeralStorageCapacity string
	MemoryCapacity           string
	PodsCapacity             string
}

type Result struct {
	Metric map[string]string `json:"metric"`
	Values []interface{}     `json:"value"`
}

type Data struct {
	ResultType string   `json:"resultType"`
	Results    []Result `json:"result"`
}

type UsageResponse struct {
	Status string `json:"status"`
	Data   Data   `json:"data"`
}

// searchAPIBody
type Filter struct {
	Property string   `json:"property"`
	Values   []string `json:"values"`
}

type Input struct {
	Keywords []string `json:"keywords"`
	Filters  []Filter `json:"filters"`
}

type Variables struct {
	Input []Input `json:"input"`
}

type SearchAPIBody struct {
	OperationName string    `json:"operationName"`
	Variables     Variables `json:"variables"`
	Query         string    `json:"query"`
}

// searchAPIResponse
type Item struct {
	ClusterNameSpace   string `json:"_clusterNamespace"`
	HubClusterResource string `json:"_hubClusterResource"`
	Uid                string `json:"_uid"`
	ApiVersion         string `json:"_apiversion"`
	Cluster            string `json:"cluster"`
	Container          string `json:"container"`
	Created            string `json:"created"`
	HostIP             string `json:"hostIP"`
	Image              string `json:"image"`
	Kind               string `json:"kind"`
	KindPlural         string `json:"kind_plural"`
	Label              string `json:"label"`
	Name               string `json:"name"`
	Namespace          string `json:"namespace"`
	PodIP              string `json:"podIP"`
	Restarts           string `json:"restarts"`
	StartedAt          string `json:"startedAt"`
	Status             string `json:"status"`
}

type SearchResult struct {
	Items    []Item `json:"items"`
	TypeName string `json:"__typename"`
}

type SearchAPIData struct {
	SearchResult []SearchResult `json:"searchResult"`
}

type SearchAPIResponse struct {
	Data SearchAPIData `json:"data"`
}
