// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package iam

import (
	"encoding/json"
	"log"
	"strings"

	"tah/upt/source/tah-core/core"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	awsIam "github.com/aws/aws-sdk-go/service/iam"
)

type Config struct {
	Svc *awsIam.IAM
}

type PolicyDocument struct {
	Version   string
	Statement []StatementEntry
}

type TrusPolicyDocument struct {
	Version   string
	Statement []TrustStatementEntry
}

type StatementEntry struct {
	Effect   string
	Action   []string
	Resource []string
}
type TrustStatementEntry struct {
	Effect    string
	Action    []string
	Principal Principal
}

type Principal struct {
	Service string
}

func Init(solutionId, solutionVersion string) Config {
	session := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	client := core.CreateClient(solutionId, solutionVersion)
	return Config{
		Svc: awsIam.New(session, &aws.Config{HTTPClient: client}),
	}
}

func (c Config) CreateRole(name string, principal string) (string, error) {
	policyJson, err := json.Marshal(buildAssumeRolePolicy(principal))
	if err != nil {
		return "", err
	}
	intput := &awsIam.CreateRoleInput{
		RoleName:                 aws.String(name),
		AssumeRolePolicyDocument: aws.String(string(policyJson)),
	}
	out, err2 := c.Svc.CreateRole(intput)
	if err2 != nil {
		return "", err2
	}
	return *out.Role.Arn, nil
}

func (c Config) DeleteRole(name string) error {
	policyArns, err := c.ListAttachedPolicies(name)
	if err != nil {
		return err
	}
	err = c.DetachRolePolicies(name, policyArns)
	if err != nil {
		return err
	}
	intput := &awsIam.DeleteRoleInput{
		RoleName: aws.String(name),
	}
	_, err2 := c.Svc.DeleteRole(intput)
	return err2
}

func (c Config) DeleteRoleIfExists(name string) error {
	policyArns, err1 := c.ListAttachedPolicies(name)
	if err1 != nil && !isNoSuchEntityError(err1) {
		return err1
	}
	err1 = c.DetachRolePolicies(name, policyArns)
	if err1 != nil && !isNoSuchEntityError(err1) {
		return err1
	}
	intput := &awsIam.DeleteRoleInput{
		RoleName: aws.String(name),
	}
	_, err2 := c.Svc.DeleteRole(intput)
	if err2 != nil && !isNoSuchEntityError(err2) {
		return err2
	}
	return nil
}

func (c Config) ListAttachedPolicies(roleName string) ([]string, error) {
	policies := []string{}
	intput := &awsIam.ListAttachedRolePoliciesInput{
		RoleName: aws.String(roleName),
	}
	out, err := c.Svc.ListAttachedRolePolicies(intput)
	if err != nil {
		return []string{}, err
	}
	for _, policy := range out.AttachedPolicies {
		if policy != nil && policy.PolicyArn != nil {
			policies = append(policies, *policy.PolicyArn)
		}
	}
	for out.IsTruncated != nil && *out.IsTruncated {
		intput.Marker = out.Marker
		out, err := c.Svc.ListAttachedRolePolicies(intput)
		if err != nil {
			return []string{}, err
		}
		for _, policy := range out.AttachedPolicies {
			if policy != nil && policy.PolicyArn != nil {
				policies = append(policies, *policy.PolicyArn)
			}
		}
	}
	return policies, nil
}

func (c Config) RoleExists(name string) string {
	intput := &awsIam.GetRoleInput{
		RoleName: aws.String(name),
	}
	role, err := c.Svc.GetRole(intput)
	if err != nil {
		return ""
	}
	return *role.Role.Arn
}

func (c Config) SearchPoliciesByRegexp(policyName string) ([]string, error) {
	policies := []string{}
	//escaping regexp char
	policyName = strings.ReplaceAll(policyName, "-", "-")
	intput := &awsIam.ListPoliciesInput{
		PathPrefix: aws.String("/" + policyName + "/"),
	}
	log.Printf("Searching for policies with regexp: %+v ", intput)
	out, err := c.Svc.ListPolicies(intput)
	if err != nil {
		log.Printf("Erroor listing policies %+v ", err)
		return []string{}, err
	}
	log.Printf("Found: %+v ", out)

	for _, pol := range out.Policies {
		if pol != nil && pol.Arn != nil {
			policies = append(policies, *pol.Arn)
		}
	}
	return policies, nil
}

func (c Config) PolicyExists(account string, policyName string) string {
	intput := &awsIam.GetPolicyInput{
		PolicyArn: aws.String(BuildPolicyArn(account, policyName)),
	}
	log.Printf("[PolicyExists] request %+v ", intput)
	out, err := c.Svc.GetPolicy(intput)
	if err != nil {
		log.Printf("[PolicyExists] GetPolicy Error: %+v ", err)
		return ""
	}
	log.Printf("[PolicyExists] GetPolicy Response: %+v ", out)
	if out.Policy != nil && out.Policy.Arn != nil {
		log.Printf("[PolicyExists] returning policy ARN %+v ", *out.Policy.Arn)
		return *out.Policy.Arn
	}
	return ""
}

