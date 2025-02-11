// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package upt_sdk

import (
	"errors"
	"fmt"
	"log"
	"reflect"
	profilemodel "tah/upt/source/tah-core/customerprofilesmodel"
	constants "tah/upt/source/ucp-common/src/constant/admin"
	admModel "tah/upt/source/ucp-common/src/model/admin"
	model "tah/upt/source/ucp-common/src/model/traveler"
	"tah/upt/source/ucp-common/src/utils/utils"
	"time"
)

// Creates a test user and initiate auth
func (h *UptHandler) SetupTestAuth(domain string) error {
	user, pwd, err := h.CreateDomainGroupAndUser(domain)
	if err != nil {
		log.Printf("Error creating admin user: %s", err)
		return err
	}
	adminRefreshToken, adminAccessToken, err := h.Login(user, pwd)
	if err != nil {
		log.Printf("Error authenticating admin user: %s", err)
		return err
	}
	h.SetAuthTokens(adminRefreshToken, adminAccessToken)
	return nil
}

// Delete the test user and unset token
func (h UptHandler) CleanupAuth(domain string) error {
	err := h.DeleteGroupAndUser(domain)
	if err != nil {
		log.Printf("Error deleting admin user: %s", err)
		return err
	}
	appAccessGroup := fmt.Sprintf("/%x", constants.AdminPermission)
	err = h.CognitoClient.DeleteGroup("app-global-custom" + domain + appAccessGroup)
	if err != nil {
		log.Println("Error deleting cognito group")
		return err
	}
	h.UnsetAuthTokens()
	return nil
}

// create domain and poll the async Api while waiting for domain creation to complete
func (h UptHandler) CreateDomainAndWait(domain string, timeoutSeconds int) error {
	log.Printf("Creating domain %s", domain)
	createRes, err := h.CreateDomain(domain)
	if err != nil {
		log.Printf("Error creating domain: %s", err)
		return err
	}
	log.Printf("Create domain response: %+v", createRes)
	evtId := createRes.AsyncEvent.EventID
	uc := createRes.AsyncEvent.Usecase
	err = h.WaitForAsyncEvent(evtId, uc, timeoutSeconds)
	if err != nil {
		log.Printf("Error waiting for domain creation: %s", err)
		return err
	}
	log.Printf("Domain created successfully")
	return nil
}

// Wait for an async event to complete by polling the Async api
func (h UptHandler) WaitForAsyncEvent(asyncEventId, asyncEventUsecase string, timeoutSeconds int) error {
	if asyncEventId == "" || asyncEventUsecase == "" {
		return errors.New("invalid async event id or usecase")
	}
	log.Printf("[WaitForAsyncEvent] Waiting for async event %s-%s to complete with timeout %d", asyncEventId, asyncEventUsecase, timeoutSeconds)
	status := ""
	counter := 0
	for status != "success" && status != "failed" && counter < timeoutSeconds/5 {
		log.Printf("[WaitForAsyncEvent] GetAsyncEventStatus try %d", counter)
		res, err := h.GetAsyncEventStatus(asyncEventId, asyncEventUsecase)
		if err != nil {
			log.Printf("Error getting async event status: %s", err)
			return err
		}
		log.Printf("Async event response: %+v", res)
		if res.AsyncEvent.Status == "" {
			log.Printf("Async event status is empty")
			return errors.New("async event status is empty")
		}
		counter++
		status = res.AsyncEvent.Status
		log.Printf("[WaitForAsyncEvent] [%v] status for try %d: %s", asyncEventUsecase, counter, status)
		if status != "success" {
			log.Printf("[WaitForAsyncEvent] Waiting 5 seconds")
			time.Sleep(5 * time.Second)
		}
	}
	if status == "failed" {
		log.Printf("[WaitForAsyncEvent] async request failed")
		return errors.New("async event returned with a failed status")
	}
	if counter >= timeoutSeconds/5 {
		log.Printf("[WaitForAsyncEvent] timeout exceeded")
		return errors.New("timeout while waiting for async event completion")
	}
	return nil
}

