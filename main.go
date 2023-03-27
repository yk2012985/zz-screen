package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"zz-screen/models"
)

// go语言内置net/http包提供了HTTP客户端和服务端的实现

var managedClusterInfoList []models.ManagedClusterInfo

const queryURL = "http://query-frontend-open-cluster-management-observability.ictnj.ac.cn/api/v1/query"
const cpuUsageQuery = "node_namespace_pod_container:container_cpu_usage_seconds_total:sum{cluster=\"clusterName\",namespace=\"namespaceName\"}"
const memUsageQuery = "sum(container_memory_rss:sum{cluster=\"clusterName\", container!=\"\"}) by (namespace)"
const cpuLimitQuery = "sum(kube_pod_container_resource_limits:sum{cluster=\"clusterName\", resource=\"cpu\"}) by (namespace)"
const memLimitQuery = "sum(kube_pod_container_resource_limits:sum{cluster=\"clusterName\", resource=\"memory\"}) by (namespace)"

const searchAPIURL = "https://search-api-open-cluster-management.ictnj.ac.cn/searchapi/graphql"
const cpuUsageForPod = "sum(node_namespace_pod_container:container_cpu_usage_seconds_total:sum_rate{namespace=\"namespaceName\", pod=\"podName\", container!=\"POD\", cluster=\"clusterName\"})"
const memUsageForPod = "sum(container_memory_working_set_bytes{cluster=\"clusterName\", namespace=\"namespaceName\", pod=\"podName\", container!=\"POD\", container!=\"\"})"

func main() {
	//cpuUsagePod := getMemUsageByPodName("local-cluster", "argocd", "argocd-application-controller-0")
	//fmt.Printf("podResult: %v\n", cpuUsagePod)
	//fmt.Printf("namespaceResult: %v\n", getMemUsage("local-cluster", "argocd"))

	//fmt.Printf("pods with label(%v) are: %v\n", "compName=web", getPodsInfoByLable("local-cluster", "gintest0317", "compName=web"))

	fmt.Printf("rightDownDisplay: %v\n", GetRightDownDisplay("local-cluster", "gintest0317"))

	//cpuUsagePodFloat, err := strconv.ParseFloat(cpuUsagePod[0:len(cpuUsagePod)-1], 64)
	//if err != nil {
	//	fmt.Println("..")
	//}
	//fmt.Println(cpuUsagePodFloat)

	fmt.Println()
	//getManagedClusterInfo()
	//
	//fmt.Println(managedClusterInfoList)
	//
	//fmt.Printf("左半部分内容: %v\n", GetLeftDisplay("kubeflow"))
	//
	//fmt.Printf("中上部分内容：%v\n", GetMidUpDisplay("local-cluster", "kubeflow"))
	//
	//fmt.Printf("中下部份内容：%v\n", GetMidDownDisplay("local-cluster", "kubeflow"))
	//
	//fmt.Printf("右上部分内容：%v\n", GetRightUpDisplay("local-cluster", "kubeflow"))

	//fmt.Println("=============")
	//fmt.Println(string(sendHttpPostRequest(searchAPIURL, "local-cluster", "argocd", "app.kubernetes.io/name=argocd-application-controller",
	//	"Bearer sha256~7eGQhFKja5a237PWsmJOjf7IItwrWBcGKdQB1JBCezM")))
	//fmt.Println(getPodsInfoByLable("local-cluster", "default", "inge=lll"))

}

// 底层发送HTTP GET Request
func sendHttpRequest(apiUrl string, arg url.Values) []byte {
	if arg != nil {
		u, err := url.ParseRequestURI(apiUrl)
		if err != nil {
			fmt.Printf("parse url requestUrl failed, err:%v\n", err)
		}
		u.RawQuery = arg.Encode()
		apiUrl = u.String()
	}
	// 创建普通的Http Get请求
	resp, err := http.Get(apiUrl)
	if err != nil {
		fmt.Printf("get failed, err:%v\n", err)
		return nil
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("read from resp.Body failed, err:%v\n", err)
		return nil
	}
	return body
}

