package aurora

import "testing"

/*******
* Aurora
*****/

//TODO: update testt to dynamically create resources
func TestAurora(t *testing.T) {
	/*params := map[string]map[string]string{}
	params["dev"] = map[string]string{}
	params["dev"]["secretStoreArn"] = "arn:aws:secretsmanager:eu-central-1:660526416360:secret:auroraCredentialdev-XUN5Kl"
	params["dev"]["clusterArn"] = "arn:aws:rds:eu-central-1:660526416360:cluster:aurora-inventory-dev"
	params["dev"]["auroraDbName"] = "cloudrackInventory"
	params["dev"]["region"] = "eu-central-1"

	if params[env]["secretStoreArn"] != "" {
		dsConfig := aurora.InitDsWithRegion(params[env]["secretStoreArn"], params[env]["clusterArn"], params[env]["auroraDbName"], params[env]["region"])
		tx := core.NewTransaction("TEST_AURORA", "")
		dsConfig.SetTx(tx)
		_, err := dsConfig.Query("SHOW tables;")
		if err != nil {
			t.Errorf("[Aurora] Error conneccting DB: %+v ", err)
		}
		secCfg := secret.InitWithRegion(params[env]["secretStoreArn"], "eu-central-1")
		pswd := secCfg.Get("password")
		log.Printf("Aurora passowrd from secret manager: %+v", pswd)

		tName := "test_table"
		_, err = dsConfig.Query(fmt.Sprintf("CREATE TABLE %s (test_field INT);", tName))
		if err != nil {
			t.Errorf("[Aurora] CreateTable Error: %+v ", err)
		}
		if !dsConfig.HasTable(tName) {
			t.Errorf("[Aurora] HasTable Error: %+v ", errors.New("Should return TRUE"))

		}
		cols, err := dsConfig.GetColumns(tName)
		if err != nil {
			t.Errorf("[Aurora] GetColumns Error: %+v ", err)
		}
		if len(cols) != 1 {
			t.Errorf("[Aurora] GetColumns Invalid response, should return only 1 column %+v ", cols)
		}
		if cols[0].Field != "test_field" || cols[0].Type != "int(11)" {
			t.Errorf("[Aurora] GetColumns Invalid response %+v ", cols)
		}

		err = dsConfig.DropTable(tName)
		if err != nil {
			t.Errorf("[Aurora] CreateTable Error: %+v ", err)
		}
		if dsConfig.HasTable(tName) {
			t.Errorf("[Aurora] HasTable Error: %+v ", errors.New("Should return FALSE"))

		}
	}*/
}
