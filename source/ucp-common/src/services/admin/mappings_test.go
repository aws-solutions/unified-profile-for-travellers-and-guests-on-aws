// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package admin

import (
	"log"
	customerprofiles "tah/upt/source/storage"
	"tah/upt/source/tah-core/core"
	model "tah/upt/source/ucp-common/src/model/admin"
	"testing"
)

var accpMapping = customerprofiles.ObjectMapping{
	Name: "test_object",
	Fields: []customerprofiles.FieldMapping{
		{
			Type:        "STRING",
			Source:      "_source.firstName",
			Target:      "_profile.FirstName",
			Indexes:     []string{},
			Searcheable: true,
		},
		{
			Type:        "STRING",
			Source:      "_source.lastName",
			Target:      "_profile.LastName",
			Indexes:     []string{},
			Searcheable: true,
		},
		{
			Type:        "STRING",
			Source:      "_source.profile_id",
			Target:      "_profile.Attributes.profile_id",
			Indexes:     []string{"PROFILE"},
			Searcheable: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.booking_id",
			Target:  "bookingId",
			Indexes: []string{"UNIQUE"},
			KeyOnly: true,
		},
	},
}

var objectMapping = model.ObjectMapping{
	Name: "test_object",
	Fields: []model.FieldMapping{
		{
			Type:        "STRING",
			Source:      "_source.firstName",
			Target:      "_profile.FirstName",
			Indexes:     []string{},
			Searcheable: true,
		},
		{
			Type:        "STRING",
			Source:      "_source.lastName",
			Target:      "_profile.LastName",
			Indexes:     []string{},
			Searcheable: true,
		},
		{
			Type:        "STRING",
			Source:      "_source.profile_id",
			Target:      "_profile.Attributes.profile_id",
			Indexes:     []string{"PROFILE"},
			Searcheable: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.booking_id",
			Target:  "bookingId",
			Indexes: []string{"UNIQUE"},
			KeyOnly: true,
		},
	},
}

var objectMapping2 = model.ObjectMapping{
	Name: "test_object",
	Fields: []model.FieldMapping{
		{
			Type:        "STRING",
			Source:      "_source.firstName",
			Target:      "_profile.FirstName",
			Indexes:     []string{},
			Searcheable: false,
		},
		{
			Type:        "STRING",
			Source:      "_source.lastName",
			Target:      "_profile.LastName2",
			Indexes:     []string{},
			Searcheable: true,
		},
		{
			Type:        "STRING",
			Source:      "_source.profile_id",
			Target:      "_profile.Attributes.profile_id",
			Searcheable: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.booking_id",
			Target:  "bookingId2",
			Indexes: []string{"UNIQUE"},
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.lotalty_id",
			Target:  "lotaltyId",
			Indexes: []string{"UNIQUE"},
			KeyOnly: true,
		},
	},
}

var objectMapping3 = model.ObjectMapping{
	Name: "test_object2",
	Fields: []model.FieldMapping{
		{
			Type:        "STRING",
			Source:      "_source.firstName",
			Target:      "_profile.FirstName",
			Indexes:     []string{},
			Searcheable: true,
		},
		{
			Type:        "STRING",
			Source:      "_source.lastName",
			Target:      "_profile.LastName",
			Indexes:     []string{},
			Searcheable: true,
		},
		{
			Type:        "STRING",
			Source:      "_source.profile_id",
			Target:      "_profile.Attributes.profile_id",
			Indexes:     []string{"PROFILE"},
			Searcheable: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.booking_id",
			Target:  "bookingId",
			Indexes: []string{"UNIQUE"},
			KeyOnly: true,
		},
	},
}

