// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package admin

import (
	"errors"
	"strings"
	customerprofiles "tah/upt/source/storage"
	"tah/upt/source/tah-core/core"
	model "tah/upt/source/ucp-common/src/model/admin"
)

func DiffMappings(tx core.Transaction, effectiveMapping model.ObjectMapping, latestMapping model.ObjectMapping) ([]model.FieldMapping, []model.FieldMapping, error) {
	tx.Debug("Computing diff between effective mapping %v and latest mapping %s", effectiveMapping.Name, latestMapping.Name)
	if effectiveMapping.Name != latestMapping.Name && effectiveMapping.Name != "" {
		return nil, nil, errors.New("Cannot diff mappings for different objects")
	}
	effectiveBySource := organizeBySource(effectiveMapping.Fields)
	newMappings := []model.FieldMapping{}
	updatedMappings := []model.FieldMapping{}
	for _, latest := range latestMapping.Fields {
		if effective, hasIt := effectiveBySource[latest.Source]; hasIt {
			updatedMappings = AddUpdatedMapping(tx, updatedMappings, effective, latest)
		} else {
			tx.Debug("New mapping: %s", latest.Source)
			newMappings = append(newMappings, latest)
		}
	}
	return newMappings, updatedMappings, nil
}

func AddUpdatedMapping(tx core.Transaction, updatedMappings []model.FieldMapping, effective model.FieldMapping, latest model.FieldMapping) []model.FieldMapping {
	if effective.Target != latest.Target && !latest.KeyOnly {
		tx.Debug("Mapping %s target has been updated from %v to %v", latest.Source, effective.Target, latest.Target)
		updatedMappings = append(updatedMappings, latest)
		return updatedMappings
	}
	if effective.Type != latest.Type {
		tx.Debug("Mapping %s Type has been updated from %v to %v", latest.Source, effective.Type, latest.Type)
		updatedMappings = append(updatedMappings, latest)
		return updatedMappings
	}
	if strings.Join(effective.Indexes, "-") != strings.Join(latest.Indexes, "-") {
		tx.Debug("Mapping %s Indexes has been updated from %v to %v", latest.Source, effective.Indexes, latest.Indexes)
		updatedMappings = append(updatedMappings, latest)
		return updatedMappings
	}
	if effective.KeyOnly != latest.KeyOnly {
		tx.Debug("Mapping %s KeyOnly has been updated from %v to %v", latest.Source, effective.KeyOnly, latest.KeyOnly)
		updatedMappings = append(updatedMappings, latest)
		return updatedMappings
	}
	//mappings with indexes are always serchable
	if effective.Searcheable != latest.Searcheable && len(latest.Indexes) == 0 {
		tx.Debug("Mapping %s Searcheable has been updated from %v to %v", latest.Source, effective.Searcheable, latest.Searcheable)
		updatedMappings = append(updatedMappings, latest)
		return updatedMappings
	}
	return updatedMappings
}

func organizeBySource(mappings []model.FieldMapping) map[string]model.FieldMapping {
	organized := map[string]model.FieldMapping{}
	for _, mapping := range mappings {
		organized[mapping.Source] = mapping
	}
	return organized
}

func OrganizeByObjectType(mappings []model.ObjectMapping) map[string]model.ObjectMapping {
	organized := map[string]model.ObjectMapping{}
	for _, mapping := range mappings {
		organized[mapping.Name] = mapping
	}
	return organized
}

func HasOutdatedMappings(tx core.Transaction, effectiveMapping model.ObjectMapping, latestMapping model.ObjectMapping) (bool, error) {
	newMappings, updatedMapping, err := DiffMappings(tx, effectiveMapping, latestMapping)
	return len(newMappings) > 0 || len(updatedMapping) > 0, err
}

func ParseMappings(profileMappings []customerprofiles.ObjectMapping) []model.ObjectMapping {
	modelMappings := []model.ObjectMapping{}
	for _, profileMapping := range profileMappings {
		modelMappings = append(modelMappings, ParseMapping(profileMapping))
	}
	return modelMappings
}
func ParseMapping(profileMapping customerprofiles.ObjectMapping) model.ObjectMapping {
	return model.ObjectMapping{
		Name:    profileMapping.Name,
		Version: profileMapping.Version,
		Fields:  ParseFieldMappings(profileMapping.Fields),
	}
}
func ParseFieldMappings(profileMappings []customerprofiles.FieldMapping) []model.FieldMapping {
	modelMappings := []model.FieldMapping{}
	for _, profileMapping := range profileMappings {
		modelMappings = append(modelMappings, ParseFieldMapping(profileMapping))
	}
	return modelMappings
}
func ParseFieldMapping(profileMappings customerprofiles.FieldMapping) model.FieldMapping {
	mapping := model.FieldMapping{
		Type:        profileMappings.Type,
		Source:      profileMappings.Source,
		Target:      profileMappings.Target,
		Indexes:     profileMappings.Indexes,
		Searcheable: profileMappings.Searcheable,
		KeyOnly:     profileMappings.KeyOnly,
	}
	return mapping
}
