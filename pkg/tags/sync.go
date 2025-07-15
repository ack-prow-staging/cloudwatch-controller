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

package tags

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/cloudwatch/cloudwatchiface"

	"github.com/aws-controllers-k8s/cloudwatch-controller/apis/v1alpha1"
)

// TagsGetterInterface provides methods for getting AWS tags for a resource
type TagsGetterInterface interface {
	GetResourceTags(
		ctx context.Context,
		resourceARN string,
	) ([]*v1alpha1.Tag, error)
}

// TagsSyncerInterface provides methods for syncing AWS tags for a resource
type TagsSyncerInterface interface {
	SyncTags(
		ctx context.Context,
		desired []*v1alpha1.Tag,
		latest []*v1alpha1.Tag,
		resourceARN string,
	) error
}

// TagsClient implements the TagsGetterInterface and TagsSyncerInterface
type TagsClient struct {
	sdkapi cloudwatchiface.CloudWatchAPI
}

// NewTagsClient returns a new TagsClient
func NewTagsClient(sdkapi cloudwatchiface.CloudWatchAPI) *TagsClient {
	return &TagsClient{
		sdkapi: sdkapi,
	}
}

// GetResourceTags returns the tags for a resource
func (tc *TagsClient) GetResourceTags(
	ctx context.Context,
	resourceARN string,
) ([]*v1alpha1.Tag, error) {
	input := &cloudwatch.ListTagsForResourceInput{
		ResourceARN: aws.String(resourceARN),
	}

	resp, err := tc.sdkapi.ListTagsForResourceWithContext(ctx, input)
	if err != nil {
		return nil, err
	}

	tags := make([]*v1alpha1.Tag, 0, len(resp.Tags))
	for _, tag := range resp.Tags {
		tags = append(tags, &v1alpha1.Tag{
			Key:   tag.Key,
			Value: tag.Value,
		})
	}

	return tags, nil
}

// compareStringPtrs compares two string pointers for equality
func compareStringPtrs(a, b *string) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

// SyncTags syncs the tags for a resource
func (tc *TagsClient) SyncTags(
	ctx context.Context,
	desired []*v1alpha1.Tag,
	latest []*v1alpha1.Tag,
	resourceARN string,
) error {
	toAdd := []*v1alpha1.Tag{}
	toDelete := []*string{}

	// Build a map of the latest tags
	latestTagMap := make(map[string]*v1alpha1.Tag, len(latest))
	for _, tag := range latest {
		latestTagMap[*tag.Key] = tag
	}

	// Find tags to add or update
	for _, desiredTag := range desired {
		latestTag, found := latestTagMap[*desiredTag.Key]
		if !found || !compareStringPtrs(latestTag.Value, desiredTag.Value) {
			toAdd = append(toAdd, desiredTag)
		}
		// Remove from the map so we can find tags to delete
		delete(latestTagMap, *desiredTag.Key)
	}

	// Any tags left in the map need to be deleted
	for key := range latestTagMap {
		toDelete = append(toDelete, aws.String(key))
	}

	// Add tags
	if len(toAdd) > 0 {
		tagInput := &cloudwatch.TagResourceInput{
			ResourceARN: aws.String(resourceARN),
			Tags:        convertTags(toAdd),
		}
		_, err := tc.sdkapi.TagResourceWithContext(ctx, tagInput)
		if err != nil {
			return err
		}
	}

	// Delete tags
	if len(toDelete) > 0 {
		untagInput := &cloudwatch.UntagResourceInput{
			ResourceARN: aws.String(resourceARN),
			TagKeys:     toDelete,
		}
		_, err := tc.sdkapi.UntagResourceWithContext(ctx, untagInput)
		if err != nil {
			return err
		}
	}

	return nil
}

// convertTags converts from ACK tags to CloudWatch tags
func convertTags(tags []*v1alpha1.Tag) []*cloudwatch.Tag {
	cwTags := make([]*cloudwatch.Tag, len(tags))
	for i, tag := range tags {
		cwTags[i] = &cloudwatch.Tag{
			Key:   tag.Key,
			Value: tag.Value,
		}
	}
	return cwTags
}