func (c Config) CreateRoleWithActionsResource(name string, principal string, actions []string, resources []string) (string, string, error) {
	return c.CreateRoleWithPolicy(name, principal, buildPolicyFromActionsResource(actions, resources))
}

func (c Config) CreateRoleWithActionResource(name string, principal string, action string, resource string) (string, string, error) {
	return c.CreateRoleWithPolicy(name, principal, buildPolicyFromActionResource(action, resource))
}

func (c Config) CreateRoleWithPolicy(name string, principal string, doc PolicyDocument) (string, string, error) {
	roleArn, err := c.CreateRole(name, principal)
	if err != nil {
		return "", "", err
	}
	//Recreating policy to account ffor potential changes
	log.Printf("Checking if policy %s exists in account %v", name, AccountFromRoleArn(roleArn))
	policyArn := c.PolicyExists(AccountFromRoleArn(roleArn), name)
	if policyArn != "" {
		log.Printf("Found policy: %v. Deleting it", policyArn)
		err = c.DeletePolicy(policyArn)
		if err != nil {
			return "", "", err
		}
		log.Printf("Policy %v successfully deleted", policyArn)
	}
	log.Printf("Creating new policy")
	policyArn, err = c.CreatePolicy(name, doc)
	if err != nil {
		return roleArn, "", err
	}
	log.Printf("Attache policy %s to role", policyArn)
	err = c.AttachRolePolicy(name, policyArn)
	if err != nil {
		return "", "", err
	}
	return roleArn, policyArn, c.AttachRolePolicy(name, policyArn)
}

func (c Config) AttachRolePolicy(roleName string, policyArn string) error {
	intput := &awsIam.AttachRolePolicyInput{
		PolicyArn: aws.String(policyArn),
		RoleName:  aws.String(roleName),
	}
	_, err := c.Svc.AttachRolePolicy(intput)
	return err
}

func (c Config) DetachRolePolicy(roleName string, policyArn string) error {
	intput := &awsIam.DetachRolePolicyInput{
		PolicyArn: aws.String(policyArn),
		RoleName:  aws.String(roleName),
	}
	_, err := c.Svc.DetachRolePolicy(intput)
	return err
}

func (c Config) DetachRolePolicies(roleName string, policyArns []string) error {
	for _, arn := range policyArns {
		err := c.DetachRolePolicy(roleName, arn)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c Config) CreatePolicy(name string, doc PolicyDocument) (string, error) {
	policyJson, err := json.Marshal(doc)
	if err != nil {
		return "", err
	}
	intput := &awsIam.CreatePolicyInput{
		PolicyName:     aws.String(name),
		PolicyDocument: aws.String(string(policyJson)),
	}
	out, err2 := c.Svc.CreatePolicy(intput)
	if err2 != nil {
		return "", err2
	}
	return *out.Policy.Arn, nil
}

func (c Config) DeletePolicy(policyArn string) error {
	intput := &awsIam.DeletePolicyInput{
		PolicyArn: aws.String(policyArn),
	}
	_, err2 := c.Svc.DeletePolicy(intput)
	return err2
}

func AccountFromRoleArn(roleArn string) string {
	parts := strings.Split(roleArn, ":")
	if len(parts) < 5 {
		return ""
	}
	return parts[4]
}

func BuildPolicyArn(account string, policyName string) string {
	return "arn:aws:iam::" + account + ":policy/" + policyName
}

func buildAssumeRolePolicy(principal string) TrusPolicyDocument {
	return TrusPolicyDocument{
		Version: "2012-10-17",
		Statement: []TrustStatementEntry{
			{
				Effect: "Allow",
				Principal: Principal{
					Service: principal,
				},
				Action: []string{"sts:AssumeRole"},
			},
		},
	}
}

func buildPolicyFromActionResource(action string, resource string) PolicyDocument {
	return buildPolicyFromActionsResource([]string{action}, []string{resource})
}
func buildPolicyFromActionsResource(actions []string, resources []string) PolicyDocument {
	return PolicyDocument{
		Version: "2012-10-17",
		Statement: []StatementEntry{
			{
				Effect:   "Allow",
				Resource: resources,
				Action:   actions,
			},
		},
	}
}

// More information on handling AWS errors: https://pkg.go.dev/github.com/aws/aws-sdk-go/aws/awserr
func isNoSuchEntityError(err error) bool {
	if aerr, ok := err.(awserr.Error); ok {
		switch aerr.Code() {
		case awsIam.ErrCodeNoSuchEntityException:
			return true
		default:
			return false
		}
	}
	return false
}