// Delete domain and wait for domain deletion to complete by polling the async API
func (h UptHandler) DeleteDomainAndWait(domain string) error {
	log.Printf("Deleting domain %s", domain)
	createRes, err := h.DeleteDomain(domain)
	if err != nil {
		log.Printf("Error deleting domain: %s", err)
		return err
	}
	evtId := createRes.AsyncEvent.EventID
	uc := createRes.AsyncEvent.Usecase
	err = h.WaitForAsyncEvent(evtId, uc, 300)
	if err != nil {
		log.Printf("Error waiting for domain deletion: %s", err)
		return err
	}
	log.Printf("Domain deleting successfully")
	return nil
}

// Run all Glue Jobs and Wait for completion. the function first waits for all jobs to be in "RUNNING" status
// and then waits for all jobs to be in "SUCCEEDED" status. the function returns an error is timeouts or at least
// one job returns "FAILED"
func (h UptHandler) RunAllJobsAndWait(domain string, timeout int) error {
	log.Printf("Running all jobs")
	res, err := h.RunAllJobs(domain)
	if err != nil {
		log.Printf("Error running all jobs: %s", err)
		return err
	}
	log.Printf("Run all jobs response: %+v", res)
	log.Printf("Waiting for jobs to run")
	err = h.WaitForJobToRun(domain, timeout)
	if err != nil {
		log.Printf("Error waiting for jobs to run: %s", err)
		return err
	}
	log.Printf("Wait for jobs to complete")
	err = h.WaitForJobsToComplete(domain, timeout)
	if err != nil {
		log.Printf("Error waiting for jobs to complete: %s", err)
		return err
	}
	log.Printf("Jobs run successfully")
	return nil
}

// Wait for all jobs to be in "RUNNING" status
func (h UptHandler) WaitForJobToRun(domainName string, timeout int) error {
	return h.WaitForJobsToBuInStatus(domainName, "RUNNING", timeout)
}

// Wait for all jobs to be in "SUCCEEDED" status
func (h UptHandler) WaitForJobsToComplete(domainName string, timeout int) error {
	return h.WaitForJobsToBuInStatus(domainName, "SUCCEEDED", timeout)
}

// Wait for all jobs to be in a specific non-FAILED status (returns an error is this status is failed)
func (h UptHandler) WaitForJobsToBuInStatus(domainName, status string, timeout int) error {
	if status == "" {
		return errors.New("non empty status must be provided")
	}
	if status == "FAILED" {
		return errors.New("status must be different from FAILED")
	}

	areAllInStatus := false
	counter := 0
	for !areAllInStatus && counter/5 < timeout {
		counter++
		res, err := h.GetJobsStatus(domainName)
		if err != nil {
			log.Printf("Error getting domain config: %s", err)
			return err
		}
		areAllInStatus, err = h.checkAllJobsStatus(res.AwsResources.Jobs, status)
		if err != nil {
			log.Printf("Error checking jobs status: %s", err)
			return err
		}
		if !areAllInStatus {
			log.Printf("Waiting 5 second")
			time.Sleep(5 * time.Second)
		}
	}
	log.Printf("All jobs are in status %s", status)
	return nil
}

func (h UptHandler) checkAllJobsStatus(jobs []admModel.JobSummary, status string) (bool, error) {
	if len(jobs) == 0 {
		return false, errors.New("no job status returned")
	}
	for _, job := range jobs {
		if job.Status != status {
			log.Printf("Job %s is not in status %s", job.JobName, status)
			return false, nil
		}
		if job.Status == "FAILED" {
			log.Printf("Job %s is in status FAILED. Retuning", job.JobName)
			return false, fmt.Errorf("job %s failed", job.JobName)
		}
	}
	return true, nil
}