// 底层发送HTTP POST Request
func sendHttpPostRequest(url string, requestBody []byte) []byte {
	var client *http.Client
	var request *http.Request
	var resp *http.Response

	// 跳过证书验证
	client = &http.Client{Transport: &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}}

	request, err := http.NewRequest("POST", url, strings.NewReader(string(requestBody)))
	if err != nil {
		fmt.Println("GetHttpSkip Request Error:", err)
		return nil
	}

	token, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/token")
	if err != nil {
		fmt.Println("read file failed, err:", err)
		return nil
	}

	// 加入token
	request.Header.Add("Authorization", "Bearer "+string(token))
	request.Header.Set("Content-Type", "application/json")
	resp, err = client.Do(request)
	if err != nil {
		fmt.Println("GetHttpSkip Response Error:", err)
		return nil
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	defer client.CloseIdleConnections()
	return body
}

// 获取纳管集群的资源总量信息，只获取一次，cpu：只有数字（单位是核），存储：数字加单位Gi， mem：数字加单位Mi
func getManagedClusterInfo() []models.ManagedClusterInfo {
	b := sendHttpRequest("http://127.0.0.1:8001/apis/cluster.open-cluster-management.io/v1/managedclusters", nil)

	// json string -> struct
	var managedClusters models.ManagedClusters
	err := json.Unmarshal(b, &managedClusters)
	if err != nil {
		fmt.Printf("json.Unmarshal failed, err:%v\n", err)
		return nil
	}
	for i := 0; i < len(managedClusters.Items); i++ {
		var managedClusterInfo models.ManagedClusterInfo
		managedClusterInfo.Name = managedClusters.Items[i].Metadata.Name

		managedClusterInfo.CPUCapacity = managedClusters.Items[i].Status.Capacity.CPU

		num, err := strconv.Atoi(managedClusters.Items[i].Status.Capacity.EphemeralStorage[0 : len(managedClusters.Items[i].Status.Capacity.EphemeralStorage)-2])
		if err != nil {
			fmt.Println("存储转换数字失败")
		}
		managedClusterInfo.EphemeralStorageCapacity = strconv.FormatFloat(math.Ceil(float64(num/1000000)), 'f', -1, 64) + "Gi"

		num, err = strconv.Atoi(managedClusters.Items[i].Status.Capacity.Memory[0 : len(managedClusters.Items[i].Status.Capacity.Memory)-2])
		if err != nil {
			fmt.Println("memory转换数字失败")
		}
		managedClusterInfo.MemoryCapacity = strconv.FormatFloat(math.Ceil(float64(num/1000)), 'f', -1, 64) + "Mi"

		managedClusterInfoList = append(managedClusterInfoList, managedClusterInfo)
	}
	return managedClusterInfoList
}

// 获取某cluster某namespace的cpu使用量，单位：m
func getCPUUsage(clusterName string, namespace string) string {
	var cpuUsage models.UsageResponse
	cpuData := url.Values{}
	cpuData.Set("query", strings.Replace(strings.Replace(cpuUsageQuery, "clusterName", clusterName, 1), "namespaceName", namespace, 1))
	cpuRespBody := sendHttpRequest(queryURL, cpuData)
	err := json.Unmarshal(cpuRespBody, &cpuUsage)
	if err != nil {
		fmt.Printf("json.Unmarshal failed, err:%v\n", err)
		return ""
	}
	if len(cpuUsage.Data.Results) > 0 {
		result, ok := cpuUsage.Data.Results[0].Values[1].(string)
		if ok {
			parseFloat, err := strconv.ParseFloat(result, 32)
			if err != nil {
				fmt.Printf("字符转化为数字失败, 失败原因：%v", err)
				return ""
			}
			return fmt.Sprintf("%.4f", parseFloat) + "m"
		}
		fmt.Println("结果解析失败")
		return ""
	}
	fmt.Println("未获取到相关结果")
	return ""
}

