package lex

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"tah/core/s3"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lexmodelbuildingservice"
)

var SLOT_ARTEFACT_KEY = "hotelSlots.json"
var BOT_NAME = "BookARoomBot"
var SLOT_SELECTION_STRATEGY_TOP_RESOLUTION = "TOP_RESOLUTION"

/**
{
    "enumerationValues": [
        {
            "value": "Long Beach Resort, Mauritius"
        },
        {
            "value": "Intercontinental Hotel Chicago Magnificent Mile",
            "synonyms": [
                "Intercontinental in Chicago"
            ]
        },
        {
            "value": "Hyatt Regency Miami",
            "synonyms": [
                "Hyatt Miami",
                "Hyatt in Miami",
                "Regency Miami"
            ]
        }
    ],
    "valueSelectionStrategy": "TOP_RESOLUTION",
    "name": "CloudrackPropertiesdev",
    "description": "List of Bookable properties (Shoulld be dynamically extracted from CRS at build time)"
}
*/
type SlotConfig struct {
	EnumerationValues      []SlotValue `json:"enumerationValues"`
	ValueSelectionStrategy string      `json:"valueSelectionStrategy"`
	Name                   string      `json:"name"`
	Description            string      `json:"description"`
	Checksum               string      `json:"checksum,omitempty"`
}
type SlotValue struct {
	Value    string   `json:"value"`
	Synonyms []string `json:"synonyms"`
}

type LexService struct {
	Svc     *lexmodelbuildingservice.LexModelBuildingService
	BotName string
}

func Init(botName string) LexService {
	return LexService{
		Svc:     lexmodelbuildingservice.New(session.New()),
		BotName: botName,
	}
}

func (ls LexService) GetSlotType(name string) (SlotConfig, error) {
	input := &lexmodelbuildingservice.GetSlotTypeInput{
		Name:    aws.String(name),
		Version: aws.String("$LATEST"),
	}

	result, err := ls.Svc.GetSlotType(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case lexmodelbuildingservice.ErrCodeNotFoundException:
				fmt.Println(lexmodelbuildingservice.ErrCodeNotFoundException, aerr.Error())
			case lexmodelbuildingservice.ErrCodeLimitExceededException:
				fmt.Println(lexmodelbuildingservice.ErrCodeLimitExceededException, aerr.Error())
			case lexmodelbuildingservice.ErrCodeInternalFailureException:
				fmt.Println(lexmodelbuildingservice.ErrCodeInternalFailureException, aerr.Error())
			case lexmodelbuildingservice.ErrCodeBadRequestException:
				fmt.Println(lexmodelbuildingservice.ErrCodeBadRequestException, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
			return SlotConfig{}, errors.New(aerr.Error())
		}
		return SlotConfig{}, err
	}
	slotType := SlotConfig{
		Name:     *result.Name,
		Checksum: *result.Checksum,
	}
	return slotType, nil
}

func (ls LexService) UpdateSlotType(slotConfig SlotConfig) error {
	//retreiving existing slotype first as overriding required the checksum
	curSlotType, err := ls.GetSlotType(slotConfig.Name)
	input := &lexmodelbuildingservice.PutSlotTypeInput{
		Description:            aws.String(slotConfig.Description),
		EnumerationValues:      []*lexmodelbuildingservice.EnumerationValue{},
		Name:                   aws.String(slotConfig.Name),
		ValueSelectionStrategy: aws.String(slotConfig.ValueSelectionStrategy),
	}
	if err == nil && curSlotType.Checksum != "" {
		input.Checksum = aws.String(curSlotType.Checksum)
	}
	for _, slot := range slotConfig.EnumerationValues {
		enumValue := &lexmodelbuildingservice.EnumerationValue{
			Value:    aws.String(slot.Value),
			Synonyms: []*string{},
		}
		for _, syn := range slot.Synonyms {
			enumValue.Synonyms = append(enumValue.Synonyms, aws.String(syn))
		}
		input.EnumerationValues = append(input.EnumerationValues, enumValue)
	}
	log.Printf("[LEX] Updating slot type: %+v", input)
	_, err = ls.Svc.PutSlotType(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case lexmodelbuildingservice.ErrCodeConflictException:
				fmt.Println(lexmodelbuildingservice.ErrCodeConflictException, aerr.Error())
			case lexmodelbuildingservice.ErrCodeLimitExceededException:
				fmt.Println(lexmodelbuildingservice.ErrCodeLimitExceededException, aerr.Error())
			case lexmodelbuildingservice.ErrCodeInternalFailureException:
				fmt.Println(lexmodelbuildingservice.ErrCodeInternalFailureException, aerr.Error())
			case lexmodelbuildingservice.ErrCodeBadRequestException:
				fmt.Println(lexmodelbuildingservice.ErrCodeBadRequestException, aerr.Error())
			case lexmodelbuildingservice.ErrCodePreconditionFailedException:
				fmt.Println(lexmodelbuildingservice.ErrCodePreconditionFailedException, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
			return errors.New(aerr.Error())
		}
		fmt.Println(err.Error())
		return err
	}
	//updating the bot
	log.Printf("[LEX]Re building the bot %+v", input)
	return ls.RebuildBot()
}

func (ls LexService) RebuildBot() error {
	getIntput := &lexmodelbuildingservice.GetBotInput{
		VersionOrAlias: aws.String("$LATEST"),
		Name:           aws.String(ls.BotName),
	}
	getOutput, err := ls.Svc.GetBot(getIntput)
	if err != nil {
		log.Printf("[LEX] Error while Re building the bot %+v", err)
	}
	putInput := &lexmodelbuildingservice.PutBotInput{
		Name:                getOutput.Name,
		Checksum:            getOutput.Checksum,
		AbortStatement:      getOutput.AbortStatement,
		ChildDirected:       getOutput.ChildDirected,
		ClarificationPrompt: getOutput.ClarificationPrompt,
		Description:         getOutput.Description,
		DetectSentiment:     getOutput.DetectSentiment,
		//EnableModelImprovements:      getOutput.EnableModelImprovements,
		IdleSessionTTLInSeconds: getOutput.IdleSessionTTLInSeconds,
		Intents:                 getOutput.Intents,
		Locale:                  getOutput.Locale,
		//NluIntentConfidenceThreshold: getOutput.NluIntentConfidenceThreshold,
		VoiceId: getOutput.VoiceId,
	}
	_, err = ls.Svc.PutBot(putInput)
	if err != nil {
		log.Printf("[LEX] Error while Re building the bot %+v", err)
	}
	return err
}

func UpdateHotelSlots(hotel cfgModel.Hotel, env string, s3c s3.S3Config) error {
	cfgRecord := SlotConfig{
		EnumerationValues:      []SlotValue{},
		ValueSelectionStrategy: SLOT_SELECTION_STRATEGY_TOP_RESOLUTION,
		Name:                   "CloudrackProperties" + env,
		Description:            "List of Bookable properties",
	}
	s3c.Get(SLOT_ARTEFACT_KEY, &cfgRecord)
	if cfgRecord.Name == "" {
		cfgRecord.EnumerationValues = []SlotValue{SlotValue{
			Value:    hotel.Name,
			Synonyms: []string{"property " + hotel.Code},
		}}
		cfgRecord.ValueSelectionStrategy = SLOT_SELECTION_STRATEGY_TOP_RESOLUTION
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
			cfgRecord.EnumerationValues = append(cfgRecord.EnumerationValues, SlotValue{
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
