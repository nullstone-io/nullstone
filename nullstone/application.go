package nullstone

type Application struct {
	Id        int    `json:"id"`
	Name      string `json:"name"`
	OrgName   string `json:"orgName"`
	StackName string `json:"stackName"`
	Repo      string `json:"repo"`
	Framework string `json:"framework"`
	Block     *Block `json:"block"`
}