// create domain and poll the async Api while waiting for domain creation to complete
func (h UptHandler) RebuildCacheAndWait(domainName string, cacheMode string, timeoutSeconds int) (string, error) {
	log.Printf("Rebuilding cache for domain %s", domainName)
	res, err := h.RebuildCache(domainName, cacheMode)
	if err != nil {
		log.Printf("Error rebuilding cache: %s", err)
		return "", err
	}
	log.Printf("rebuild cache response: %+v", res)
	evtId := res.AsyncEvent.EventID
	uc := res.AsyncEvent.Usecase
	err = h.WaitForAsyncEvent(evtId, uc, timeoutSeconds)
	if err != nil {
		log.Printf("Error waiting for rebuild cache: %s", err)
		return "", err
	}
	log.Printf("Cache tasks executed successfully")
	return evtId, nil
}

func (h UptHandler) CreatePrivacySearchAndWait(domainName string, connectIds []string, timeout int) error {
	res, err := h.CreatePrivacySearch(domainName, connectIds)
	if err != nil {
		log.Printf("Error creating privacy search: %s", err)
		return err
	}
	log.Printf("create privacy search response: %+v", res)
	evtId := res.AsyncEvent.EventID
	uc := res.AsyncEvent.Usecase
	err = h.WaitForAsyncEvent(evtId, uc, timeout)
	if err != nil {
		log.Printf("Error waiting for privacy search: %s", err)
		return err
	}
	log.Printf("Privacy search created successfully")
	return nil
}

func (h UptHandler) PrivacyPurgeProfilesAndWait(domainName string, connectIds []string, timeout int) error {
	_, err := h.StartProfilePurge(domainName, connectIds)
	if err != nil {
		log.Printf("Error purging profiles: %s", err)
		return err
	}
	it := 0
	for it*5 < timeout {
		res, err := h.GetIsPurgeRunning(domainName)
		if err != nil {
			log.Printf("Error getting purge status: %v", err)
			continue
		}
		if res.PrivacyPurgeStatus.IsPurgeRunning {
			log.Printf("Purge is still running")
		} else {
			log.Printf("Purge completed successfully")
			return nil
		}
		log.Printf("Waiting 5 seconds")
		time.Sleep(5 * time.Second)
		it += 1
	}
	log.Printf("Privacy purge executed successfully")
	return nil
}

func (h UptHandler) GetSpecificProfileObject(domainName, objectId string, profileId string, objectTypeName string) (profilemodel.ProfileObject, error) {
	res, err := h.RetreiveProfile(domainName, profileId, constants.AccpRecordsNames())
	if err != nil {
		return profilemodel.ProfileObject{}, err
	}
	profiles := utils.SafeDereference(res.Profiles)
	if len(profiles) == 0 {
		return profilemodel.ProfileObject{}, fmt.Errorf("no profile exists with id %s", profileId)
	}
	profile := profiles[0]
	obj := profilemodel.ProfileObject{}

	log.Printf("[GetSpecificProfileObject] looking for object type %s in profile %+v", objectTypeName, profile)
	switch objectTypeName {
	case constants.ACCP_RECORD_AIR_LOYALTY:
		obj, err = createProfileObject(toInterfaceSlice(profile.AirLoyaltyRecords), objectId, constants.ACCP_RECORD_AIR_LOYALTY)
	case constants.ACCP_RECORD_ANCILLARY:
		obj, err = createProfileObject(toInterfaceSlice(profile.AncillaryServiceRecords), objectId, constants.ACCP_RECORD_ANCILLARY)
	case constants.ACCP_RECORD_CLICKSTREAM:
		obj, err = createProfileObject(toInterfaceSlice(profile.ClickstreamRecords), objectId, constants.ACCP_RECORD_CLICKSTREAM)
	case constants.ACCP_RECORD_HOTEL_BOOKING:
		obj, err = createProfileObject(toInterfaceSlice(profile.HotelBookingRecords), objectId, constants.ACCP_RECORD_HOTEL_BOOKING)
	case constants.ACCP_RECORD_HOTEL_LOYALTY:
		obj, err = createProfileObject(toInterfaceSlice(profile.HotelLoyaltyRecords), objectId, constants.ACCP_RECORD_HOTEL_LOYALTY)
	case constants.ACCP_RECORD_AIR_BOOKING:
		obj, err = createProfileObject(toInterfaceSlice(profile.AirBookingRecords), objectId, constants.ACCP_RECORD_AIR_BOOKING)
	case constants.ACCP_RECORD_LOYALTY_TX:
		obj, err = createProfileObject(toInterfaceSlice(profile.LoyaltyTxRecords), objectId, constants.ACCP_RECORD_LOYALTY_TX)
	default:
		log.Printf("[GetSpecificProfileObject] unsupported object type %s", objectTypeName)
		return profilemodel.ProfileObject{}, fmt.Errorf("object type %s not supported", objectTypeName)
	}
	log.Printf("[GetSpecificProfileObject] found: %+v", obj)
	return obj, err
}