// 获取某cluster某namespace的mem使用量，单位：Mi
func getMemUsage(clusterName string, namespace string) string {
	var memUsage models.UsageResponse
	memData := url.Values{}
	memData.Set("query", strings.Replace(memUsageQuery, "clusterName", clusterName, 1))
	memRespBody := sendHttpRequest(queryURL, memData)
	err := json.Unmarshal(memRespBody, &memUsage)
	if err != nil {
		fmt.Printf("json.Unmarshal failed, err:%v\n", err)
		return ""
	}

	if len(memUsage.Data.Results) > 0 {
		for i := 0; i < len(memUsage.Data.Results); i++ {
			namespaceName, ok := memUsage.Data.Results[i].Metric["namespace"]
			if ok {
				if namespaceName == namespace {
					value, ok := memUsage.Data.Results[i].Values[1].(string)
					if ok {
						parseInt, err := strconv.ParseInt(value, 0, 64)
						if err != nil {
							fmt.Printf("string to int failed: %v", err.Error())
							return ""
						}
						valueToMi := float64(parseInt / 1000000)
						return fmt.Sprintf("%.2f", valueToMi) + "Mi"
					}
				}
			}
		}
	}
	return ""
}

// 获取某cluster某namespace的cpu limit
func getCPULimit(clusterName string, namespace string) string {
	var cpuLimit models.UsageResponse
	cpuData := url.Values{}
	cpuData.Set("query", strings.Replace(cpuLimitQuery, "clusterName", clusterName, 1))
	cpuRespBody := sendHttpRequest(queryURL, cpuData)
	err := json.Unmarshal(cpuRespBody, &cpuLimit)
	if err != nil {
		fmt.Printf("json.Unmarshal failed, err:%V\n", err)
		return ""
	}
	for i := 0; i < len(cpuLimit.Data.Results); i++ {
		if cpuLimit.Data.Results[i].Metric["namespace"] == namespace {
			result, ok := cpuLimit.Data.Results[i].Values[1].(string)
			if ok {
				return result + "m"
			}
		}
	}
	return ""
}

// 获取某cluster某namespace的mem limit，单位是Byte
func getMemLimit(clusterName string, namespace string) string {
	var memLimit models.UsageResponse
	memData := url.Values{}
	memData.Set("query", strings.Replace(memLimitQuery, "clusterName", clusterName, 1))
	memRespBody := sendHttpRequest(queryURL, memData)
	err := json.Unmarshal(memRespBody, &memLimit)
	if err != nil {
		fmt.Printf("json.Unmarshal failed, err:%V\n", err)
		return ""
	}
	for i := 0; i < len(memLimit.Data.Results); i++ {
		if memLimit.Data.Results[i].Metric["namespace"] == namespace {
			result, ok := memLimit.Data.Results[i].Values[1].(string)
			if ok {
				return result + "Mi"
			}
		}
	}
	return ""
}

// 获取某个集群某个namespace下的资源使用量（CPU：m，Mem：Mi，Disk：Gi）
func getResourceUsageForClsuterAndNamespace(clusterName string, namespace string) map[string]string {

	result := map[string]string{}
	if len(getCPUUsage(clusterName, namespace)) > 0 {
		result["CPUUsage"] = getCPUUsage(clusterName, namespace)
	} else {
		result["CPUUsage"] = "0.00%"
	}

	if len(getMemUsage(clusterName, namespace)) > 0 {
		result["MemUsage"] = getMemUsage(clusterName, namespace)
	} else {
		result["MemUsage"] = "0.00%"
	}

	// 获取存储使用情况
	//...

	return result
}

