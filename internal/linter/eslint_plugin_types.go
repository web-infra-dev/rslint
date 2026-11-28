package linter

import (
	"context"
	"crypto/rand"
	"errors"
	"slices"
	"sync"

	"github.com/microsoft/TypeScript/tsc/shim/api/encoder"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/languages/js/estree"
	"github.com/web-infra-dev/rslint/internal/typeservice"
)

// The opaque session and binding identify the actual Program without loading
// its descriptor or compiler syntax. Flags are available to ordinary probes.
type EslintPluginFileTypeServices struct {
	SyntaxOnly             bool                `json:"syntaxOnly,omitempty,omitzero"`
	Binding                typeservice.Binding `json:"binding"`
	Session                string              `json:"session"`
	EmitDecoratorMetadata  bool                `json:"emitDecoratorMetadata,omitempty,omitzero"`
	ExperimentalDecorators bool                `json:"experimentalDecorators,omitempty,omitzero"`
	IsolatedDeclarations   bool                `json:"isolatedDeclarations,omitempty,omitzero"`
}

// EslintPluginTypeRequest opens a Program descriptor, projects one target's
// origins, issues a finite query batch, or reads an exact source. Open and
// projection may be combined on first use. The transports own attachments.
type EslintPluginTypeRequest struct {
	Session      string              `json:"session"`
	Generation   uint32              `json:"generation"`
	Open         bool                `json:"open,omitempty,omitzero"`
	Projection   *string             `json:"projection,omitempty"`
	Queries      []typeservice.Query `json:"queries,omitempty"`
	Source       *string             `json:"source,omitempty"`
	MetadataOnly bool                `json:"metadataOnly,omitempty,omitzero"`
}

type EslintPluginTypeResult struct {
	Status     string                      `json:"status,omitempty"`
	Descriptor *typeservice.Descriptor     `json:"descriptor,omitempty"`
	Projection *EslintPluginTypeProjection `json:"projection,omitempty"`
	Packets    [][]byte                    `json:"packets,omitempty"`
	Source     *EslintPluginTypeSource     `json:"source,omitempty"`
}

// EslintPluginTypeProjectionMeta carries the JSON-serializable origin tables
// without inline Tree bytes so transports can attach Tree as binary data
// without importing internal/languages/js/estree.
type EslintPluginTypeProjectionMeta struct {
	estree.OriginTables
}

type EslintPluginTypeProjection struct {
	EslintPluginTypeProjectionMeta
	Tree []byte `json:"tree"`
}

var ErrPluginTypeSessionClosed = errors.New("type service session is stale or closed")

// PluginTypeFailure preserves lifecycle status through native IPC and LSP,
// independently from their generic error message presentation.
func PluginTypeFailure(err error) *EslintPluginTypeResult {
	switch {
	case errors.Is(err, ErrPluginTypeSessionClosed), errors.Is(err, typeservice.ErrClosed):
		return &EslintPluginTypeResult{Status: "closed"}
	case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
		return &EslintPluginTypeResult{Status: "cancelled"}
	}
	return nil
}

type EslintPluginTypeSource struct {
	FileName string `json:"fileName"`
	Path     string `json:"path,omitempty"`
	Tree     []byte `json:"tree,omitempty"`
}

type pluginTypeSession struct {
	syntaxOnly bool
	mu         sync.RWMutex
	session    *typeservice.Session
	generation uint32
	descriptor func() (typeservice.Descriptor, error)
	targets    map[string]pluginTypeTarget
	ctx        context.Context
	cancel     context.CancelFunc
}

type pluginTypeTarget struct {
	// file is nil for syntax-only targets so syntax-only leases never retain
	// a Program or compiler SourceFile; QueryEslintPluginTypes obtains the
	// lazily cached detached AST via session.Source on demand.
	file       *ast.SourceFile
	sourceType string
}

var pluginTypeSessions = struct {
	sync.RWMutex
	byID map[string]*pluginTypeSession
}{byID: make(map[string]*pluginTypeSession)}

// pluginTypeLease contains opaque registry identities, never detached compiler
// pointers. Closing it cancels pending reads before releasing the producer.
type pluginTypeLease struct {
	once      sync.Once
	ids       []string
	syntaxIDs []string
	release   ReleaseFunc
}

