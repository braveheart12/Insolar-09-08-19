package executor

import (
	"context"
	"fmt"

	"github.com/pkg/errors"

	"github.com/insolar/insolar/insolar"
	"github.com/insolar/insolar/insolar/bus"
	"github.com/insolar/insolar/insolar/jet"
	"github.com/insolar/insolar/insolar/payload"
	"github.com/insolar/insolar/insolar/record"
	"github.com/insolar/insolar/instrumentation/inslogger"
)

//go:generate minimock -i github.com/insolar/insolar/ledger/light/executor.RequestChecker -o ./ -s _mock.go -g

type RequestChecker interface {
	CheckRequest(ctx context.Context, requestID insolar.ID, request record.Request) error
}

type RequestCheckerDefault struct {
	filaments   FilamentCalculator
	coordinator jet.Coordinator
	fetcher     JetFetcher
	sender      bus.Sender
}

func NewRequestChecker(
	fc FilamentCalculator,
	c jet.Coordinator,
	jf JetFetcher,
	sender bus.Sender,
) *RequestCheckerDefault {
	return &RequestCheckerDefault{
		filaments:   fc,
		coordinator: c,
		fetcher:     jf,
		sender:      sender,
	}
}

func (c *RequestCheckerDefault) CheckRequest(ctx context.Context, requestID insolar.ID, request record.Request) error {
	if request.ReasonRef().IsEmpty() {
		return errors.New("reason id is empty")
	}
	reasonRef := request.ReasonRef()
	reasonID := *reasonRef.Record()

	if reasonID.Pulse() > requestID.Pulse() {
		return errors.New("request is older than its reason")
	}

	switch r := request.(type) {
	case *record.IncomingRequest:
		// Cannot be detached.
		if r.IsDetached() {
			return errors.Errorf("incoming request cannot be detached (got mode %v)", r.ReturnMode)
		}

		// Reason should exist.
		// FIXME: replace with remote request check.
		if !request.IsAPIRequest() {
			reasonObject := r.ReasonAffinityRef()
			if reasonObject.IsEmpty() {
				return errors.New("reason affinity is not set on incoming request")
			}

			err := c.checkReasonExists(ctx, *reasonObject.Record(), reasonID)
			if err != nil {
				return errors.Wrap(err, "reason not found")
			}
		}

	case *record.OutgoingRequest:
		if request.IsCreationRequest() {
			return errors.New("outgoing cannot be creating request")
		}

		// FIXME: replace with "FindRequest" calculator method.
		requests, err := c.filaments.OpenedRequests(
			ctx,
			requestID.Pulse(),
			*request.AffinityRef().Record(),
			false,
		)
		if err != nil {
			return errors.Wrap(err, "failed fetch pending requests")
		}

		reasonInPendings := contains(requests, reasonID)
		if !reasonInPendings {
			return errors.New("request reason not found in opened requests")
		}
	}

	return nil
}

func (c *RequestCheckerDefault) checkReasonExists(
	ctx context.Context, objectID insolar.ID, reasonID insolar.ID,
) error {
	isBeyond, err := c.coordinator.IsBeyondLimit(ctx, reasonID.Pulse())
	if err != nil {
		return errors.Wrap(err, "failed to calculate limit")
	}
	var node *insolar.Reference
	if isBeyond {
		node, err = c.coordinator.Heavy(ctx)
		if err != nil {
			return errors.Wrap(err, "failed to calculate node")
		}
	} else {
		jetID, err := c.fetcher.Fetch(ctx, objectID, reasonID.Pulse())
		if err != nil {
			return errors.Wrap(err, "failed to fetch jet")
		}
		node, err = c.coordinator.NodeForJet(ctx, *jetID, reasonID.Pulse())
		if err != nil {
			return errors.Wrap(err, "failed to calculate node")
		}
	}
	inslogger.FromContext(ctx).Debug("check reason. request: ", reasonID.DebugString())
	msg, err := payload.NewMessage(&payload.GetRequest{
		ObjectID:  objectID,
		RequestID: reasonID,
	})
	if err != nil {
		return errors.Wrap(err, "failed to check an object existence")
	}

	reps, done := c.sender.SendTarget(ctx, msg, *node)
	defer done()
	res, ok := <-reps
	if !ok {
		return errors.New("no reply for reason check")
	}

	pl, err := payload.UnmarshalFromMeta(res.Payload)
	if err != nil {
		return errors.Wrap(err, "failed to unmarshal reply")
	}

	switch concrete := pl.(type) {
	case *payload.Request:
		return nil
	case *payload.Error:
		return errors.New(concrete.Text)
	default:
		return fmt.Errorf("unexpected reply %T", pl)
	}
}

func contains(pendings []record.CompositeFilamentRecord, requestID insolar.ID) bool {
	for _, p := range pendings {
		if p.RecordID == requestID {
			return true
		}
	}
	return false
}
