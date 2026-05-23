package monitoring

import "context"

type ArmsClient interface {
	GetRUMApps(ctx context.Context, pageToken string, pageSize int) (*AdapterResult, error)
	GetRUMExceptionStack(ctx context.Context, pid string, line, column int, sourcemapType, exceptionBinaryImages string) (*AdapterResult, error)
}

type AtomicService struct {
	sls  *SLSClient
	arms ArmsClient
}

func NewAtomicService(sls *SLSClient, armsClient ArmsClient) *AtomicService {
	return &AtomicService{sls: sls, arms: armsClient}
}

func (s *AtomicService) ArmsListRUMApps(ctx context.Context, pageToken string, pageSize int) map[string]any {
	return Wrap(func() (*AdapterResult, error) {
		return s.arms.GetRUMApps(ctx, pageToken, pageSize)
	})
}

func (s *AtomicService) ArmsResolveExceptionStack(ctx context.Context, pid string, line, column int, sourcemapType, exceptionBinaryImages string) map[string]any {
	return Wrap(func() (*AdapterResult, error) {
		return s.arms.GetRUMExceptionStack(ctx, pid, line, column, sourcemapType, exceptionBinaryImages)
	})
}

func (s *AtomicService) SLSSearchLogs(ctx context.Context, project, logstore string, fromMS, toMS int64, pageToken string, pageSize int, reverse bool, query string) map[string]any {
	return Wrap(func() (*AdapterResult, error) {
		return s.sls.SearchLogs(ctx, project, logstore, fromMS, toMS, pageToken, pageSize, reverse, query)
	})
}

func (s *AtomicService) SLSGetLogContext(ctx context.Context, project, logstore, packID, packMeta string, backLines, forwardLines int) map[string]any {
	return Wrap(func() (*AdapterResult, error) {
		return s.sls.GetLogContext(ctx, project, logstore, packID, packMeta, backLines, forwardLines)
	})
}
