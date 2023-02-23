package location

import "testing"

/*******
* LOCATION
*****/
func TestLocation(t *testing.T) {
	locSvc := Init("cloudrackTestIndex")
	err := locSvc.FindOrCreateIndex()
	if err != nil {
		t.Errorf("[core][Location] Error while creating index:" + err.Error())
	}
	//test omnipotence
	err = locSvc.FindOrCreateIndex()
	if err != nil {
		t.Errorf("[core][Location] Error while FindOrCreateIndex with existing index:" + err.Error())
	}
	err = locSvc.DeleteIndex()
	if err != nil {
		t.Errorf("[core][Location] Error while Deleting index:" + err.Error())
	}
}
