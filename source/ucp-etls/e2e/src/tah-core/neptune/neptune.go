package neptune

import (
	"encoding/json"
	"log"
	"strconv"
	"strings"
	core "tah/core/core"
)

type NeptuneConfig struct {
	Endpoint string
	Region   string
	Debug    bool
}

/*{
    "requestId": "f1330734-3c84-4694-b499-2e68c99b3c8d",
    "status": {
        "message": "",
        "code": 200,
        "attributes": {
            "@type": "g:Map",
            "@value": []
        }
    },
    "result": {
        "data": {
            "@type": "g:List",
            "@value": [{
                    "@type": "g:Vertex",
                    "@value": {
                        "id": "bebec108-1241-6170-d44b-c182bba9847c",
                        "label": "airport",
                        "properties": {
                            "id": [
                                {
                                    "@type": "g:VertexProperty",
                                    "@value": {
                                        "id": {
                                            "@type": "g:Int32",
                                            "@value": -281149059
                                        },
                                        "value": "ATL",
                                        "label": "id"
                                    }
                                }
                            ]
                        }
                    }
                }]
        },
        "meta": {
            "@type": "g:Map",
            "@value": []
        }
    }
}*/
type NeptuneResponse struct {
	Code            string                `json:"code"`
	DetailedMessage string                `json:"detailedMessage"`
	RequestID       string                `json:"requestId"`
	Status          NeptuneResponseStatus `json:"status"`
	Result          NeptuneResponseResult `json:"result"`
}

func (p NeptuneResponse) Decode(dec json.Decoder) (error, core.JSONObject) {
	return dec.Decode(&p), p
}

type NeptuneResponseStatus struct {
	Message    string              `json:"message"`
	Code       int                 `json:"code"`
	Attributes NeptuneResponseList `json:"attributes"`
}

type NeptuneResponseResult struct {
	Data NeptuneResponseList `json:"data"`
	Meta NeptuneResponseList `json:"meta"`
}

type NeptuneResponseList struct {
	Type  string          `json:"@type"`
	Value []NeptuneObject `json:"@value"`
}
type NeptuneObject struct {
	Type  string             `json:"@type"`
	Value NeptuneObjectValue `json:"@value"`
}
type NeptuneObjectValue struct {
	ID         string                             `json:"id"`
	Label      string                             `json:"label"`
	Properties map[string][]NeptuneObjectProperty `json:"properties"`
}

type NeptuneObjectProperty struct {
	Type  string                     `json:"@type"`
	Value NeptuneObjectPropertyValue `json:"@value"`
}

type NeptuneObjectPropertyValue struct {
	ID    NeptunePropertyId `json:"id"`
	Value string            `json:"value"`
	Label string            `json:"label"`
}

type NeptunePropertyId struct {
	Type  string `json:"@type"`
	Value int    `json:"@value"`
}

func Init(endpoint string, region string) NeptuneConfig {
	return InitWithDebug(endpoint, region, false)

}

func InitDebug(endpoint string, region string) NeptuneConfig {
	return InitWithDebug(endpoint, region, true)
}

func InitWithDebug(endpoint string, region string, debug bool) NeptuneConfig {
	return NeptuneConfig{
		Endpoint: endpoint,
		Region:   region,
		Debug:    debug,
	}
}

func (n NeptuneConfig) Query(query string) (NeptuneResponse, error) {
	log.Printf("[NEPTUNE]Connecting to %+v", n.Endpoint)

	url := "https://" + n.Endpoint + "/gremlin"
	log.Printf("[NEPTUNE] Url: %+v", url)
	svc := core.HttpInitWithRegion(url, n.Region)
	rq := map[string]string{
		"gremlin": query,
	}
	log.Printf("[NEPTUNE] Sending request: %+v", rq)
	res := NeptuneResponse{}
	options := core.RestOptions{
		SigV4Service: "neptune-db",
	}
	if n.Debug {
		options.Debug = true
	}
	res2, err := svc.HttpPost(rq, NeptuneResponse{}, options)
	res = res2.(NeptuneResponse)
	if err != nil {
		log.Printf("[NEPTUNE] Error for query  %+v => %v", query, err)
		return NeptuneResponse{}, err
	}
	log.Printf("[NEPTUNE]response: %+v", res)
	return res, err
}

////////////////////
// Gremelin Script generator
////////////////////////////

func (n NeptuneConfig) NewTraversal() *Travelsal {
	return &Travelsal{
		steps: []string{"g"},
		cfg:   n,
	}
}

type Travelsal struct {
	steps []string
	cfg   NeptuneConfig
}

func (g *Travelsal) Clear() {
	g.steps = []string{"g"}
}

func (g *Travelsal) ToScript() string {
	return strings.Join(g.steps, ".")
}
func (g *Travelsal) ToPrettyScript() string {
	return strings.Join(g.steps, ".\n")
}

func (g *Travelsal) AddStep(value string) *Travelsal {
	log.Printf("[NEPTUNE] Adding script step: %s", value)
	g.steps = append(g.steps, value)
	log.Printf("[NEPTUNE] New Script:\n %s", g.ToPrettyScript())
	return g
}

func (g *Travelsal) V(id string) *Travelsal {
	if id == "" {
		g.AddStep("V()")
	} else {
		g.AddStep("V('" + id + "')")
	}
	return g
}

func (g *Travelsal) E(id string) *Travelsal {
	if id == "" {
		g.AddStep("E()")
	} else {
		g.AddStep("E('" + id + "')")
	}
	return g
}

func (g *Travelsal) Has(attName string, attValue string) *Travelsal {
	g.AddStep("has('" + attName + "','" + attValue + "')")
	return g
}
func (g *Travelsal) Out() *Travelsal {
	return g
}
func (g *Travelsal) Limit(val int) *Travelsal {
	g.AddStep("limit(" + strconv.Itoa(val) + ")")
	return g
}

func (g *Travelsal) SimplePath() *Travelsal {
	g.AddStep("simplePath()")
	return g
}
func (g *Travelsal) Path() *Travelsal {
	g.AddStep("path()")
	return g
}

func (g *Travelsal) ToList() *Travelsal {
	g.AddStep("toList()")
	return g
}

func (g *Travelsal) By(val string) *Travelsal {
	g.AddStep("by('" + val + "')")
	return g
}

func (g *Travelsal) AddV(val string) *Travelsal {
	g.AddStep("addV('" + val + "')")
	return g
}
func (g *Travelsal) AddE(val string) *Travelsal {
	g.AddStep("addE('" + val + "')")
	return g
}
func (g *Travelsal) To(val string) *Travelsal {
	g.AddStep("to(g.V('" + val + "'))")
	return g
}
func (g *Travelsal) SID(attValue string) *Travelsal {
	g.AddStep("property(id,'" + attValue + "')")
	return g
}
func (g *Travelsal) SProperty(attName string, attValue string) *Travelsal {
	g.AddStep("property('" + attName + "','" + attValue + "')")
	return g
}
func (g *Travelsal) IPropertyI(attName string, attValue int) *Travelsal {
	g.AddStep("property('" + attName + "'," + strconv.Itoa(attValue) + ")")
	return g
}

func (g *Travelsal) Run() (NeptuneResponse, error) {
	log.Printf("[NEPTUNE] Running Gremlin Script to %s", g.ToScript())
	res, err := g.cfg.Query(g.ToScript())
	g.Clear()
	return res, err
}
