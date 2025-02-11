// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package customerprofileslcs

import (
	"errors"
	"fmt"
	"strings"
	profilemodel "tah/upt/source/tah-core/customerprofilesmodel"
	"time"
)

/********************************************************
* Setup
*********************************************************/

type Domain struct {
	Name                 string
	NObjects             int64
	NProfiles            int64
	AllCounts            map[string]int64
	MatchingEnabled      bool
	RuleBaseStitchingOn  bool
	IsLowCostEnabled     bool
	CacheMode            CacheMode
	DefaultEncryptionKey string
	Created              time.Time
	LastUpdated          time.Time
	Tags                 map[string]string
}

var ErrProfileNotFound = fmt.Errorf("profile not found")

type MatchList struct {
	ConfidenceScore float64
	ProfileIds      []string
}

type ProfilePair struct {
	SourceID     string
	TargetID     string
	MergeContext ProfileMergeContext
}

type ObjectMapping struct {
	Name        string         `json:"name"`
	Version     string         `json:"version"`
	Description string         `json:"description"`
	Fields      []FieldMapping `json:"fields"`
}

type FieldMapping struct {
	Type        string   `json:"type"`
	Source      string   `json:"source"`
	Target      string   `json:"target"`
	Indexes     []string `json:"indexes"`
	Searcheable bool     `json:"searchable"`
	KeyOnly     bool     `json:"keyOnly"`
}

type FieldMappings []FieldMapping

type PaginationOptions struct {
	Page       int    `json:"page"`
	PageSize   int    `json:"pageSize"`
	ObjectType string `json:"objectType"`
}

type DeleteProfileOptions struct {
	DeleteHistory bool   `json:"deleteProfileHistoryOnDelete"`
	OperatorID    string `json:"operatorId"`
}

/************************************************************
* Identity Resolution
*******************************************************/

type RuleSet struct {
	Name                    string    `json:"name"`
	Rules                   []Rule    `json:"rules"`
	LatestVersion           int       `json:"latestVersion"`
	LastActivationTimestamp time.Time `json:"lastActivationTimestamp"`
}

