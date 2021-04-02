package nullstone

import "time"

type Block struct {
	Id                  int               `json:"id"`
	Name                string            `json:"name"`
	OrgName             string            `json:"orgName"`
	StackName           string            `json:"stackName"`
	Layer               string            `json:"layer"`
	ModuleSource        string            `json:"moduleSource"`
	ModuleSourceVersion string            `json:"moduleSourceVersion"`
	ParentBlocks        map[string]string `json:"parentBlocks"`
	CreatedAt           time.Time         `json:"createdAt"`
	UpdatedAt           time.Time         `json:"updatedAt"`
}
