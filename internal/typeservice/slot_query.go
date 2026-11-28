package typeservice

import (
	"context"
	"errors"
	"fmt"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/core"
)

// SlotQuery describes one binary-RPC semantic query from an independent worker slot.
type SlotQuery struct {
	Slot        int
	WorkerEpoch uint32
	Query       Query
	FileIdx     uint32
	IsCancelled func() bool
}

// SlotResult carries the isolated outcome for one SlotQuery.
type SlotResult struct {
	Packet []byte
	Err    error
}

// QuerySlots executes a batch of independent worker-slot semantic queries under
// a single exclusive checker lease, isolating per-slot validation and handle
// errors while preserving session-fatal panic poisoning and per-slot
// supplemental-record deduplication.
func (s *Session) QuerySlots(ctx context.Context, reqs []SlotQuery) (out []SlotResult) {
	out = make([]SlotResult, len(reqs))
	failAllUnset := func(err error) {
		for i := range out {
			if out[i].Packet == nil && out[i].Err == nil {
				out[i].Err = err
			}
		}
	}
	defer func() {
		if value := recover(); value != nil {
			s.poisoned.Store(true)
			failAllUnset(fmt.Errorf("type service: %v", value))
		}
	}()
	if len(reqs) == 0 {
		return out
	}
	if s.poisoned.Load() {
		failAllUnset(ErrPoisoned)
		return out
	}
	if err := ctx.Err(); err != nil {
		failAllUnset(err)
		return out
	}

	s.lifeMu.RLock()
	defer s.lifeMu.RUnlock()
	defer func() {
		if value := recover(); value != nil {
			s.poisoned.Store(true)
			panic(value)
		}
	}()
	if s.poisoned.Load() {
		failAllUnset(ErrPoisoned)
		return out
	}
	if s.isClosed() {
		failAllUnset(ErrClosed)
		return out
	}
	if s.syntaxOnly {
		failAllUnset(errors.New("semantic queries require a typed project"))
		return out
	}

	p := s.program
	natives := make([]*ast.Node, len(reqs))
	results := make([]Result, len(reqs))
	records := make([][]Record, len(reqs))
	pendingSent := make([][]recordCacheKey, len(reqs))
	topCacheable := make([]bool, len(reqs))
	needChecker := false

	for i, req := range reqs {
		if req.IsCancelled != nil && req.IsCancelled() {
			out[i].Err = context.Canceled
			continue
		}
		q := req.Query
		if q.Generation != s.generation {
			out[i].Err = errors.New("stale or foreign type service handle")
			continue
		}
		if q.Request == nil {
			out[i].Err = errors.New("missing type service request")
			continue
		}
		if _, isProgram := q.Request.(*ProgramRequest); isProgram {
			out[i].Err = errors.New("program queries are not supported on binary RPC")
			continue
		}
		if nr, ok := q.Request.(*NodeRequest); ok && nr.Op == "suggestionDiagnostics" {
			out[i].Err = errors.New("suggestionDiagnostics is not supported on binary RPC")
			continue
		}
		if ref := requestNode(q.Request); ref != nil {
			var f *ast.SourceFile
			if req.FileIdx != BinaryInlineFileSentinel {
				if uint64(req.FileIdx) >= uint64(len(s.sourceFiles)) {
					out[i].Err = errors.New("invalid type service file index")
					continue
				}
				f = s.sourceFiles[req.FileIdx]
			} else {
				f = s.sourceFile(ref.File)
			}
			node, resolveErr := resolveNode(f, *ref)
			if resolveErr != nil {
				out[i].Err = resolveErr
				continue
			}
			natives[i] = node
			if _, isNodeRec := isNodeRecordRequest(q.Request); isNodeRec {
				results[i] = nodeFacts(f, node)
				continue
			}
		}
		if tag, id, isRec := directRecordCacheKey(q.Request); isRec {
			if cached, ok := s.getCachedRecord(tag, id); ok {
				out[i].Packet = cached.packet
				continue
			}
		}
		needChecker = true
	}

	if needChecker {
		if batchErr := s.checkSlots(ctx, reqs, natives, out, results, records, pendingSent, topCacheable); batchErr != nil {
			failAllUnset(batchErr)
			return out
		}
	}

	if s.poisoned.Load() {
		failAllUnset(ErrPoisoned)
		return out
	}

	type sentCommit struct {
		slot  int
		epoch uint32
		keys  []recordCacheKey
	}
	var commits []sentCommit

	for i, req := range reqs {
		if out[i].Err != nil || out[i].Packet != nil {
			continue
		}
		if s.poisoned.Load() {
			out[i].Err = ErrPoisoned
			continue
		}
		if err := ctx.Err(); err != nil {
			out[i].Err = err
			continue
		}
		if req.IsCancelled != nil && req.IsCancelled() {
			out[i].Err = context.Canceled
			continue
		}
		if p != nil {
			materializeReferences(p, results[i])
			for j := range records[i] {
				if records[i][j].packet == nil {
					materializeReferences(p, records[i][j].Fact)
				}
			}
		}
		var encErr error
		for j := range records[i] {
			if records[i][j].packet != nil {
				continue
			}
			encoded, err := EncodePacket(records[i][j].Fact)
			if err != nil {
				encErr = err
				break
			}
			if records[i][j].cacheable {
				s.putCachedRecord(recordTag(records[i][j].Op), records[i][j].ID, encoded, recordSubSymbols(records[i][j].Fact))
			} else {
				markPacketNonCacheable(encoded)
			}
			records[i][j].packet = encoded
		}
		if encErr != nil {
			out[i].Err = encErr
			continue
		}
		packet, err := EncodeReply(results[i], records[i])
		if err != nil {
			out[i].Err = err
			continue
		}
		if tag, id, isRec := directRecordCacheKey(req.Query.Request); isRec && len(records[i]) == 0 {
			if topCacheable[i] {
				s.putCachedRecord(tag, id, packet, recordSubSymbols(results[i]))
			} else {
				markPacketNonCacheable(packet)
			}
		}
		if err := ctx.Err(); err != nil {
			out[i].Err = err
			continue
		}
		if req.IsCancelled != nil && req.IsCancelled() {
			out[i].Err = context.Canceled
			continue
		}
		out[i].Packet = packet
		if len(pendingSent[i]) > 0 && req.Slot >= 0 && req.Slot < maxSentMaskSlots && req.WorkerEpoch != 0 {
			commits = append(commits, sentCommit{slot: req.Slot, epoch: req.WorkerEpoch, keys: pendingSent[i]})
		}
	}

	if len(commits) > 0 {
		s.sentMu.Lock()
		if !s.poisoned.Load() && !s.isClosed() {
			if s.sentMask == nil {
				s.sentMask = make(map[recordCacheKey]uint64)
			}
			for _, commit := range commits {
				if s.slotEpoch[commit.slot] != commit.epoch {
					continue
				}
				bit := uint64(1) << uint(commit.slot)
				for _, key := range commit.keys {
					if _, exists := s.sentMask[key]; !exists && len(s.sentMask) >= maxSentMaskGenEntries {
						s.sentMaskPrev = s.sentMask
						s.sentMask = make(map[recordCacheKey]uint64)
					}
					s.sentMask[key] |= bit
				}
			}
		}
		s.sentMu.Unlock()
	}

	return out
}