type Rule struct {
	Index       int         `json:"index"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Conditions  []Condition `json:"conditions"`
}

func (r Rule) IsProfileLevel() bool {
	for _, cond := range r.Conditions {
		if cond.ExistingObjectType != PROFILE_OBJECT_TYPE_NAME || cond.IncomingObjectType != PROFILE_OBJECT_TYPE_NAME {
			return false
		}
	}
	return true
}

func (r Rule) IsObjectLevel() bool {
	for _, cond := range r.Conditions {
		if cond.ExistingObjectType == PROFILE_OBJECT_TYPE_NAME || cond.IncomingObjectType == PROFILE_OBJECT_TYPE_NAME {
			return false
		}
	}
	return true
}

func (r Rule) IsHybrid() bool {
	for _, cond := range r.Conditions {
		if (cond.ExistingObjectType == PROFILE_OBJECT_TYPE_NAME && cond.IncomingObjectType != PROFILE_OBJECT_TYPE_NAME) ||
			(cond.ExistingObjectType != PROFILE_OBJECT_TYPE_NAME && cond.IncomingObjectType == PROFILE_OBJECT_TYPE_NAME) {
			return true
		}
	}
	return false
}

/*
	There are 3 types of conditions:

	- SKIP condition skip the record entirely (meaning the rule does not apply to the record). In a SKIP condition
	  all fields apply to the incoming objects. The condition can specify the relation between a field in eth incoming record
	  and a static value defined in the SkipConditionValue field
	- FILTER condition applies to a match group (a group or profiles that satisfies all the MATCH conditions in a rule)
	- MATCH condition is true if the relation between IncomingObjectType IncomingObjectField ExistingObjectType ExistingObjectField
	  satisfies the Op operator for at least on of the (ExistingObjectType,ExistingObjectField) in the match table

*******************
*/
type Condition struct {
	Index         int    `json:"index"`
	ConditionType string `json:"conditionType"` //MATCH, FILTER, SKIP
	//contains incoming object type and field name
	IncomingObjectType  string `json:"incomingObjectType"` //_profile or <objectTypeName>
	IncomingObjectField string `json:"incomingObjectField"`
	Op                  string `json:"op"` //EQUALS, NOT_EQUALS, MATCH_REGERXP...
	//contains type ane value in case both fields are in the incoming object (filter or skip condition)
	IncomingObjectType2  string `json:"incomingObjectType2"` //_profile or <objectTypeName>
	IncomingObjectField2 string `json:"incomingObjectField2"`
	//contains type and field name for the existing data to be matched
	ExistingObjectType  string `json:"existingObjectType"` //_profile or <objectTypeName>
	ExistingObjectField string `json:"existingObjectField"`
	//static value or regexp provided for Skip condition
	SkipConditionValue Value `json:"skipConditionValue"`
	//static value or regexp provided for Filter condition
	FilterConditionValue Value `json:"filterConditionValue"`
	// this defines how fields are normalized for indexing. by default, we lowercase and trim
	IndexNormalization IndexNormalizationSettings `json:"indexNormalization"`
}

type IndexNormalizationSettings struct {
	Lowercase             bool `json:"lowercase"`
	Trim                  bool `json:"trim"`
	RemoveSpaces          bool `json:"removeSpaces"`
	RemoveNonNumeric      bool `json:"removeNonNumeric"`
	RemoveNonAlphaNumeric bool `json:"removeNonAlphaNumeric"`
	RemoveLeadingZeros    bool `json:"removeLeadingZeros"`
}

type RuleIndexRecord struct {
	ConnectID        string
	RuleIndex        int
	RuleName         string
	IndexValue1      string
	IndexValue2      string
	FilterFields     []FilterField
	FilterConditions []Condition
	//timestamp of the interaction being indexed
	Timestamp time.Time
}

type Value struct {
	ValueType   string `json:"valueType"`
	IntValue    int    `json:"intValue"`
	StringValue string `json:"stringValue"`
}

// this struct materialize the field name and values used in the Filter section of the IR table
type FilterField struct {
	Name  string
	Value string
}

type ProfileDataMap struct {
	ProfileId string
	Data      map[string]string
}

func (p ProfileDataMap) GetProfileValue(field string) string {
	return p.Data[field]
}

func (p *ProfileDataMap) SetProfileValue(field, value string) error {
	if p.Data == nil {
		p.Data = make(map[string]string)
	}
	if field == "" {
		return errors.New("field is empty")
	}
	p.Data[field] = value
	return nil
}

type RuleBasedMatch struct {
	RuleIndex string
	SourceID  string
	MatchIDs  []string
}

type RuleObjectPair struct {
	RuleID     string
	ObjectType string
}

type DynamoRule struct {
	Pk                      string    `json:"pk"`
	Sk                      string    `json:"sk"`
	Rule                    Rule      `json:"rule"`
	Latest                  int       `json:"latest"`
	LastActivationTimestamp time.Time `json:"last_activation_timestamp"`
}

/************************************************************
* Data Plane
*******************************************************/

type DynamoProfileRecord struct {
	Pk     string `json:"connect_id"`
	Sk     string `json:"record_type"`
	Domain string
	// Profile level fields
	LastUpdatedUnixMilliTimestamp int64  `json:"lastUpdated,omitempty"`
	ProfileId                     string `json:"profileId,omitempty"`
	AccountNumber                 string `json:"accountNumber,omitempty"`
	FirstName                     string `json:"firstName,omitempty"`
	MiddleName                    string `json:"middleName,omitempty"`
	LastName                      string `json:"lastName,omitempty"`
	BirthDate                     string `json:"birthDate,omitempty"`
	Gender                        string `json:"gender,omitempty"`
	PhoneNumber                   string `json:"phoneNumber,omitempty"`
	MobilePhoneNumber             string `json:"mobilePhoneNumber,omitempty"`
	HomePhoneNumber               string `json:"homePhoneNumber,omitempty"`
	BusinessPhoneNumber           string `json:"businessPhoneNumber,omitempty"`
	PersonalEmailAddress          string `json:"personalEmailAddress,omitempty"`
	BusinessEmailAddress          string `json:"businessEmailAddress,omitempty"`
	EmailAddress                  string `json:"emailAddress,omitempty"`
	Attributes                    map[string]string
	Address                       DynamoAddress `json:"address,omitempty"`
	MailingAddress                DynamoAddress `json:"mailingAddress,omitempty"`
	BillingAddress                DynamoAddress `json:"billingAddress,omitempty"`
	ShippingAddress               DynamoAddress `json:"shippingAddress,omitempty"`
	BusinessName                  string        `json:"businessName,omitempty"`
	// Object level fields
	ID   string `json:"id"`
	Type string `json:"type"`
}

type DynamoTravelerIndex struct {
	ConnectId  string `json:"connect_id"`
	RecordType string `json:"record_type"`
	TravelerId string `json:"profileId"`
}

type DynamoAddress struct {
	Address1   string `json:"Address1,omitempty"`
	Address2   string `json:"Address2,omitempty"`
	Address3   string `json:"Address3,omitempty"`
	Address4   string `json:"Address4,omitempty"`
	City       string `json:"City,omitempty"`
	State      string `json:"State,omitempty"`
	Province   string `json:"Province,omitempty"`
	PostalCode string `json:"PostalCode,omitempty"`
	Country    string `json:"Country,omitempty"`
}

type ProfileHistory struct {
	toMergeUPTID      string
	toUnmergeUPTID    string
	mergedIntoUPTID   string
	unmergedFromUPTID string
	Timestamp         time.Time
	Depth             int64
}

type ProfileMergeContext struct {
	Timestamp              time.Time `json:"timestamp"`
	MergeType              string    `json:"mergeType"`
	ConfidenceUpdateFactor float64   `json:"confidenceUpdateFactor"`
	RuleID                 string    `json:"ruleId"`
	RuleSetVersion         string    `json:"ruleSetVersion"`
	OperatorID             string    `json:"operatorId"`
	ToMergeConnectId       string    `json:"toMergeConnectId"`
	MergeIntoConnectId     string    `json:"mergeIntoConnectId"`
}

type MergeRequest struct {
	SourceID      string              `json:"sourceId"`
	TargetID      string              `json:"targetId"`
	Context       ProfileMergeContext `json:"context"`
	DomainName    string              `json:"domainName"`
	TransactionID string              `json:"transactionId"`
}

/************************************************************
* Control Plane
*******************************************************/

// user-defined options
type DomainOptions struct {
	AiIdResolutionOn       bool     `json:"aiIdResolutionOn"`
	RuleBaseIdResolutionOn bool     `json:"ruleBaseIdResolutionOn"` //true is rule based id resolution enabled
	ObjectTypePriority     []string `json:"objectTypePriority"`     //optionally define the relative priority between object type, this will define how profile-level data is updated
	DynamoMode             *bool    `json:"dynamoMode"`
	CustomerProfileMode    *bool    `json:"customerProfileMode"`
	FeatureSetVersion      int      `json:"featureSetVersion"`
	CpExportDestinationArn string   `json:"cpExportDestination"`
}

// Primary version of DomainConfig that should be used
type LcDomainConfig struct {
	Name                    string        `json:"name"`
	AiIdResolutionBucket    string        `json:"aiIdResolutionBucket"`
	QueueUrl                string        `json:"queueUrl"`
	KmsArn                  string        `json:"kmsArn"`
	EnvName                 string        `json:"envName"`
	AiIdResolutionGlueTable string        `json:"aiIdResolutionGlueTable"`
	Options                 DomainOptions `json:"options"`
	CustomerProfileMode     bool          `json:"customerProfileMode"`
	DynamoMode              bool          `json:"dynamoMode"`
	DomainVersion           int           `json:"domainVersion"`
	CompatibleVersion       int           `json:"CompatibleVersion"`
}

// Alternative version of DomainConfig with pk and sk, used specifically for
// saving and retrieving config items from Dynamo
type DynamoLcDomainConfig struct {
	Pk                      string        `json:"pk"`
	Sk                      string        `json:"sk"`
	AiIdResolutionBucket    string        `json:"aiIdResolutionBucket"`
	QueueUrl                string        `json:"queueUrl"`
	KmsArn                  string        `json:"kmsArn"`
	EnvName                 string        `json:"envName"`
	AiIdResolutionGlueTable string        `json:"aiIdResolutionGlueTable"`
	Options                 DomainOptions `json:"options"`
	CustomerProfileMode     bool          `json:"customerProfileMode"`
	DynamoMode              bool          `json:"dynamoMode"`
	DomainVersion           int           `json:"domainVersion"`
	CompatibleVersion       int           `json:"CompatibleVersion"`
}

type DynamoObjectMapping struct {
	Pk            string `json:"pk"`
	Sk            string `json:"sk"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	FieldMappings []FieldMapping
}