func TestParseMappings(t *testing.T) {
	ucpMapping := ParseMappings([]customerprofiles.ObjectMapping{accpMapping})
	log.Printf("ucpMapping: %+v", ucpMapping)
	if len(ucpMapping) != 1 {
		t.Errorf("Expected 1 object mapping, got %v", len(ucpMapping))
	}
	if len(ucpMapping[0].Fields) != 4 {
		t.Errorf("Expected 4 fields, got %v", len(ucpMapping[0].Fields))
	}
	if ucpMapping[0].Fields[0].Type != "STRING" {
		t.Errorf("Expected STRING, got %v", ucpMapping[0].Fields[0].Type)
	}
	if ucpMapping[0].Fields[0].Source != "_source.firstName" {
		t.Errorf("Expected _source.firstName, got %v", ucpMapping[0].Fields[0].Source)
	}
	if ucpMapping[0].Fields[0].Target != "_profile.FirstName" {
		t.Errorf("Expected _profile.FirstName, got %v", ucpMapping[0].Fields[0].Target)
	}
	if ucpMapping[0].Fields[2].Indexes[0] != "PROFILE" {
		t.Errorf("Expected PROFILE, got %v", ucpMapping[0].Fields[3].Indexes[0])
	}
	if !ucpMapping[0].Fields[3].KeyOnly {
		t.Errorf("Expected true, got %v", ucpMapping[0].Fields[4].KeyOnly)
	}
	if !ucpMapping[0].Fields[0].Searcheable {
		t.Errorf("Expected true, got %v", ucpMapping[0].Fields[0].Searcheable)
	}
}

func TestHasOutdatedMappings(t *testing.T) {
	tx := core.NewTransaction(t.Name(), "", core.LogLevelDebug)
	hasOutdatedMapping, err := HasOutdatedMappings(tx, objectMapping, objectMapping)
	if err != nil {
		t.Errorf("HasOutdatedMappings Error: %v", err)
	}
	if hasOutdatedMapping {
		t.Errorf("Expected HasOutdatedMappings=false for identical mapping, got true")
	}
	_, err = HasOutdatedMappings(tx, objectMapping, objectMapping3)
	if err == nil {
		t.Errorf("Expected error when creating diff between differnt objects, got nil")
	}
	hasOutdatedMapping, _ = HasOutdatedMappings(tx, objectMapping, objectMapping2)
	if !hasOutdatedMapping {
		t.Errorf("Expected HasOutdatedMappings=true for different objects of the same type, got false")
	}

}

func TestOrganizeVyObjectType(t *testing.T) {
	organized := OrganizeByObjectType([]model.ObjectMapping{objectMapping2, objectMapping3})
	if len(organized) != 2 {
		t.Errorf("[OrganizeByObjectType] Expected 2 object, got %v", len(organized))
	}
	if organized[objectMapping2.Name].Name != objectMapping2.Name || organized[objectMapping3.Name].Name != objectMapping3.Name {
		t.Errorf("[OrganizeByObjectType] incorect object type map %+v", organized)
	}
}

func TestDiff(t *testing.T) {
	tx := core.NewTransaction(t.Name(), "", core.LogLevelDebug)
	newMap, updatedMap, err := DiffMappings(tx, objectMapping, objectMapping2)
	if err != nil {
		t.Errorf("DiffMappings Error: %v", err)
	}
	if len(newMap) != 1 {
		t.Errorf("Expected 1 new mapping, got %v", len(newMap))
	}
	if len(updatedMap) != 3 {
		t.Errorf("Expected 3 updated mapping, got %v", len(updatedMap))
	}
	if newMap[0].Source != "_source.lotalty_id" {
		t.Errorf("Expected new mapping to be on _source.lotalty_id, got %v", newMap[0].Source)
	}
	if updatedMap[0].Source != "_source.firstName" {
		t.Errorf("Expected updated mapping to be on _source.firstName, got %v", updatedMap[0].Source)
	}
	if updatedMap[1].Source != "_source.lastName" {
		t.Errorf("Expected updated mapping to be on _source.lastName, got %v", updatedMap[1].Source)
	}
	if updatedMap[2].Source != "_source.profile_id" {
		t.Errorf("Expected updated mapping to be on _source.profile_id, got %v", updatedMap[2].Source)
	}

	emptyObject := model.ObjectMapping{
		Name:   "",
		Fields: []model.FieldMapping{},
	}
	newMap, _, err = DiffMappings(tx, emptyObject, objectMapping)
	if err != nil {
		t.Errorf("DiffMappings should not return an error when effective mapping does not exist (new object). Got error: %v", err)
	}
	if len(newMap) != 4 {
		t.Errorf("DiffMappings should have 4  new mappings. got %v", len(newMap))
	}

}