func (s *Session) prepareSlotSentBit(slot int, workerEpoch uint32) uint64 {
	if slot < 0 || slot >= maxSentMaskSlots || workerEpoch == 0 {
		return 0
	}
	slotBit := uint64(1) << uint(slot)
	s.sentMu.Lock()
	if s.slotEpoch[slot] != workerEpoch {
		// Respawned or newly attached worker on this slot: clear stale sent bits.
		clearMask := ^slotBit
		for _, m := range []map[recordCacheKey]uint64{s.sentMask, s.sentMaskPrev} {
			for k, bits := range m {
				if nb := bits & clearMask; nb == 0 {
					delete(m, k)
				} else if nb != bits {
					m[k] = nb
				}
			}
		}
		s.slotEpoch[slot] = workerEpoch
	}
	s.sentMu.Unlock()
	return slotBit
}

func (s *Session) isSentToSlot(key recordCacheKey, slotBit uint64) bool {
	if slotBit == 0 {
		return false
	}
	s.sentMu.Lock()
	sent := (s.sentMask != nil && s.sentMask[key]&slotBit != 0) ||
		(s.sentMaskPrev != nil && s.sentMaskPrev[key]&slotBit != 0)
	s.sentMu.Unlock()
	return sent
}

func (s *Session) checkSlots(
	ctx context.Context,
	reqs []SlotQuery,
	natives []*ast.Node,
	out []SlotResult,
	results []Result,
	records [][]Record,
	pendingSent [][]recordCacheKey,
	topCacheable []bool,
) error {
	s.checkMu.Lock()
	defer s.checkMu.Unlock()
	if s.poisoned.Load() {
		return ErrPoisoned
	}
	defer func() {
		if value := recover(); value != nil {
			s.poisoned.Store(true)
			panic(value)
		}
	}()
	if err := ctx.Err(); err != nil {
		return err
	}
	c, done := s.program.TypeCheckerForFileExclusive(core.WithCheckerLifetime(ctx, core.CheckerLifetimeAPI), s.anchor)
	defer done()
	if err := ctx.Err(); err != nil {
		return err
	}
	if c == nil {
		return errors.New("program cannot provide type services")
	}
	if s.checker == nil {
		s.checker = c
		s.types = make(map[uint32]*checker.Type)
		s.symbols = make(map[uint32]*ast.Symbol)
		s.signatures = make(map[uint32]*checker.Signature)
	} else if c != s.checker {
		return errors.New("type service checker identity changed")
	}

	for i, req := range reqs {
		if out[i].Err != nil || out[i].Packet != nil || results[i] != nil {
			continue
		}
		if s.poisoned.Load() {
			return ErrPoisoned
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		if req.IsCancelled != nil && req.IsCancelled() {
			out[i].Err = context.Canceled
			continue
		}

		slotBit := s.prepareSlotSentBit(req.Slot, req.WorkerEpoch)

		res, err := s.execute(ctx, c, req.Query.Request, natives[i])
		if err != nil {
			out[i].Err = err
			continue
		}
		results[i] = res
		if tag, id, isRec := directRecordCacheKey(req.Query.Request); isRec {
			topCacheable[i] = s.isCacheableRecord(tag, id)
		}
		if !req.Query.SkipRecords {
			recs, pending, recErr := s.directRecordsForSlot(ctx, c, req.Query.Request.operation(), res, slotBit)
			if recErr != nil {
				out[i].Err = recErr
				results[i] = nil
				continue
			}
			records[i] = recs
			pendingSent[i] = pending
		}
	}
	return nil
}
