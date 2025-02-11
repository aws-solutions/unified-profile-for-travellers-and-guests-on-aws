// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package admin

import (
	"errors"
	"fmt"
	"log"
	"net/url"
	"regexp"
	"slices"
	"strings"

	customerprofiles "tah/upt/source/storage"
	"tah/upt/source/tah-core/core"
	model "tah/upt/source/ucp-backend/src/business-logic/model/common"
	"tah/upt/source/ucp-backend/src/business-logic/usecase/registry"
	constant "tah/upt/source/ucp-common/src/constant/admin"
	accpMappings "tah/upt/source/ucp-common/src/model/accp-mappings"
	"tah/upt/source/ucp-common/src/utils/utils"

	"github.com/aws/aws-lambda-go/events"
)

type UpdatePortalConfig struct {
	name string
	tx   *core.Transaction
	reg  *registry.Registry
}

func NewUpdatePortalConfig() *UpdatePortalConfig {
	return &UpdatePortalConfig{name: "UpdatePortalConfig"}
}

func (u *UpdatePortalConfig) Name() string {
	return u.name
}
func (u *UpdatePortalConfig) Tx() core.Transaction {
	return *u.tx
}
func (u *UpdatePortalConfig) SetTx(tx *core.Transaction) {
	u.tx = tx
}
func (u *UpdatePortalConfig) SetRegistry(reg *registry.Registry) {
	u.reg = reg
}
func (u *UpdatePortalConfig) Registry() *registry.Registry {
	return u.reg
}

func (u *UpdatePortalConfig) CreateRequest(req events.APIGatewayProxyRequest) (model.RequestWrapper, error) {
	return registry.CreateRequest(u, req)
}

func (u *UpdatePortalConfig) AccessPermission() constant.AppPermission {
	return constant.SaveHyperlinkPermission
}

func (u *UpdatePortalConfig) ValidateRequest(rq model.RequestWrapper) error {
	u.tx.Debug("[%v] Validating request", u.Name())
	mappings := rq.PortalConfig.HyperlinkMappings
	for _, mapping := range mappings {
		if mapping.FieldName == "" {
			return errors.New("fieldName is required")
		}
		if mapping.Template == "" {
			return errors.New("template is required")
		}
		if !constant.IsValidAccpRecord(mapping.AccpObject) {
			return fmt.Errorf("invalid accp record %v, should be within %v", mapping.AccpObject, constant.AccpRecordsNames())
		}
		utilFunction := customerprofiles.LCSUtils{Tx: *u.tx}
		objectMappings := []customerprofiles.ObjectMapping{}
		for _, businessObject := range constant.ACCP_RECORDS {
			accpRecName := businessObject.Name
			objMapping := accpMappings.ACCP_OBJECT_MAPPINGS[accpRecName]()
			objectMappings = append(objectMappings, objMapping)
		}
		_, objectFields := utilFunction.MappingsToDbColumns(objectMappings)
		fieldNames, ok := objectFields[mapping.AccpObject]
		if !ok {
			log.Printf("invalid object, should be within %v", objectFields)
			return fmt.Errorf("invalid object name, should be within %v", objectFields)
		}
		if !slices.Contains(fieldNames, mapping.FieldName) {
			log.Printf("invalid field name should be within %v", fieldNames)
			return fmt.Errorf("invalid field name, should be within %v", fieldNames)
		}
		err := validateHyperlinkTemplate(mapping.Template, fieldNames)
		if err != nil {
			return err
		}
	}
	return nil
}

func validateHyperlinkTemplate(tpl string, fieldNames []string) error {
	// Make sure the URL starts with "https://"
	if !strings.HasPrefix(tpl, "https://") {
		return errors.New("should be a valid HTTPS URL")
	}

	// Parse the URL to make sure it's valid
	parsedURL, err := url.ParseRequestURI(tpl)
	if err != nil {
		return errors.New("should be a valid URL")
	}

	// Make sure the host is not empty
	if parsedURL.Host == "" {
		return errors.New("URL should have a host")
	}

	brackets := 0
	for _, c := range tpl {
		if c == '{' {
			brackets++
		}
		if c == '}' {
			brackets--
		}
	}
	if brackets > 0 {
		return errors.New("invalid template syntax. validate that you have the right number of brackets. Ex: {{fieldName}}")
	}

	//validating template fields
	// Create a regular expression to match placeholders in the template
	re := regexp.MustCompile(`{{\w+}}`)

	// Find all matches in the template string
	matches := re.FindAllString(tpl, -1)

	for _, match := range matches {
		match = strings.Replace(match, "{{", "", -1)
		match = strings.Replace(match, "}}", "", -1)
		if !utils.ContainsString(fieldNames, match) {
			return fmt.Errorf("invalid placeholder %v. should be within %v", match, fieldNames)
		}
	}
	return nil
}

func (u *UpdatePortalConfig) Run(req model.RequestWrapper) (model.ResponseWrapper, error) {
	current := []model.DynamoHyperlinkMapping{}
	mappings := req.PortalConfig.HyperlinkMappings
	u.tx.Info("Saving mappings")
	dynamoMappings := []model.DynamoHyperlinkMapping{}
	for _, mapping := range mappings {
		dynamoMappings = append(dynamoMappings, model.DynamoHyperlinkMapping{
			Pk:         PORTAL_CONFIG_HYPERLINKS_PK,
			Sk:         PORTAL_CONFIG_HYPERLINKS_SK_PREFIX + mapping.AccpObject + mapping.FieldName,
			AccpObject: mapping.AccpObject,
			FieldName:  mapping.FieldName,
			Template:   mapping.Template,
		})

	}

	err := u.reg.PortalConfigDB.FindStartingWith(PORTAL_CONFIG_HYPERLINKS_PK, PORTAL_CONFIG_HYPERLINKS_SK_PREFIX, &current)
	if err != nil {
		return model.ResponseWrapper{}, err
	}
	toDelete := mappingsToDelete(current, dynamoMappings)
	err = u.reg.PortalConfigDB.DeleteMany(toDelete)
	if err != nil {
		return model.ResponseWrapper{}, err
	}

	err = u.reg.PortalConfigDB.SaveMany(dynamoMappings)
	if err != nil {
		return model.ResponseWrapper{}, err
	}
	return model.ResponseWrapper{PortalConfig: &model.PortalConfig{HyperlinkMappings: mappings}}, err
}

func mappingsToDelete(current []model.DynamoHyperlinkMapping, new []model.DynamoHyperlinkMapping) []model.DynamoHyperlinkMappingForDelete {
	toDelete := []model.DynamoHyperlinkMappingForDelete{}
	for _, c := range current {
		found := false
		for _, n := range new {
			if c.Sk == n.Sk {
				found = true
				break
			}
		}
		if !found {
			toDelete = append(toDelete, model.DynamoHyperlinkMappingForDelete{
				Pk: c.Pk,
				Sk: c.Sk,
			})
		}
	}

	return toDelete
}

func (u *UpdatePortalConfig) CreateResponse(res model.ResponseWrapper) (events.APIGatewayProxyResponse, error) {
	return registry.CreateResponse(u, res)
}
