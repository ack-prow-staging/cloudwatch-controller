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

package metricstream

import (
	"context"
	"fmt"

	ackv1alpha1 "github.com/aws-controllers-k8s/runtime/apis/core/v1alpha1"

	"github.com/aws-controllers-k8s/cloudwatch-controller/pkg/tags"
)

// getTags returns the tags for a given MetricStream
func (rm *resourceManager) getTags(
	ctx context.Context,
	resourceARN string,
) []*ackv1alpha1.Tag {
	tagsClient := tags.NewTagsClient(rm.sdkapi)
	tags, err := tagsClient.GetResourceTags(ctx, resourceARN)
	if err != nil {
		return nil
	}
	return tags
}

// syncTags synchronizes tags between the desired and latest resource
func (rm *resourceManager) syncTags(
	ctx context.Context,
	latest *resource,
	desired *resource,
) error {
	if latest.ko.Status.ARN == nil {
		return fmt.Errorf("ARN is nil for resource %s", *desired.ko.Spec.Name)
	}

	tagsClient := tags.NewTagsClient(rm.sdkapi)
	return tagsClient.SyncTags(
		ctx,
		desired.ko.Spec.Tags,
		latest.ko.Spec.Tags,
		*latest.ko.Status.ARN,
	)
}