func (l *pluginTypeLease) close() {
	if l == nil {
		return
	}
	l.once.Do(func() {
		release := l.release
		l.release = nil
		if release != nil {
			defer release()
		}
		for _, id := range slices.Concat(l.ids, l.syntaxIDs) {
			pluginTypeSessions.Lock()
			entry := pluginTypeSessions.byID[id]
			delete(pluginTypeSessions.byID, id)
			pluginTypeSessions.Unlock()
			if entry != nil {
				func() {
					entry.cancel()
					entry.mu.Lock()
					defer entry.mu.Unlock()
					if entry.session != nil {
						entry.session.Close()
					}
					entry.targets = nil
				}()
			}
		}
	})
}

//nolint:nilnil // No eligible plugin targets require no semantic lease.
func bindPluginTypeServices(ctx context.Context, plan *LintPlan, inputs []EslintPluginFileInput) (*pluginTypeLease, error) {
	if plan == nil || len(inputs) == 0 {
		return nil, nil
	}
	lease := &pluginTypeLease{}
	transferred := false
	defer func() {
		if !transferred {
			lease.close()
		}
	}()
	cursor := 0
	for _, p := range plan.programs {
		type matchedInput struct {
			file  *ast.SourceFile
			input *EslintPluginFileInput
		}
		var typed, syntax []matchedInput
		for _, file := range p.files {
			if file.rules == nil || len(file.rules.plugins) == 0 {
				continue
			}
			if cursor >= len(inputs) {
				break
			}
			input := &inputs[cursor]
			cursor++
			if !p.program.IsValid() || input.SourceFile == nil {
				continue
			}
			if input.TypeServicesEnabled && p.program.CanProvideTypeChecker(file.file) {
				typed = append(typed, matchedInput{file: file.file, input: input})
			} else {
				syntax = append(syntax, matchedInput{file: file.file, input: input})
			}
		}
		bindGroup := func(items []matchedInput, syntaxOnly bool) error {
			if len(items) == 0 {
				return nil
			}
			admitted := make([]*ast.SourceFile, len(items))
			for i, item := range items {
				admitted[i] = item.file
			}
			var (
				session   *typeservice.Session
				createErr error
			)
			if syntaxOnly {
				session, createErr = typeservice.NewSyntaxOnly(ctx, p.program, admitted)
			} else {
				session, createErr = typeservice.New(ctx, p.program, admitted)
			}
			if createErr != nil {
				return createErr
			}
			options := p.program.Options()
			emitDecoratorMetadata := !syntaxOnly && options != nil && options.EmitDecoratorMetadata.IsTrue()
			experimentalDecorators := !syntaxOnly && options != nil && options.ExperimentalDecorators.IsTrue()
			isolatedDeclarations := !syntaxOnly && options != nil && options.IsolatedDeclarations.IsTrue()
			id := rand.Text()
			entry := &pluginTypeSession{
				syntaxOnly: syntaxOnly,
				session:    session,
				descriptor: sync.OnceValues(session.Descriptor),
				targets:    make(map[string]pluginTypeTarget, len(items)),
			}
			for _, item := range items {
				binding, bindErr := session.BindTarget(item.file)
				if bindErr != nil {
					session.Close()
					return bindErr
				}
				entry.generation = binding.Generation
				var targetFile *ast.SourceFile
				if !syntaxOnly {
					targetFile = item.file
				}
				entry.targets[binding.SourceFileName] = pluginTypeTarget{
					file:       targetFile,
					sourceType: item.input.pluginSourceType(),
				}
				item.input.TypeServices = &EslintPluginFileTypeServices{
					SyntaxOnly:             syntaxOnly,
					Binding:                binding,
					Session:                id,
					EmitDecoratorMetadata:  emitDecoratorMetadata,
					ExperimentalDecorators: experimentalDecorators,
					IsolatedDeclarations:   isolatedDeclarations,
				}
			}
			registerPluginTypeSession(id, entry)
			if syntaxOnly {
				lease.syntaxIDs = append(lease.syntaxIDs, id)
			} else {
				lease.ids = append(lease.ids, id)
			}
			return nil
		}
		if err := bindGroup(typed, false); err != nil {
			return nil, err
		}
		if err := bindGroup(syntax, true); err != nil {
			return nil, err
		}
	}
	if len(lease.ids) == 0 && len(lease.syntaxIDs) == 0 {
		return nil, nil
	}
	transferred = true
	return lease, nil
}

func registerPluginTypeSession(id string, entry *pluginTypeSession) {
	entry.ctx, entry.cancel = context.WithCancel(context.Background())
	pluginTypeSessions.Lock()
	pluginTypeSessions.byID[id] = entry
	pluginTypeSessions.Unlock()
}

