package internal

type Config struct {
	ApiKey   string
	ZoneName string
	ApiUrl   string
}

type RecordResponse struct {
	Name    string   `json:"name"`
	Id      string   `json:"id"`
	Type    string   `json:"type"`
	Records []Record `json:"records"`
}

type ZoneResponse []Zone

type Record struct {
	Name       string `json:"name"`
	RootName   string `json:"rootName"`
	Type       string `json:"type"`
	Content    string `json:"content"`
	ChangeDate string `json:"changeDate"`
	Ttl        int    `json:"ttl"`
	Disabled   bool   `json:"disabled"`
	Id         string `json:"id"`
}

type Zone struct {
	Id   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

type Verification struct {
	Name  string `json:"name"`
	Token string `json:"token"`
}
