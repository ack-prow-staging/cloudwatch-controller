// Copyright Amazon.com Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may
// not use this file except in compliance with the License. A copy of the
// License is located at
//
//     http://aws.amazon.com/apache2.0/
//
// or in the "license" file accompanying this file. This file is distributed
// on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
// express or implied. See the License for the specific language governing
// permissions and limitations under the License.

package tags_test

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/cloudwatch/cloudwatchiface"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/aws-controllers-k8s/cloudwatch-controller/apis/v1alpha1"
	"github.com/aws-controllers-k8s/cloudwatch-controller/pkg/tags"
)

type mockCloudWatchClient struct {
	cloudwatchiface.CloudWatchAPI
	mock.Mock
}

func (m *mockCloudWatchClient) ListTagsForResourceWithContext(ctx context.Context, input *cloudwatch.ListTagsForResourceInput, opts ...request.Option) (*cloudwatch.ListTagsForResourceOutput, error) {
	args := m.Called(ctx, input)
	return args.Get(0).(*cloudwatch.ListTagsForResourceOutput), args.Error(1)
}

func (m *mockCloudWatchClient) TagResourceWithContext(ctx context.Context, input *cloudwatch.TagResourceInput, opts ...request.Option) (*cloudwatch.TagResourceOutput, error) {
	args := m.Called(ctx, input)
	return args.Get(0).(*cloudwatch.TagResourceOutput), args.Error(1)
}

func (m *mockCloudWatchClient) UntagResourceWithContext(ctx context.Context, input *cloudwatch.UntagResourceInput, opts ...request.Option) (*cloudwatch.UntagResourceOutput, error) {
	args := m.Called(ctx, input)
	return args.Get(0).(*cloudwatch.UntagResourceOutput), args.Error(1)
}

func TestGetResourceTags(t *testing.T) {
	mockClient := &mockCloudWatchClient{}
	tagsClient := tags.NewTagsClient(mockClient)
	ctx := context.Background()
	resourceARN := "arn:aws:cloudwatch:us-west-2:123456789012:metric-stream/test-stream"

	expectedTags := []*cloudwatch.Tag{
		{
			Key:   aws.String("key1"),
			Value: aws.String("value1"),
		},
		{
			Key:   aws.String("key2"),
			Value: aws.String("value2"),
		},
	}

	mockClient.On("ListTagsForResourceWithContext", ctx, &cloudwatch.ListTagsForResourceInput{
		ResourceARN: aws.String(resourceARN),
	}).Return(&cloudwatch.ListTagsForResourceOutput{
		Tags: expectedTags,
	}, nil)

	tags, err := tagsClient.GetResourceTags(ctx, resourceARN)

	assert.NoError(t, err)
	assert.Len(t, tags, 2)
	assert.Equal(t, "key1", *tags[0].Key)
	assert.Equal(t, "value1", *tags[0].Value)
	assert.Equal(t, "key2", *tags[1].Key)
	assert.Equal(t, "value2", *tags[1].Value)
	mockClient.AssertExpectations(t)
}

func TestSyncTags(t *testing.T) {
	mockClient := &mockCloudWatchClient{}
	tagsClient := tags.NewTagsClient(mockClient)
	ctx := context.Background()
	resourceARN := "arn:aws:cloudwatch:us-west-2:123456789012:metric-stream/test-stream"

	// Test case: Add new tags, update existing tags, and remove old tags
	desired := []*v1alpha1.Tag{
		{
			Key:   aws.String("key1"),
			Value: aws.String("new-value1"), // Updated value
		},
		{
			Key:   aws.String("key3"),
			Value: aws.String("value3"), // New tag
		},
	}

	latest := []*v1alpha1.Tag{
		{
			Key:   aws.String("key1"),
			Value: aws.String("value1"),
		},
		{
			Key:   aws.String("key2"),
			Value: aws.String("value2"), // Will be removed
		},
	}

	// Expect TagResource to be called with the new and updated tags
	mockClient.On("TagResourceWithContext", ctx, &cloudwatch.TagResourceInput{
		ResourceARN: aws.String(resourceARN),
		Tags: []*cloudwatch.Tag{
			{
				Key:   aws.String("key1"),
				Value: aws.String("new-value1"),
			},
			{
				Key:   aws.String("key3"),
				Value: aws.String("value3"),
			},
		},
	}).Return(&cloudwatch.TagResourceOutput{}, nil)

	// Expect UntagResource to be called with the removed tag keys
	mockClient.On("UntagResourceWithContext", ctx, &cloudwatch.UntagResourceInput{
		ResourceARN: aws.String(resourceARN),
		TagKeys:     []*string{aws.String("key2")},
	}).Return(&cloudwatch.UntagResourceOutput{}, nil)

	err := tagsClient.SyncTags(ctx, desired, latest, resourceARN)

	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}