// QueryEslintPluginTypes only admits a currently leased session. Teardown
// cancels and joins reads, so no query observes a retired compiler generation.
func QueryEslintPluginTypes(ctx context.Context, req EslintPluginTypeRequest) (res *EslintPluginTypeResult, err error) {
	pluginTypeSessions.RLock()
	entry := pluginTypeSessions.byID[req.Session]
	if entry != nil {
		entry.mu.RLock()
	}
	pluginTypeSessions.RUnlock()
	if entry == nil {
		return nil, ErrPluginTypeSessionClosed
	}
	defer entry.mu.RUnlock()
	if entry.ctx.Err() != nil {
		return nil, ErrPluginTypeSessionClosed
	}
	queryCtx, cancel := context.WithCancelCause(ctx)
	stop := context.AfterFunc(entry.ctx, func() { cancel(ErrPluginTypeSessionClosed) })
	defer func() {
		if abortErr := queryAbortErr(queryCtx, entry); abortErr != nil {
			res = nil
			err = abortErr
		}
		stop()
		cancel(nil)
	}()
	if abortErr := queryAbortErr(queryCtx, entry); abortErr != nil {
		return nil, abortErr
	}
	if req.Generation != entry.generation {
		return nil, errors.New("foreign type service query generation")
	}
	result := &EslintPluginTypeResult{}
	if req.Open || req.Projection != nil {
		if req.Source != nil || len(req.Queries) != 0 {
			return nil, errors.New("type service open and projection cannot include queries or source")
		}
		var (
			target    pluginTypeTarget
			hasTarget bool
		)
		if req.Projection != nil {
			target, hasTarget = entry.targets[*req.Projection]
			if !hasTarget {
				if entry.syntaxOnly {
					return nil, errors.New("foreign syntax projection target")
				}
				return nil, errors.New("source is not eligible for typed projection")
			}
		}
		if req.Open {
			var descriptor typeservice.Descriptor
			if entry.syntaxOnly && hasTarget {
				descriptor, err = entry.session.SyntaxDescriptorForFile(*req.Projection)
			} else {
				descriptor, err = entry.descriptor()
			}
			if err != nil {
				return nil, err
			}
			result.Descriptor = &descriptor
		}
		if req.Projection != nil {
			sf := target.file
			if sf == nil {
				sf, err = entry.session.Source(req.Generation, *req.Projection)
				if err != nil {
					return nil, err
				}
				if sf == nil {
					return nil, errors.New("foreign syntax projection target")
				}
			}
			tree, table, encErr := encoder.EncodeSourceFile(sf)
			if encErr != nil {
				return nil, encErr
			}
			var builder estree.Builder
			origins, projErr := builder.ProjectOrigins(queryCtx, sf, estree.Options{SourceType: target.sourceType}, table)
			if projErr != nil {
				return nil, projErr
			}
			result.Projection = &EslintPluginTypeProjection{
				EslintPluginTypeProjectionMeta: EslintPluginTypeProjectionMeta{OriginTables: origins},
				Tree:                           tree,
			}
		}
		return result, nil
	}
	if req.Source != nil {
		if len(req.Queries) != 0 {
			return nil, errors.New("type service request must select queries or source")
		}
		file, srcErr := entry.session.Source(req.Generation, tspath.NormalizePath(*req.Source))
		if srcErr != nil {
			return nil, srcErr
		}
		if file != nil && req.MetadataOnly {
			result.Source = &EslintPluginTypeSource{FileName: file.FileName(), Path: string(file.Path())}
		} else if file != nil {
			tree, _, encErr := encoder.EncodeSourceFile(file)
			if encErr != nil {
				return nil, encErr
			}
			result.Source = &EslintPluginTypeSource{FileName: file.FileName(), Path: string(file.Path()), Tree: tree}
		}
		return result, nil
	}
	if len(req.Queries) == 0 {
		return nil, errors.New("type service query batch is empty")
	}
	packets, batchErr := entry.session.QueryBatch(queryCtx, req.Queries)
	if batchErr != nil {
		return nil, batchErr
	}
	result.Packets = packets
	return result, nil
}

func queryAbortErr(queryCtx context.Context, entry *pluginTypeSession) error {
	if entry.ctx.Err() != nil {
		return ErrPluginTypeSessionClosed
	}
	if queryCtx.Err() != nil {
		return context.Cause(queryCtx)
	}
	return nil
}