func toInterfaceSlice(slice interface{}) []model.AccpObject {
	s := reflect.ValueOf(slice)
	if s.Kind() != reflect.Slice {
		panic("toInterfaceSlice() given a non-slice type")
	}

	ret := make([]model.AccpObject, s.Len())

	for i := 0; i < s.Len(); i++ {
		ret[i] = s.Index(i).Interface().(model.AccpObject)
	}

	return ret
}

func createProfileObject(objs []model.AccpObject, objectId string, objectType string) (profilemodel.ProfileObject, error) {
	for _, obj := range objs {
		if obj.ID() == objectId {
			ai, err := buildAttributeInterface(obj)
			if err != nil {
				return profilemodel.ProfileObject{}, err
			}
			return profilemodel.ProfileObject{
				ID:                  obj.ID(),
				Type:                objectType,
				AttributesInterface: ai,
			}, nil
		}
	}
	return profilemodel.ProfileObject{}, fmt.Errorf("object of type '%s' with  id '%s' not found in profile", objectType, objectId)
}

func createHotelLoyaltyProfileObject(objs []model.HotelLoyalty, objectId string) (profilemodel.ProfileObject, error) {
	for _, obj := range objs {
		if obj.AccpObjectID == objectId {
			ai, err := buildAttributeInterface(obj)
			if err != nil {
				return profilemodel.ProfileObject{}, err
			}
			return profilemodel.ProfileObject{
				ID:                  obj.AccpObjectID,
				Type:                "hotel_loyalty",
				AttributesInterface: ai,
			}, nil
		}
	}
	return profilemodel.ProfileObject{}, fmt.Errorf("object not found")
}

func (h UptHandler) MergeProfilesAndWait(domainName, connectIdOne, connectIdTwo string, timeout int) (model.Traveller, error) {
	log.Printf("Merging profiles %s and %s in domain %s and wait %d s", connectIdOne, connectIdTwo, domainName, timeout)
	_, err := h.MergeProfiles(domainName, connectIdOne, connectIdTwo)
	if err != nil {
		log.Printf("Error merging profiles %s and %s: %v", connectIdOne, connectIdTwo, err)
		return model.Traveller{}, err
	}
	err = h.WaitForProfilesToBeMerged(domainName, connectIdOne, connectIdTwo, timeout)
	if err != nil {
		log.Printf("Error waiting for profiles %s and %s to be merged : %v", connectIdOne, connectIdTwo, err)
		return model.Traveller{}, err
	}
	res, err := h.RetreiveProfile(domainName, connectIdOne, constants.AccpRecordsNames())
	if err != nil {
		log.Printf("Error retrieving remaining profile profile %s after merge: %v", connectIdOne, err)
		return model.Traveller{}, err
	}
	profiles := utils.SafeDereference(res.Profiles)
	if len(profiles) == 0 {
		log.Printf("Error retrieving remaining profile profile %s after merge: no profile returned", connectIdOne)
		return model.Traveller{}, errors.New("no profile returned after merge")
	}
	return profiles[0], nil
}

