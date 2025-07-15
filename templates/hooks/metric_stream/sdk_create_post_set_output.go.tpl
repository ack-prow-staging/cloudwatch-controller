	if ko.Spec.Tags != nil {
		// Mark the resource as needing to be synced
		// This will trigger a requeue and allow the tags to be synced
		// in the next reconciliation loop
		ackcondition.SetSynced(&resource{ko}, corev1.ConditionFalse, nil, nil)
	}