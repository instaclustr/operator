/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package models

type KafkaUser struct {
	Password             string `json:"password,omitempty"`
	OverrideExistingUser bool   `json:"overrideExistingUser,omitempty"`
	SASLSCRAMMechanism   string `json:"saslScramMechanism,omitempty"`
	AuthMechanism        string `json:"authMechanism"`
	ClusterID            string `json:"clusterId"`
	InitialPermissions   string `json:"initialPermissions"`
	Username             string `json:"username"`
}

type CertificateRequest struct {
	ClusterID     string `json:"clusterId"`
	CSR           string `json:"csr"`
	KafkaUsername string `json:"kafkaUsername"`
	ValidPeriod   int    `json:"validPeriod"`
}
