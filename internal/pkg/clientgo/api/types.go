package api

type PodInfo struct {
	Namespace       string
	Name            string
	Labels          map[string]string
	Annotations     map[string]string
	NodeSelector    map[string]string
	CPUmilliRequest int64
	MemoryBytes     int64
}
