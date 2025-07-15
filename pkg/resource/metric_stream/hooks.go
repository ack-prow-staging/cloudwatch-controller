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

package metric_stream

import (
	"context"
	"fmt"

	"github.com/aws-controllers-k8s/cloudwatch-controller/apis/v1alpha1"
	"github.com/aws-controllers-k8s/cloudwatch-controller/pkg/tags"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/cloudwatch/cloudwatchiface"
)

// getCloudWatchClient returns a CloudWatch client that can be used with the tags package
func (rm *resourceManager) getCloudWatchClient() cloudwatchiface.CloudWatchAPI {
	// Create a new CloudWatch client using a new session
	// This is needed because the resource manager uses the AWS SDK v2, but our tags package
	// uses the AWS SDK v1
	sess := session.Must(session.NewSession())
	return cloudwatch.New(sess)
}

// getTags returns the tags for a given MetricStream
func (rm *resourceManager) getTags(
	ctx context.Context,
	resourceARN string,
) []*v1alpha1.Tag {
	tagsClient := tags.NewTagsClient(rm.getCloudWatchClient())
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
	if latest.ko.Status.ACKResourceMetadata == nil || latest.ko.Status.ACKResourceMetadata.ARN == nil {
		return fmt.Errorf("ARN is nil for resource %s", *desired.ko.Spec.Name)
	}

	tagsClient := tags.NewTagsClient(rm.getCloudWatchClient())
	return tagsClient.SyncTags(
		ctx,
		desired.ko.Spec.Tags,
		latest.ko.Spec.Tags,
		string(*latest.ko.Status.ACKResourceMetadata.ARN),
	)
}