func (h UptHandler) UnmergeProfilesAndWait(domainName, connectIdOne, connectIdTwo string, timeout int) (model.Traveller, error) {
	log.Printf("Unmerging profiles %s and %s in domain %s and wait %d s", domainName, connectIdOne, connectIdTwo, timeout)
	_, err := h.UnmergeProfiles(domainName, connectIdOne, connectIdTwo)
	if err != nil {
		log.Printf("Error unmerging profiles %s and %s: %v", connectIdOne, connectIdTwo, err)
		return model.Traveller{}, err
	}
	recreatedProfile, err := h.WaitForProfileRecreation(domainName, connectIdTwo, timeout)
	if err != nil {
		log.Printf("Error waiting for profile to be recreated: %v after unmerge", recreatedProfile)
		return model.Traveller{}, err
	}
	log.Printf("successfully unmerged profiles %s and %s", connectIdOne, connectIdTwo)
	return recreatedProfile, nil
}

// Wait for profile with a given ID to be merged
func (h UptHandler) WaitForProfilesToBeMerged(domainName, connectId1, connectId2 string, timeout int) error {
	log.Printf("Waiting %v seconds for profiles %s and %s to be merged", timeout, connectId1, connectId2)
	it := 0
	for it*5 < timeout {
		res1, err := h.RetreiveProfile(domainName, connectId1, []string{})
		if err != nil {
			log.Printf("error retrieving profile %s %v", connectId1, err)
			return errors.New("error retrieving profile")
		}
		res2, err := h.RetreiveProfile(domainName, connectId2, []string{})
		if err != nil {
			log.Printf("error retrieving profile %s %v", connectId2, err)
			return errors.New("error retrieving profile")
		}
		if res1.Profiles != nil && res2.Profiles != nil {
			p1 := *res1.Profiles
			p2 := *res2.Profiles
			if p1[0].ConnectID == p2[0].ConnectID {
				log.Printf("Profile %s and %s have been merged. returning", connectId1, connectId2)
				return nil
			}
		}
		log.Printf("profiles have not been merged: Waiting 5 seconds")
		time.Sleep(5 * time.Second)
		it += 1
	}
	return errors.New("profiles were not merged: timeout expired")
}

// Wait for profile with a given ID to be deleted
func (h UptHandler) WaitForProfileToBeDelete(domainName, connectID string, timeout int) error {
	log.Printf("Waiting %v seconds for profile with connect ID %v to be deleted", timeout, connectID)
	it := 0
	for it*5 < timeout {
		res, err := h.RetreiveProfile(domainName, connectID, []string{})
		if err != nil && res.Profiles == nil {
			log.Printf("Profile %s no longer exist. returning", connectID)
			return nil
		}
		log.Printf("profile still exist: Waiting 5 seconds")
		time.Sleep(5 * time.Second)
		it += 1
	}
	return errors.New("could not find profile ID: timeout expired")
}

// Wait for profile with a given connect ID to be re-created (after unmerge)
func (h UptHandler) WaitForProfileRecreation(domainName, connectID string, timeout int) (model.Traveller, error) {
	log.Printf("Waiting %v seconds for profile %v re-creation (after unmerge)", timeout, connectID)
	it := 0
	for it*5 < timeout {
		res, err := h.RetreiveProfile(domainName, connectID, constants.AccpRecordsNames())
		profiles := utils.SafeDereference(res.Profiles)
		if err == nil && profiles[0].ConnectID == connectID {
			log.Printf("Profile %s exist. returning", connectID)
			return profiles[0], nil
		}
		log.Printf("Profile %s has not been recreated yet: Waiting 5 seconds", connectID)
		time.Sleep(5 * time.Second)
		it += 1
	}
	return model.Traveller{}, fmt.Errorf("could not find profile with connect ID %s: timeout expired", connectID)
}