// 根据clusterName，namespace，label从ocm searchAPI获取pod信息
func getPodsInfoByLable(clusterName, namespace, label string) []string {
	filter1 := models.Filter{
		Property: "cluster",
		Values:   []string{clusterName},
	}
	filter2 := models.Filter{
		Property: "namespace",
		Values:   []string{namespace},
	}
	filter3 := models.Filter{
		Property: "label",
		Values:   []string{label},
	}
	filter4 := models.Filter{
		Property: "kind",
		Values:   []string{"Pod"},
	}
	input := models.Input{
		Keywords: []string{},
		Filters:  []models.Filter{filter1, filter2, filter3, filter4},
	}
	variables := models.Variables{
		Input: []models.Input{input},
	}
	searchAPIBody := models.SearchAPIBody{
		OperationName: "searchResultItems",
		Variables:     variables,
		Query:         "query searchResultItems($input: [SearchInput]) {\n  searchResult: search(input: $input) {\n    items\n    __typename\n  }\n}",
	}
	searchAPIBodySerialize, err := json.Marshal(searchAPIBody)
	if err != nil {
		fmt.Printf("json.Marshal failed, err:%v\n", err)
		return nil
	}
	searchResponse := sendHttpPostRequest(searchAPIURL, searchAPIBodySerialize)
	if searchResponse == nil {
		fmt.Println("get pods info failed")
		return nil
	}
	searchAPIResult := models.SearchAPIResponse{}
	err = json.Unmarshal(searchResponse, &searchAPIResult)
	if err != nil {
		fmt.Printf("searchAPIResonse json.Unmarshal failed, err:%v\n", err)
		return nil
	}
	if len(searchAPIResult.Data.SearchResult) == 0 {
		fmt.Println("have no data")
		return nil
	}
	var result []string
	for i := 0; i < len(searchAPIResult.Data.SearchResult[0].Items); i++ {
		result = append(result, searchAPIResult.Data.SearchResult[0].Items[i].Name)
	}
	return result

}

// 根据pod名称查询CPU使用量，单位：m
func getCPUUsageByPodName(clusterName, namespace, podName string) string {
	var cpuUsage models.UsageResponse
	cpuData := url.Values{}
	cpuData.Set("query", strings.Replace(strings.Replace(strings.Replace(cpuUsageForPod, "clusterName", clusterName, 1), "namespaceName", namespace, 1), "podName", podName, 1))
	cpuRespBody := sendHttpRequest(queryURL, cpuData)
	err := json.Unmarshal(cpuRespBody, &cpuUsage)
	if err != nil {
		fmt.Printf("json.Unmarshal failed, err:%v\n", err)
		return ""
	}
	if len(cpuUsage.Data.Results) > 0 {
		result, ok := cpuUsage.Data.Results[0].Values[1].(string)
		if ok {
			parseFloat, err := strconv.ParseFloat(result, 32)
			if err != nil {
				fmt.Printf("string to float64 failed：%v", err)
				return ""
			}
			return fmt.Sprintf("%.4f", parseFloat) + "m"
		}
		fmt.Println("result pasing failed")
		return ""
	}
	fmt.Printf("cluster: %v, namespace: %v, podName: %v have no  cpuUsage data", clusterName, namespace, podName)
	return ""
}

// 根据pod名称查询Mem使用量，单位：Byte
func getMemUsageByPodName(clusterName, namespace, podName string) string {
	var memUsage models.UsageResponse
	memData := url.Values{}
	memData.Set("query", strings.Replace(strings.Replace(strings.Replace(memUsageForPod, "clusterName", clusterName, 1), "namespaceName", namespace, 1), "podName", podName, 1))
	memRespBody := sendHttpRequest(queryURL, memData)
	err := json.Unmarshal(memRespBody, &memUsage)
	if err != nil {
		fmt.Printf("json.Unmarshal failed, err:%v\n", err)
		return ""
	}
	if len(memUsage.Data.Results) > 0 {
		result, ok := memUsage.Data.Results[0].Values[1].(string)
		if ok {
			parseFloat, err := strconv.ParseFloat(result, 32)
			if err != nil {
				fmt.Printf("string to float64 failed：%v", err)
				return ""
			}
			return fmt.Sprintf("%.4f", parseFloat/1000000) + "Mi"
		}
		fmt.Println("result pasing failed")
		return ""
	}
	fmt.Println("have no data")
	return ""
}

