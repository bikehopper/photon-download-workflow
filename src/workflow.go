package photon_download_workflow

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

func PhotonDownload(ctx workflow.Context) error {
	workflow.GetLogger(ctx).Info("Schedule workflow started.", "StartTime", workflow.Now(ctx))
	so := &workflow.SessionOptions{
		CreationTimeout:  time.Minute,
		ExecutionTimeout: 24 * time.Hour,
	}
	sessionCtx, err := workflow.CreateSession(ctx, so)
	if err != nil {
		return err
	}
	defer workflow.CompleteSession(sessionCtx)

	ao := workflow.ActivityOptions{
		TaskQueue:           "photon-download",
		StartToCloseTimeout: 6 * time.Hour,
		RetryPolicy: &temporal.RetryPolicy{
			BackoffCoefficient: 2.0,
			MaximumAttempts:    3,
		},
	}
	activitySessionCtx := workflow.WithActivityOptions(sessionCtx, ao)

	var a *PhotonDownloadActivityObject

	var newArchiveActivityResult CheckForNewArchiveActivityResult
	var downloadArchiveActivityResult DownloadArchiveActivityResult
	var uploadArchiveActivityResult UploadArchiveActivityResult

	err = workflow.ExecuteActivity(activitySessionCtx, a.CheckForNewArchiveActivity).Get(activitySessionCtx, &newArchiveActivityResult)
	if err != nil {
		return err
	}

	if newArchiveActivityResult.NewArchiveAvailable {
		err = workflow.ExecuteActivity(activitySessionCtx, a.DownloadArchiveActivity).Get(activitySessionCtx, &downloadArchiveActivityResult)
		if err != nil {
			return err
		}
		err = workflow.ExecuteActivity(activitySessionCtx, a.UploadArchiveActivity, &downloadArchiveActivityResult).Get(activitySessionCtx, &uploadArchiveActivityResult)
		if err != nil {
			return err
		}

		err = workflow.ExecuteActivity(activitySessionCtx, a.CreateLatestArchiveActivity, &uploadArchiveActivityResult).Get(activitySessionCtx, nil)
		if err != nil {
			return err
		}
	}
	return nil
}