// CP Domain Params
type CpDomainParams struct {
	kmsArn    string
	streamArn string
	queueUrl  string
}

/*********************
* Event Streaming
**********************/

const SchemaVersion int64 = 1

type ChangeEventParams struct {
	EventType      string
	EventID        string
	Time           time.Time
	DomainName     string
	Profile        profilemodel.Profile
	IsRealTime     bool
	ObjectTypeName string                 // only included in create/update events
	ObjectID       string                 // only included in create/update events
	MergeID        string                 // only included in merge events
	MergedProfiles []profilemodel.Profile // only included in merge events
}

func (fms FieldMappings) GetSourceNames() []string {
	names := []string{}
	for _, v := range fms {
		split := strings.Split(v.Source, ".")
		names = append(names, split[len(split)-1])
	}
	return names
}

func (fms FieldMappings) GetTypeNames() []string {
	names := []string{}
	for _, v := range fms {
		split := strings.Split(v.Type, ".")
		names = append(names, split[len(split)-1])
	}
	return names
}

/*********************
 * Advanced Querying *
 *********************/

type BaseCriteria interface{}
type CriterionValue interface{}

type CriterionOperator string

const (
	EQ      CriterionOperator = "="
	NEQ     CriterionOperator = "!="
	LT      CriterionOperator = "<"
	LTE     CriterionOperator = "<="
	GT      CriterionOperator = ">"
	GTE     CriterionOperator = ">="
	IN      CriterionOperator = "IN"
	NOTIN   CriterionOperator = "NOT IN"
	LIKE    CriterionOperator = "LIKE"
	NOTLIKE CriterionOperator = "NOT LIKE"
)

type LogicalOperator string

const (
	AND LogicalOperator = "AND"
	OR  LogicalOperator = "OR"
)

type SearchCriterion struct {
	BaseCriteria
	Column   string
	Operator CriterionOperator
	Value    CriterionValue
}

type SearchOperator struct {
	BaseCriteria
	Operator LogicalOperator
}

type SearchGroup struct {
	BaseCriteria
	Criteria []BaseCriteria
}
