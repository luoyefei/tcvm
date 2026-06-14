package cmd

import "tcvm/internal/tencent"

func newClient() (*tencent.Client, error) {
	if regionFlag != "" {
		return tencent.NewClientWithRegion(regionFlag)
	}
	return tencent.NewClient()
}
