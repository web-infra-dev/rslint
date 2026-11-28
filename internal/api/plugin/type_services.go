package plugin

import (
	"context"
	"fmt"

	"github.com/web-infra-dev/rslint/internal/ipc"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/typeservice"
)

type typeServiceProjectionAttachment struct {
	linter.EslintPluginTypeProjectionMeta
	TreeAttachment int `json:"treeAttachment"`
}

type typeServiceSourceAttachment struct {
	FileName       string `json:"fileName"`
	Path           string `json:"path,omitempty"`
	TreeAttachment *int   `json:"treeAttachment,omitempty"`
}

type typeServiceWireResponse struct {
	Descriptor        *typeservice.Descriptor          `json:"descriptor,omitempty"`
	Projection        *typeServiceProjectionAttachment `json:"projection,omitempty"`
	Packets           [][]byte                         `json:"packets,omitempty"`
	PacketAttachments []int                            `json:"packetAttachments,omitempty"`
	Source            *typeServiceSourceAttachment     `json:"source,omitempty"`
}

// Small binary packets stay inline to avoid claiming 16 MiB shared-memory slots
// and emitting separate transportRelease control frames on every checker query.
const minSharedTypeServiceBytes = 4 << 10

// HandleTypeServices answers nested semantic reads through the same native
// connection that dispatched plugin work. Binary facts and syntax use generic
// response attachments; the transport owns publication and acknowledgement.
func HandleTypeServices(ctx context.Context, msg *ipc.Message) (any, error) {
	var request linter.EslintPluginTypeRequest
	if err := msg.Decode(&request); err != nil {
		return nil, fmt.Errorf("decode typeService request: %w", err)
	}
	result, err := linter.QueryEslintPluginTypes(ctx, request)
	if err != nil {
		if status := linter.PluginTypeFailure(err); status != nil {
			return status, nil
		}
		return nil, err
	}
	return encodeTypeServiceWireResponse(result), nil
}

func encodeTypeServiceWireResponse(result *linter.EslintPluginTypeResult) any {
	var attachments []ipc.Attachment
	wire := typeServiceWireResponse{Descriptor: result.Descriptor}
	if projection := result.Projection; projection != nil {
		wire.Projection = &typeServiceProjectionAttachment{
			EslintPluginTypeProjectionMeta: projection.EslintPluginTypeProjectionMeta,
			TreeAttachment:                 len(attachments),
		}
		attachments = append(attachments, ipc.Bytes(projection.Tree))
	}
	if len(result.Packets) > 0 {
		totalPacketBytes := 0
		for _, packet := range result.Packets {
			totalPacketBytes += len(packet)
		}
		if totalPacketBytes >= minSharedTypeServiceBytes {
			for _, packet := range result.Packets {
				wire.PacketAttachments = append(wire.PacketAttachments, len(attachments))
				attachments = append(attachments, ipc.Bytes(packet))
			}
		} else {
			wire.Packets = result.Packets
		}
	}
	if result.Source != nil {
		wire.Source = &typeServiceSourceAttachment{FileName: result.Source.FileName, Path: result.Source.Path}
		if len(result.Source.Tree) > 0 {
			index := len(attachments)
			wire.Source.TreeAttachment = &index
			attachments = append(attachments, ipc.Bytes(result.Source.Tree))
		}
	}
	if len(attachments) == 0 {
		return wire
	}
	return ipc.ResponseWithAttachments(wire, attachments...)
}