// Wait for profile with a given ID to be created
func (h UptHandler) WaitForProfile(domainName, profileID string, timeout int) (model.Traveller, error) {
	log.Printf("Waiting %v seconds for profile ID %v creation", timeout, profileID)
	it := 0
	for it*5 < timeout {
		res, err := h.SearchProfile(domainName, "travellerId", []string{profileID})
		if err != nil {
			log.Printf("[WaitForProfile] SearchProfiles error: %v", err)
			return model.Traveller{}, err
		}
		profiles := utils.SafeDereference(res.Profiles)
		if len(profiles) > 0 {
			profile := profiles[0]
			log.Printf("Found profile ID: %v. (ConnectId: %v)", profile.TravellerID, profile.ConnectID)
			res, err := h.RetreiveProfile(domainName, profile.ConnectID, constants.AccpRecordsNames())
			if err == nil {
				profiles := utils.SafeDereference(res.Profiles)
				if len(profiles) == 0 {
					return model.Traveller{}, fmt.Errorf("no profile exists with id %s", profileID)
				}
				fullProfile := profiles[0]
				return fullProfile, nil
			}
		}
		log.Printf("Not found. Waiting 5 seconds")
		time.Sleep(5 * time.Second)
		it += 1
	}
	return model.Traveller{}, fmt.Errorf("could not find profile ID %s: timeout expired", profileID)
}

// Wait for profile with a given ID to be created
func (h UptHandler) WaitForProfileWithCondition(domainName, profileID string, timeout int, cond func(profile model.Traveller) bool) (model.Traveller, error) {
	log.Printf("Waiting %v seconds for profile ID %v creation", timeout, profileID)
	it := 0
	for it*5 < timeout {
		res, err := h.SearchProfile(domainName, "travellerId", []string{profileID})
		if err != nil {
			log.Printf("[WaitForProfile] SearchProfiles error: %v", err)
			return model.Traveller{}, err
		}
		profiles := utils.SafeDereference(res.Profiles)
		if len(profiles) > 0 {
			profile := profiles[0]
			log.Printf("Found profile ID: %v. (ConnectId: %v)", profile.TravellerID, profile.ConnectID)
			res, err := h.RetreiveProfile(domainName, profile.ConnectID, constants.AccpRecordsNames())
			if err != nil {
				log.Printf("Profile found by search but error retrieving it: %v", err)
				return model.Traveller{}, err
			}
			profiles := utils.SafeDereference(res.Profiles)
			if len(profiles) == 0 {
				log.Printf("Profile found by search but empty response when retrieving it")
				return model.Traveller{}, fmt.Errorf("no profile exists with id %s", profileID)
			}
			fullProfile := profiles[0]
			if cond(fullProfile) {
				log.Printf("Profile found and meets condition. returning")
				return fullProfile, nil
			} else {
				log.Printf("Profile found but does not meet condition")
			}
		}
		log.Printf("Waiting 5 seconds")
		time.Sleep(5 * time.Second)
		it += 1
	}
	return model.Traveller{}, fmt.Errorf("could not find profile ID %s: timeout expired", profileID)
}

// Wait for profile with a an object of a given type and ID to be created for a given profile ID
func (h UptHandler) WaitForProfileObject(domainName, objectId string, profileId string, objectType string, timeout int) (profilemodel.ProfileObject, error) {
	log.Printf("Waiting %v seconds for profile object type '%s' with ID '%v' creation", timeout, objectType, objectId)
	it := 0
	var err error
	for it*5 < timeout {
		profObject, err := h.GetSpecificProfileObject(domainName, objectId, profileId, objectType)
		if err != nil {
			log.Printf("GetSpecificProfileObject error: %v", err)
			log.Printf("Not found. Waiting 5 seconds")
			time.Sleep(5 * time.Second)
			it += 1
		} else {
			log.Printf("Profile object: %+v", profObject.Attributes)
			return profObject, nil
		}
	}
	return profilemodel.ProfileObject{}, fmt.Errorf("could not find profile object ID: timeout expired last error: %v", err)
}

// this function re-creates the attribute interface from a deserialized struct
func buildAttributeInterface(obj interface{}) (map[string]interface{}, error) {
	objType := reflect.TypeOf(obj)
	objValue := reflect.ValueOf(obj)

	// Ensure input is a struct
	if objType.Kind() != reflect.Struct {
		return nil, fmt.Errorf("input is not a struct")
	}

	result := make(map[string]interface{})

	for i := 0; i < objType.NumField(); i++ {
		field := objType.Field(i)
		fieldValue := objValue.Field(i).Interface()
		fieldName := field.Name

		result[fieldName] = fieldValue
	}

	return result, nil
}
