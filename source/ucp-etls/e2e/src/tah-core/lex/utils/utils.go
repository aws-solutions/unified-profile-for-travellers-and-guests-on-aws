package utils

import (
	cfgModel "cloudrack-lambda-core/config/model"
	lex "cloudrack-lambda-core/lex"
	model "cloudrack-lambda-core/lex/model"
	s3 "cloudrack-lambda-core/s3"
	"encoding/json"
	"log"
)

var SLOT_ARTEFACT_KEY = "hotelSlots.json"
var BOT_NAME = "BookARoomBot"

func UpdateHotelSlots(hotel cfgModel.Hotel, env string, s3c s3.S3Config) error {
	cfgRecord := model.SlotConfig{
		EnumerationValues:      []model.SlotValue{},
		ValueSelectionStrategy: model.SLOT_SELECTION_STRATEGY_TOP_RESOLUTION,
		Name:                   "CloudrackProperties" + env,
		Description:            "List of Bookable properties",
	}
	s3c.Get(SLOT_ARTEFACT_KEY, &cfgRecord)
	if cfgRecord.Name == "" {
		cfgRecord.EnumerationValues = []model.SlotValue{model.SlotValue{
			Value:    hotel.Name,
			Synonyms: []string{"property " + hotel.Code},
		}}
		cfgRecord.ValueSelectionStrategy = model.SLOT_SELECTION_STRATEGY_TOP_RESOLUTION
		cfgRecord.Name = "CloudrackProperties" + env
		cfgRecord.Description = "List of Bookable properties"
	} else {
		isNew := true
		for i, slotValue := range cfgRecord.EnumerationValues {
			if slotValue.Synonyms[0] == "property "+hotel.Code {
				cfgRecord.EnumerationValues[i].Value = hotel.Name
				isNew = false
			}
		}
		if isNew {
			cfgRecord.EnumerationValues = append(cfgRecord.EnumerationValues, model.SlotValue{
				Value:    hotel.Name,
				Synonyms: []string{"property " + hotel.Code},
			})
		}
	}
	var configJson []byte
	configJson, err := json.Marshal(cfgRecord)
	if err != nil {
		log.Printf("[LEX][UTILS] Marshall error while saving slots configuration: %v", err)
		return err
	}
	err = s3c.SaveJson("", SLOT_ARTEFACT_KEY, configJson)
	if err != nil {
		log.Printf("[LEX][UTILS] error while saving slots config to S3: %v", err)
		return err
	}
	err = lex.Init(BOT_NAME + env).UpdateSlotType(cfgRecord)
	if err != nil {
		log.Printf("[LEX][UTILS] error while updating lex bot slot config: %v", err)
		return err
	}
	return nil
}
