// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package admin

import (
	"log"
	"testing"

	customerprofiles "tah/upt/source/storage"
	"tah/upt/source/tah-core/core"
	"tah/upt/source/tah-core/db"
	model "tah/upt/source/ucp-backend/src/business-logic/model/common"
	"tah/upt/source/ucp-backend/src/business-logic/usecase/registry"
	constant "tah/upt/source/ucp-common/src/constant/admin"
	"tah/upt/source/ucp-common/src/utils/config"
)

func TestValidateRequest(t *testing.T) {
	t.Parallel()
	cfgs := []TestCfg{
		{Tpl: "http://www.example.com", Fields: []string{"id", "test1", "test2"}, Valid: false},
		{Tpl: "", Fields: []string{"id", "test1", "test2"}, Valid: false},
		{Tpl: "https:/www.example.com", Fields: []string{"id", "test1", "test2"}, Valid: false},
		{Tpl: "https//www.example.com", Fields: []string{"id", "test1", "test2"}, Valid: false},
		{Tpl: "https:www.example.com", Fields: []string{"id", "test1", "test2"}, Valid: false},
		{Tpl: "https://www.example.com", Fields: []string{"id", "test1", "test2"}, Valid: true},
		{Tpl: "https://www.example.com", Fields: []string{}, Valid: true},
		{Tpl: "https://www.example.com/{{id}}", Fields: []string{"id", "test1", "test2"}, Valid: true},
		{Tpl: "https://www.example.com/{{id}}?test={{test1}}&test2={{test2}}", Fields: []string{"id", "test1", "test2"}, Valid: true},
		{Tpl: "https://www.example.com/{{invalid_field}}", Fields: []string{"id", "test1", "test2"}, Valid: false},
		{Tpl: "https://www.example.com/{{id}}?test={{test1}}&test2={{invalidField}}", Fields: []string{"id", "test1", "test2"}, Valid: false},
		{Tpl: "https://www.example.com/{{id}}?test={{test1}}&test2={{invalid", Fields: []string{"id", "test1", "test2"}, Valid: false},
	}
	for _, cfg := range cfgs {
		err := validateHyperlinkTemplate(cfg.Tpl, cfg.Fields)
		if err != nil && cfg.Valid {
			t.Errorf("[TestValidateRequest] error validate template: %v => %v. Should be valid", cfg, err)
		}
		if err == nil && !cfg.Valid {
			t.Errorf("[TestValidateRequest] error validate template: %v. Should be invalid", cfg)
		}
	}
}

func TestUpdatePortalConfig(t *testing.T) {
	t.Parallel()
	// Set up resources
	envCfg, err := config.LoadEnvConfig()
	if err != nil {
		t.Fatalf("unable to load configs: %v", err)
	}
	tName := "ucp-test-portal-config-" + core.GenerateUniqueId()
	dynamo_portal_config_pk := "config_item"
	dynamo_portal_config_sk := "config_item_category"
	tx := core.NewTransaction(t.Name(), "", core.LogLevelDebug)

	cfg, err := db.InitWithNewTable(tName, dynamo_portal_config_pk, dynamo_portal_config_sk, "", "")
	if err != nil {
		t.Fatalf("[TestUpdatePortalConfig] error init with new table: %v", err)
	}
	t.Cleanup(func() {
		err = cfg.DeleteTable(tName)
		if err != nil {
			t.Errorf("Error deleting table %v", err)
		}
	})
	cfg.WaitForTableCreation()

	mock := customerprofiles.InitMockV2()
	mock.On("GetProfileLevelFields").Return([]string{"FirstName", "LastName"}, nil)

	reg := registry.NewRegistry(envCfg.Region, core.LogLevelDebug, registry.ServiceHandlers{PortalConfigDB: &cfg, Accp: mock})

	uc := NewUpdatePortalConfig()
	uc.SetRegistry(&reg)
	uc.SetTx(&tx)
	getuc := NewGetPortalConfig()
	getuc.SetRegistry(&reg)
	getuc.SetTx(&tx)

	varlidRq := model.RequestWrapper{
		PortalConfig: model.PortalConfig{
			HyperlinkMappings: []model.HyperlinkMapping{
				{
					AccpObject: "air_booking",
					FieldName:  "booking_id",
					Template:   "https://www.example.com",
				},
				{
					AccpObject: "air_booking",
					FieldName:  "from",
					Template:   "https://www.example.com/{{booking_id}}",
				},
			},
		},
	}

	invalidObject := model.RequestWrapper{
		PortalConfig: model.PortalConfig{
			HyperlinkMappings: []model.HyperlinkMapping{
				{
					AccpObject: "invalid_object",
					FieldName:  "booking",
					Template:   "https://www.example.com",
				},
			},
		},
	}
	invalidFieldName := model.RequestWrapper{
		PortalConfig: model.PortalConfig{
			HyperlinkMappings: []model.HyperlinkMapping{
				{
					AccpObject: "air_booking",
					FieldName:  "invalid_field_name",
					Template:   "https://www.example.com",
				},
			},
		},
	}
	invalidTemplate := model.RequestWrapper{
		PortalConfig: model.PortalConfig{
			HyperlinkMappings: []model.HyperlinkMapping{
				{
					AccpObject: "air_booking",
					FieldName:  "invalid_field_name",
					Template:   "",
				},
			},
		},
	}

	uc.reg.SetAppAccessPermission(constant.SaveHyperlinkPermission)
	err = uc.ValidateRequest(varlidRq)
	if err != nil {
		t.Errorf("[TestUpdatePortalConfig] request %v should be valid", varlidRq)
	}
	err = uc.ValidateRequest(invalidObject)
	if err == nil {
		t.Errorf("[TestUpdatePortalConfig] request %v should be invalid", invalidObject)
	}
	err = uc.ValidateRequest(invalidFieldName)
	if err == nil {
		t.Errorf("[TestUpdatePortalConfig] request %v should be invalid", invalidFieldName)
	}
	err = uc.ValidateRequest(invalidTemplate)
	if err == nil {
		t.Errorf("[TestUpdatePortalConfig] request %v should be invalid", invalidTemplate)
	}

	_, err = uc.Run(varlidRq)
	if err != nil {
		t.Errorf("[TestUpdatePortalConfig] error run GET request: %v", err)
	}

	res, err1 := getuc.Run(model.RequestWrapper{})
	if err1 != nil {
		t.Errorf("[TestUpdatePortalConfig] error run GET request: %v", err1)
	}
	if len(res.PortalConfig.HyperlinkMappings) != len(varlidRq.PortalConfig.HyperlinkMappings) {
		t.Errorf("[TestUpdatePortalConfig] invalid Get request. Should return %d mappings, but got %d", len(varlidRq.PortalConfig.HyperlinkMappings), len(res.PortalConfig.HyperlinkMappings))
	}
	log.Printf("get config res: %v", res)
	for i, m := range res.PortalConfig.HyperlinkMappings {
		expM := varlidRq.PortalConfig.HyperlinkMappings[i]
		if m.Template != expM.Template || m.FieldName != expM.FieldName || m.AccpObject != expM.AccpObject {
			t.Errorf("[TestUpdatePortalConfig] invalid mapping %+v, should be %+v ", m, expM)
		}
	}
}

type TestCfg struct {
	Tpl    string
	Fields []string
	Valid  bool
}
