/*
Copyright 2017 The Kubernetes Authors.

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

package source

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"sigs.k8s.io/external-dns/endpoint"
)

func TestGetTTLFromAnnotations(t *testing.T) {
	for _, tc := range []struct {
		title       string
		annotations map[string]string
		expectedTTL endpoint.TTL
		expectedErr error
	}{
		{
			title:       "TTL annotation not present",
			annotations: map[string]string{"foo": "bar"},
			expectedTTL: endpoint.TTL(0),
			expectedErr: nil,
		},
		{
			title:       "TTL annotation value is not a number",
			annotations: map[string]string{ttlAnnotationKey: "foo"},
			expectedTTL: endpoint.TTL(0),
			expectedErr: fmt.Errorf("\"foo\" is not a valid TTL value"),
		},
		{
			title:       "TTL annotation value is empty",
			annotations: map[string]string{ttlAnnotationKey: ""},
			expectedTTL: endpoint.TTL(0),
			expectedErr: fmt.Errorf("\"\" is not a valid TTL value"),
		},
		{
			title:       "TTL annotation value is negative number",
			annotations: map[string]string{ttlAnnotationKey: "-1"},
			expectedTTL: endpoint.TTL(0),
			expectedErr: fmt.Errorf("TTL value must be between [%d, %d]", ttlMinimum, ttlMaximum),
		},
		{
			title:       "TTL annotation value is too high",
			annotations: map[string]string{ttlAnnotationKey: fmt.Sprintf("%d", 1<<32)},
			expectedTTL: endpoint.TTL(0),
			expectedErr: fmt.Errorf("TTL value must be between [%d, %d]", ttlMinimum, ttlMaximum),
		},
		{
			title:       "TTL annotation value is set correctly using integer",
			annotations: map[string]string{ttlAnnotationKey: "60"},
			expectedTTL: endpoint.TTL(60),
			expectedErr: nil,
		},
		{
			title:       "TTL annotation value is set correctly using duration (whole)",
			annotations: map[string]string{ttlAnnotationKey: "10m"},
			expectedTTL: endpoint.TTL(600),
			expectedErr: nil,
		},
		{
			title:       "TTL annotation value is set correcly using duration (fractional)",
			annotations: map[string]string{ttlAnnotationKey: "20.5s"},
			expectedTTL: endpoint.TTL(20),
			expectedErr: nil,
		},
	} {
		t.Run(tc.title, func(t *testing.T) {
			ttl, err := getTTLFromAnnotations(tc.annotations)
			assert.Equal(t, tc.expectedTTL, ttl)
			assert.Equal(t, tc.expectedErr, err)
		})
	}
}

func TestGetSRVRecordTypeValuesFromAnnotations(t *testing.T) {
	svcName := "testSvc"
	for _, tc := range []struct {
		title       string
		annotations map[string]string
		expectedPriority int64
		expectedWeight int64
		expectedPort int64
		expectedErr error
	}{
		{
			title: "SRV Priority annotation does not exist",
			annotations: map[string]string{"foo": "bar"},
			expectedPriority: 0,
			expectedWeight: 0,
			expectedPort: 0,
			expectedErr: fmt.Errorf("must specify priority value for SRV record. service \"%v\"", svcName),
		},
		{
			title: "SRV Priority annotation is not a number",
			annotations: map[string]string{srvRecordTypePriorityAnnotationKey: "foo"},
			expectedPriority: 0,
			expectedWeight: 0,
			expectedPort: 0,
			expectedErr: fmt.Errorf("priorty value must be int number, got \"foo\". service \"%v\"", svcName),
		},
		{
			title: "SRV Weight annotation does not exist",
			annotations: map[string]string{srvRecordTypePriorityAnnotationKey: "0"},
			expectedPriority: 0,
			expectedWeight: 0,
			expectedPort: 0,
			expectedErr: fmt.Errorf("must specify weight value for SRV record. service \"%v\"", svcName),
		},
		{
			title: "SRV Weight annotation is not a number",
			annotations: map[string]string{
				srvRecordTypePriorityAnnotationKey: "0",
				srvRecordTypeWeightAnnotationKey: "foo",
			},
			expectedPriority: 0,
			expectedWeight: 0,
			expectedPort: 0,
			expectedErr: fmt.Errorf("weight value must be int number, got \"foo\". service \"%v\"", svcName),
		},
		{
			title: "SRV Port annotation does not exist",
			annotations: map[string]string{
				srvRecordTypePriorityAnnotationKey: "0",
				srvRecordTypeWeightAnnotationKey: "0",
			},
			expectedPriority: 0,
			expectedWeight: 0,
			expectedPort: 0,
			expectedErr: fmt.Errorf("must specify port value for SRV record. service \"%v\"", svcName),
		},
		{
			title: "SRV Port annotation is not a number",
			annotations: map[string]string{
				srvRecordTypePriorityAnnotationKey: "0",
				srvRecordTypeWeightAnnotationKey: "0",
				srvRecordTypePortAnnotationKey: "foo",
			},
			expectedPriority: 0,
			expectedWeight: 0,
			expectedPort: 0,
			expectedErr: fmt.Errorf("port value must be int number, got \"foo\". service \"%v\"", svcName),
		},
		{
			title: "SRV Port annotation is a negative number",
			annotations: map[string]string{
				srvRecordTypePriorityAnnotationKey: "0",
				srvRecordTypeWeightAnnotationKey: "0",
				srvRecordTypePortAnnotationKey: "-1",
			},
			expectedPriority: 0,
			expectedWeight: 0,
			expectedPort: 0,
			expectedErr: fmt.Errorf("port value must be between [%d, %d], got \"-1\". service \"%v\"", portMinimum, portMaximum, svcName),
		},
		{
			title: "SRV Port annotation is too high",
			annotations: map[string]string{
				srvRecordTypePriorityAnnotationKey: "0",
				srvRecordTypeWeightAnnotationKey: "0",
				srvRecordTypePortAnnotationKey: "100000",
			},
			expectedPriority: 0,
			expectedWeight: 0,
			expectedPort: 0,
			expectedErr: fmt.Errorf("port value must be between [%d, %d], got \"100000\". service \"%v\"", portMinimum, portMaximum, svcName),
		},
		{
			title: "SRV annotations is set correctly",
			annotations: map[string]string{
				srvRecordTypePriorityAnnotationKey: "5",
				srvRecordTypeWeightAnnotationKey: "7",
				srvRecordTypePortAnnotationKey: "443",
			},
			expectedPriority: 5,
			expectedWeight: 7,
			expectedPort: 443,
			expectedErr: nil,
		},
	} {
		t.Run(tc.title, func(t *testing.T) {
			priority, weight, port, err := getSRVRecordTypeValuesFromAnnotations(svcName, tc.annotations)
			assert.Equal(t, tc.expectedPriority, priority)
			assert.Equal(t, tc.expectedWeight, weight)
			assert.Equal(t, tc.expectedPort, port)
			assert.Equal(t, tc.expectedErr, err)
		})
	}
}

func TestSuitableType(t *testing.T) {
	for _, tc := range []struct {
		target, recordType, expected string
	}{
		{"8.8.8.8", "", "A"},
		{"foo.example.org", "", "CNAME"},
		{"bar.eu-central-1.elb.amazonaws.com", "", "CNAME"},
	} {

		recordType := suitableType(tc.target)

		if recordType != tc.expected {
			t.Errorf("expected %s, got %s", tc.expected, recordType)
		}
	}
}
