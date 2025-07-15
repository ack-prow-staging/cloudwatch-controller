	if ko.Status.ARN != nil {
		ko.Spec.Tags = rm.getTags(ctx, *ko.Status.ARN)
	}