// GetLeftDisplay 左半部分，显示某个namespace在各个集群中的memory使用占比
func GetLeftDisplay(namespace string) map[string]string {
	result := map[string]string{}

	for i := 0; i < len(managedClusterInfoList); i++ {
		clusterName := managedClusterInfoList[i].Name
		var memUsageFloat float64
		memUsage := getMemUsage(clusterName, namespace)
		if len(memUsage) > 0 {
			memFloat, err := strconv.ParseFloat(memUsage[0:len(memUsage)-2], 64)
			if err != nil {
				fmt.Printf("memUsage parse failed：%v\n", err.Error())
			}
			memUsageFloat = memFloat
		}
		memCapFloat, err := strconv.ParseFloat(managedClusterInfoList[i].MemoryCapacity[0:len(managedClusterInfoList[i].MemoryCapacity)-2], 64)
		if err != nil {
			fmt.Printf("memCap parse failed: %v\n", err.Error())
			return nil
		}
		memUsagePer := fmt.Sprintf("%.2f", memUsageFloat*100/memCapFloat) + "%"
		result[clusterName] = memUsagePer

	}
	return result
}

// GetMidUpDisplay 中间上半部分获取指定集群的某个namespace的资源使用占比
func GetMidUpDisplay(clusterName string, namespace string) map[string]string {
	result := map[string]string{}
	// 获取cpu使用占比
	// 获取cpu使用量，单位：m
	var cpuUsageFloat float64
	cpuUsage := getCPUUsage(clusterName, namespace)
	if len(cpuUsage) > 0 {
		cpuFloat, err := strconv.ParseFloat(cpuUsage[0:len(cpuUsage)-1], 64)
		if err != nil {
			fmt.Printf("cpuUsage parse failed:%v", err.Error())
			return nil
		}
		cpuUsageFloat = cpuFloat
	}

	// 获取mem使用占比
	// 获取mem使用量，单位：Mi
	var memUsageFloat float64
	memUsage := getMemUsage(clusterName, namespace)
	if len(memUsage) > 0 {
		memFloat, err := strconv.ParseFloat(memUsage[0:len(memUsage)-2], 64)
		if err != nil {
			fmt.Printf("memUsage parse failed:%v", err.Error())
			return nil
		}
		memUsageFloat = memFloat
	}
	for i := 0; i < len(managedClusterInfoList); i++ {
		if managedClusterInfoList[i].Name == clusterName {
			cpuCapFloat, err := strconv.ParseFloat(managedClusterInfoList[i].CPUCapacity[0:len(managedClusterInfoList[i].CPUCapacity)], 64)
			memCapFloat, err := strconv.ParseFloat(managedClusterInfoList[i].MemoryCapacity[0:len(managedClusterInfoList[i].MemoryCapacity)-2], 64)
			if err != nil {
				fmt.Printf("resource cap parse failed:%v", err.Error())
				return nil
			}
			cpuUsagePer := fmt.Sprintf("%.2f", cpuUsageFloat/10/cpuCapFloat) + "%"
			memUsagePer := fmt.Sprintf("%.2f", memUsageFloat*100/memCapFloat) + "%"
			result["cpu"] = cpuUsagePer
			result["mem"] = memUsagePer
			return result
		}
	}
	return nil
}

