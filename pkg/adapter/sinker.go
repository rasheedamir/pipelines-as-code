package adapter

import (
	"bytes"
	"context"
	"net/http"

	"github.com/openshift-pipelines/pipelines-as-code/pkg/kubeinteraction"
	"github.com/openshift-pipelines/pipelines-as-code/pkg/params"
	"github.com/openshift-pipelines/pipelines-as-code/pkg/params/info"
	"github.com/openshift-pipelines/pipelines-as-code/pkg/pipelineascode"
	"github.com/openshift-pipelines/pipelines-as-code/pkg/provider"
	"go.uber.org/zap"
)

type sinker struct {
	run    *params.Run
	vcx    provider.Interface
	kint   *kubeinteraction.Interaction
	event  *info.Event
	logger *zap.SugaredLogger
}

func (s *sinker) processEvent(ctx context.Context, request *http.Request, payload []byte) error {
	var err error
	s.event, err = s.vcx.ParsePayload(ctx, s.run, request, string(payload))
	if err != nil {
		s.logger.Errorf("failed to parse event: %v", err)
		return err
	}

	// set logger with sha and event type
	s.logger = s.logger.With("event-sha", s.event.SHA, "event-type", s.event.EventType)
	s.vcx.SetLogger(s.logger)

	s.event.Request = &info.Request{
		Header:  request.Header,
		Payload: bytes.TrimSpace(payload),
	}

	p := pipelineascode.NewPacs(s.event, s.vcx, s.run, s.kint, s.logger)
	return p.Run(ctx)
}
