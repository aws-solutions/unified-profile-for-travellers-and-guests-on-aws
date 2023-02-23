package neptune

import (
	"log"
	"os"
	"testing"
)

var env = os.Getenv("TAH_CORE_ENV")
var TAH_CORE_REGION = os.Getenv("TAH_CORE_REGION")

func TestNeptune(t *testing.T) {
	nep := Init("endpoint", TAH_CORE_REGION)
	res, err := nep.Query("g.V().limit(1)")
	log.Printf("[NEPTUNE_TEST]response: %+v", res)
	log.Printf("[NEPTUNE_TEST]error: %+v", err)

}
func TestNeptuneScript(t *testing.T) {
	nep := Init("endpoint", TAH_CORE_REGION)
	g := nep.NewTraversal()
	g.AddV("airport").SProperty("code", "BOS")
	res := "g.addV('airport').property('code','BOS')"
	if g.ToScript() != res {
		t.Errorf("[NEPTUNE] Script should be %v and not %v", res, g.ToScript())
	}
	//Clear traversal
	g.Clear()
	res = "g"
	if g.ToScript() != res {
		t.Errorf("[NEPTUNE] After clear: Script should be %v and not %v", res, g.ToScript())
	}
}