// GetMidDownDisplay 中间下半部分获取指定集群的某个namespace在该namespace下limit值中的资源使用占比(cpu和mem)
func GetMidDownDisplay(clusterName string, namespace string) map[string]string {
	result := map[string]string{}
	var cpuUsageFloat float64
	cpuUsage := getCPUUsage(clusterName, namespace)
	if len(cpuUsage) > 0 {
		cpuFloat, err := strconv.ParseFloat(cpuUsage[0:len(cpuUsage)-1], 64)
		if err != nil {
			fmt.Printf("cpu usage parse failed:%v", err.Error())
			return nil
		}
		cpuUsageFloat = cpuFloat
	}
	var cpuLimitFloat float64
	cpuLimit := getCPULimit(clusterName, namespace)
	if len(cpuLimit) > 0 {
		cpuLFloat, err := strconv.ParseFloat(cpuLimit[0:len(cpuLimit)-1], 64)
		if err != nil {
			fmt.Printf("cpu limit parse failed:%v", err.Error())
			return nil
		}
		cpuLimitFloat = cpuLFloat
	}

	if cpuLimitFloat > 0 {
		result["cpu"] = fmt.Sprintf("%.2f", cpuUsageFloat*100/cpuLimitFloat) + "%"
	} else {
		result["cpu"] = "0.00%"
	}

	var memUsageFloat float64
	memUsage := getMemUsage(clusterName, namespace)
	if len(memUsage) > 0 {
		memFloat, err := strconv.ParseFloat(memUsage[0:len(memUsage)-2], 64)
		if err != nil {
			fmt.Printf("mem usage parse failed:%v", err.Error())
			return nil
		}
		memUsageFloat = memFloat
	}

	var memLimitFloat float64
	memLimit := getMemLimit(clusterName, namespace)
	if len(memLimit) > 0 {
		memLFloat, err := strconv.ParseFloat(memLimit[0:len(memLimit)-2], 64)
		if err != nil {
			fmt.Printf("mem limit parse failed:%v", err.Error())
			return nil
		}
		memLimitFloat = memLFloat
	}

	if memLimitFloat > 0 {
		result["mem"] = fmt.Sprintf("%.2f", memUsageFloat*100000000/memLimitFloat) + "%"
	} else {
		result["mem"] = "0.00%"
	}

	return result
}

// GetRightUpDisplay 右边上半部分获取实时使用数据
func GetRightUpDisplay(clusterName string, namespace string) map[string]string {
	result := map[string]string{}
	resourceUsage := getResourceUsageForClsuterAndNamespace(clusterName, namespace)
	result["cpu"] = resourceUsage["CPUUsage"]
	result["mem"] = resourceUsage["MemUsage"]
	return result

}

// GetRightDownDisplay 右边下半部分：某个cluster，某个namespace下根据pod标签查询资源使用
func GetRightDownDisplay(clusterName, namespace string) map[string]map[string]string {
	basicService := []string{"kafka", "imageproxy", "redis", "postgre", "minio", "zookeeper"}
	algorithmService := []string{"truck_violation_control", "vehicle_parking_violation", "non-vehicle_random_parking", "city_waterlogging_early_warning"}
	webService := []string{"web", "server"}
	result := map[string]map[string]string{}
	serviceTypeList := map[string][]string{
		"BasicService":     basicService,
		"AlgorithmService": algorithmService,
		"WebService":       webService,
	}
	for serviceType, labels := range serviceTypeList {
		resultForSingleService := map[string]string{}
		var cpuUsageSingleService float64
		var memUsageSingleService float64

		// 某个类型服务资源使用量计算
		for j := 0; j < len(labels); j++ {
			// 使用该类型服务特有的label获取对应的pods(ocm search api)
			label := "compName=" + labels[j]
			podsWithLabels := getPodsInfoByLable(clusterName, namespace, label)
			if len(podsWithLabels) > 0 {
				for i := 0; i < len(podsWithLabels); i++ {
					cpuUsage := getCPUUsageByPodName(clusterName, namespace, podsWithLabels[i])
					if len(cpuUsage) > 0 {
						cpuUsageFloat, err := strconv.ParseFloat(cpuUsage[0:len(cpuUsage)-1], 64)
						if err != nil {
							fmt.Printf("string to float64 failed: %v\n", err)
							continue
						}
						cpuUsageSingleService = cpuUsageSingleService + cpuUsageFloat
					}
					memUsage := getMemUsageByPodName(clusterName, namespace, podsWithLabels[i])
					if len(memUsage) > 0 {
						memUsageFloat, err := strconv.ParseFloat(memUsage[0:len(cpuUsage)-2], 64)
						if err != nil {
							fmt.Printf("string to float64 failed: %v\n", err)
							continue
						}
						memUsageSingleService = memUsageSingleService + memUsageFloat
					}
				}
			} else {
				continue
			}
		}
		resultForSingleService["cpu"] = fmt.Sprintf("%.4f", cpuUsageSingleService) + "m"
		resultForSingleService["mem"] = fmt.Sprintf("%.2f", memUsageSingleService) + "Mi"
		result[serviceType] = resultForSingleService
	}

	return result
}
